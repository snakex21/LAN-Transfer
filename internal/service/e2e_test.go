package service_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lantransfer/internal/app"
	"lantransfer/internal/i18n"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
)

const testCode = "246810"

type peer struct {
	state    *app.State
	receiver *transfer.Receiver
	address  string
	dir      string
}

// startPeer runs a receiver on 127.0.0.1 with an in-process HTTPS server.
func startPeer(t *testing.T, configure func(*transfer.Receiver)) *peer {
	t.Helper()
	dir := t.TempDir()
	identity, err := security.NewIdentity("test")
	if err != nil {
		t.Fatal(err)
	}
	state := &app.State{ID: "receiver-id", Name: "Odbiorca", ReceiveDir: dir, Identity: identity,
		Pairing: security.NewPairing(testCode), History: transfer.LoadHistory("")}
	state.Incoming = transfer.NewIncomingTracker(state.History)
	receiver := &transfer.Receiver{Directory: dir, Incoming: state.Incoming}
	if configure != nil {
		configure(receiver)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := service.NewServerWith(state, receiver)
	go server.ServeTLS(listener, "", "")
	t.Cleanup(func() { server.Close() })
	return &peer{state: state, receiver: receiver, address: listener.Addr().String(), dir: dir}
}

func senderInfo() protocol.Info { return protocol.Info{ID: "sender-id", Name: "Nadawca"} }

func pair(t *testing.T, p *peer) *transfer.Session {
	t.Helper()
	s, err := transfer.Pair(context.Background(), p.address, testCode, senderInfo())
	if err != nil {
		t.Fatalf("pair: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func send(t *testing.T, s *transfer.Session, path, dest string) (*transfer.Plan, *transfer.Progress, error) {
	t.Helper()
	plan, err := transfer.PlanPath(path, s.Address, dest)
	if err != nil {
		t.Fatal(err)
	}
	progress := &transfer.Progress{}
	return plan, progress, s.Send(context.Background(), plan, dest, 4, progress)
}

func writeRandom(t *testing.T, path string, size int) []byte {
	t.Helper()
	data := make([]byte, size)
	rand.Read(data)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return data
}

// treeDigest maps relative paths to SHA-256 (directories map to "dir").
func treeDigest(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			out[rel] = "dir"
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// rawClient talks to the receiver without the sender package, like an attacker.
func rawClient() *http.Client {
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, Timeout: 10 * time.Second}
}

func rawUpload(t *testing.T, p *peer, token string, query url.Values, body []byte, digest string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "https://"+p.address+protocol.PathUpload+"?"+query.Encode(), io.NopCloser(bytes.NewReader(body)))
	req.ContentLength = -1
	req.Trailer = http.Header{protocol.TrailerSHA256: {digest}}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := rawClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(data)
}

func sha(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestUnpairedPeerCannotUploadOrBrowse(t *testing.T) {
	p := startPeer(t, nil)
	q := url.Values{"transfer": {"0123456789abcdef"}, "path": {"x.txt"}, "size": {"2"}}
	if status, _ := rawUpload(t, p, "", q, []byte("hi"), sha([]byte("hi"))); status != http.StatusUnauthorized {
		t.Fatalf("upload without token: %d", status)
	}
	if status, _ := rawUpload(t, p, "forged-token", q, []byte("hi"), sha([]byte("hi"))); status != http.StatusUnauthorized {
		t.Fatalf("upload with forged token: %d", status)
	}
	for _, path := range []string{protocol.PathBrowse + "?path=", protocol.PathOffset + "?" + q.Encode()} {
		resp, err := rawClient().Get("https://" + p.address + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("GET %s without token: %d", path, resp.StatusCode)
		}
	}
	entries, _ := os.ReadDir(p.dir)
	if len(entries) != 0 {
		t.Fatalf("receiver directory not empty: %v", entries)
	}
	// Plain HTTP (protocol v1 clients) must not work at all.
	if resp, err := http.Get("http://" + p.address + protocol.PathInfo); err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal("unencrypted HTTP accepted")
		}
	}
}

func TestWrongCodeIsRejectedAndLocksOut(t *testing.T) {
	p := startPeer(t, nil)
	ctx := context.Background()
	for i := 0; i < security.MaxFailuresPerHost; i++ {
		if _, err := transfer.Pair(ctx, p.address, "000000", senderInfo()); err == nil {
			t.Fatal("wrong code accepted")
		}
	}
	_, err := transfer.Pair(ctx, p.address, testCode, senderInfo())
	var remote *transfer.RemoteError
	if !errors.As(err, &remote) || remote.Status != http.StatusTooManyRequests {
		t.Fatalf("expected lockout, got %v", err)
	}
}

func TestSendFileFolderUnicodeLongPathsAndManyFiles(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	src := t.TempDir()

	small := filepath.Join(src, "mały plik ąę.txt")
	writeRandom(t, small, 1234)
	empty := filepath.Join(src, "pusty.bin")
	writeRandom(t, empty, 0)
	tree := filepath.Join(src, "Folder ŻÓŁW 日本語")
	for i := 0; i < 300; i++ {
		writeRandom(t, filepath.Join(tree, fmt.Sprintf("kat-%d", i%7), fmt.Sprintf("plik %03d ćma.dat", i)), i*37)
	}
	long := tree
	for i := 0; i < 6; i++ {
		long = filepath.Join(long, strings.Repeat("długi_segment_", 4)+fmt.Sprint(i))
	}
	writeRandom(t, filepath.Join(long, "głęboki plik.txt"), 4096)
	os.MkdirAll(filepath.Join(tree, "pusty folder"), 0o755)
	if len(long) < 300 {
		t.Fatalf("long path only %d chars", len(long))
	}

	for _, path := range []string{small, empty, tree} {
		if _, _, err := send(t, s, path, ""); err != nil {
			t.Fatalf("send %s: %v", path, err)
		}
	}
	want := treeDigest(t, src)
	got := treeDigest(t, p.dir)
	for rel, digest := range want {
		if got[rel] != digest {
			t.Errorf("%s: got %q want %q", rel, got[rel], digest)
		}
	}
	if _, err := os.Stat(filepath.Join(p.dir, ".lantransfer-partial")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("partial folder left behind: %v", err)
	}
	if n := len(p.state.Incoming.Snapshot()); n != 0 {
		t.Errorf("%d incoming jobs still active", n)
	}
	if h := p.state.History.Entries(); len(h) != 3 || h[0].Direction != "received" || h[0].Status != transfer.StatusDone {
		t.Errorf("receiver history: %+v", h)
	}
}

func TestNameCollisionsNeverOverwrite(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	src := t.TempDir()
	os.WriteFile(filepath.Join(p.dir, "raport.txt"), []byte("oryginał"), 0o644)
	writeRandom(t, filepath.Join(src, "raport.txt"), 100)
	writeRandom(t, filepath.Join(src, "Projekt", "a.txt"), 10)

	if _, _, err := send(t, s, filepath.Join(src, "raport.txt"), ""); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		os.Chtimes(filepath.Join(src, "Projekt", "a.txt"), time.Now(), time.Now().Add(time.Duration(i)*time.Second))
		if _, _, err := send(t, s, filepath.Join(src, "Projekt"), ""); err != nil {
			t.Fatal(err)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(p.dir, "raport.txt")); string(data) != "oryginał" {
		t.Fatal("existing file was overwritten")
	}
	for _, name := range []string{"raport (1).txt", filepath.Join("Projekt", "a.txt"), filepath.Join("Projekt (1)", "a.txt")} {
		if _, err := os.Stat(filepath.Join(p.dir, name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
}

func TestHostileNamesAreRejected(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	outside := filepath.Dir(p.dir)
	body := []byte("pwned")
	for _, name := range []string{
		"../uciekinier.txt", "a/../../uciekinier.txt", `..\uciekinier.txt`, filepath.Join(outside, "abs.txt"),
		"/abs.txt", "C:/abs.txt", "plik.txt:ukryty", "CON", "nul.txt", "COM1.log", "kropka.", `a\b`,
	} {
		q := url.Values{"transfer": {"0123456789abcdef"}, "path": {name}, "size": {"5"}}
		status, msg := rawUpload(t, p, s.Token, q, body, sha(body))
		if status != http.StatusBadRequest {
			t.Errorf("%q: status %d %s", name, status, msg)
		}
	}
	if _, err := os.Stat(filepath.Join(outside, "uciekinier.txt")); err == nil {
		t.Fatal("file escaped the receive folder")
	}
	entries, _ := os.ReadDir(p.dir)
	for _, e := range entries {
		if e.Name() != ".lantransfer-partial" {
			t.Errorf("unexpected file %s", e.Name())
		}
	}
}

func TestChecksumMismatchDiscardsFile(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	body := []byte("zawartość")
	q := url.Values{"transfer": {"0123456789abcdef"}, "path": {"zly.txt"}, "size": {fmt.Sprint(len(body))}}
	status, msg := rawUpload(t, p, s.Token, q, body, sha([]byte("coś innego")))
	if status != http.StatusUnprocessableEntity || !strings.Contains(msg, protocol.CodeChecksum) {
		t.Fatalf("status %d %s", status, msg)
	}
	if _, err := os.Stat(filepath.Join(p.dir, "zly.txt")); err == nil {
		t.Fatal("corrupted file was kept")
	}
	// Declared size smaller than body is refused too.
	q.Set("size", "3")
	if status, _ := rawUpload(t, p, s.Token, q, body, sha(body)); status != http.StatusBadRequest {
		t.Fatalf("oversized body: %d", status)
	}
}

func TestCancelAndResume(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	src := filepath.Join(t.TempDir(), "duzy.bin")
	data := writeRandom(t, src, 96<<20)

	plan, err := transfer.PlanPath(src, s.Address, "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	progress := &transfer.Progress{}
	go func() {
		for progress.DoneBytes.Load() < 20<<20 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()
	if err := s.Send(ctx, plan, "", 1, progress); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.dir, "duzy.bin")); err == nil {
		t.Fatal("cancelled transfer left a file under the final name")
	}

	// The receiver may still be flushing the aborted request; wait for the part.
	var resumed int64
	for i := 0; i < 100 && resumed == 0; i++ {
		time.Sleep(20 * time.Millisecond)
		progress = &transfer.Progress{}
		plan2, _ := transfer.PlanPath(src, s.Address, "")
		if plan2.ID != plan.ID {
			t.Fatal("job id is not stable")
		}
		err = s.Send(context.Background(), plan2, "", 1, progress)
		resumed = progress.Resumed.Load()
		if err != nil {
			t.Fatalf("resume: %v", err)
		}
	}
	if resumed < 1<<20 {
		t.Fatalf("nothing was resumed (%d bytes)", resumed)
	}
	got, err := os.ReadFile(filepath.Join(p.dir, "duzy.bin"))
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("resumed file differs (err %v)", err)
	}
	t.Logf("resumed %d of %d bytes", resumed, len(data))
}

func TestDiskFullIsReportedClearly(t *testing.T) {
	p := startPeer(t, func(r *transfer.Receiver) {
		r.FreeSpace = func(string) (uint64, error) { return 10 << 20, nil }
	})
	s := pair(t, p)
	src := filepath.Join(t.TempDir(), "za_duzy.bin")
	writeRandom(t, src, 1<<20)
	_, _, err := send(t, s, src, "")
	var remote *transfer.RemoteError
	if !errors.As(err, &remote) || remote.Code != protocol.CodeNoSpace || remote.Key != "err.notEnoughSpace" {
		t.Fatalf("expected no_space error, got %v", err)
	}
	// The receiver sends key + parameters, so the sender shows the message in
	// its own language with its own number format, whatever the receiver uses.
	msg := i18n.MsgOf(err)
	if msg.Key != "err.withPath" {
		t.Fatalf("expected the file name around the remote error, got %+v", msg)
	}
	for lang, want := range map[string]string{
		"en": "za_duzy.bin: not enough space on the receiver's disk (needed 1 MB, free 10 MB)",
		"pl": "za_duzy.bin: za mało miejsca na dysku odbiorcy (potrzeba 1 MB, wolne 10 MB)",
	} {
		if got := i18n.Render(lang, msg); got != want {
			t.Errorf("%s: %q, want %q", lang, got, want)
		}
	}
}

func TestPairingErrorsAreTranslatable(t *testing.T) {
	p := startPeer(t, nil)
	_, err := transfer.Pair(context.Background(), p.address, "000000", senderInfo())
	if msg := i18n.MsgOf(err); msg.Key != "err.wrongCode" {
		t.Fatalf("wrong code: %+v (%v)", msg, err)
	}
	if got := i18n.Render("pl", i18n.MsgOf(err)); got != "nieprawidłowy kod parowania" {
		t.Fatalf("pl: %q", got)
	}
}

func TestMaxFileSize(t *testing.T) {
	p := startPeer(t, func(r *transfer.Receiver) { r.MaxFileSize = 1000 })
	s := pair(t, p)
	src := filepath.Join(t.TempDir(), "x.bin")
	writeRandom(t, src, 1001)
	_, _, err := send(t, s, src, "")
	var remote *transfer.RemoteError
	if !errors.As(err, &remote) || remote.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %v", err)
	}
}

func TestDestinationFolderForPairedPeer(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	dest := t.TempDir()
	src := filepath.Join(t.TempDir(), "a.txt")
	writeRandom(t, src, 10)
	if _, _, err := send(t, s, src, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "a.txt")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := send(t, s, src, "relative\\dir"); err == nil {
		t.Fatal("relative destination accepted")
	}
	listing, err := s.Browse(context.Background(), dest)
	if err != nil || listing.Path != dest {
		t.Fatalf("browse: %+v %v", listing, err)
	}
}

func TestPinnedFingerprintDetectsDifferentServer(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	other := startPeer(t, nil)
	impostor := *s
	impostor.Address = other.address
	// Same session (pinned to p's certificate) pointed at another server.
	_, err := impostor.Browse(context.Background(), "")
	if !errors.Is(err, security.ErrFingerprintMismatch) {
		t.Fatalf("expected fingerprint mismatch, got %v", err)
	}
}

func TestManagerQueueCancelRetry(t *testing.T) {
	p := startPeer(t, nil)
	m := transfer.NewManager(senderInfo(), transfer.LoadHistory(""))
	if _, err := m.Enqueue([]string{"x"}, p.address, ""); !errors.Is(err, transfer.ErrNotPaired) {
		t.Fatalf("enqueue without pairing: %v", err)
	}
	if _, err := m.Pair(context.Background(), p.address, testCode); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "kolejka.bin")
	data := writeRandom(t, src, 64<<20)
	ids, err := m.Enqueue([]string{src}, p.address, "")
	if err != nil || len(ids) != 1 {
		t.Fatal(ids, err)
	}
	waitFor(t, func() bool { v := m.Snapshot()[0]; return v.DoneBytes > 4<<20 || v.Status == transfer.StatusDone })
	m.Cancel(ids[0])
	waitFor(t, func() bool {
		s := m.Snapshot()[0].Status
		return s == transfer.StatusCancelled || s == transfer.StatusDone
	})
	if err := m.Retry(ids[0]); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return m.Snapshot()[0].Status == transfer.StatusDone })
	got, _ := os.ReadFile(filepath.Join(p.dir, "kolejka.bin"))
	if !bytes.Equal(got, data) {
		t.Fatal("file differs after cancel + retry")
	}
	if h := m.History.Entries(); len(h) < 1 || h[0].Status != transfer.StatusDone {
		t.Fatalf("history %+v", h)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestResendIsIdempotentButRestoresDeletedFiles(t *testing.T) {
	p := startPeer(t, nil)
	s := pair(t, p)
	src := t.TempDir()
	tree := filepath.Join(src, "Zdjęcia")
	for i := 0; i < 50; i++ {
		writeRandom(t, filepath.Join(tree, fmt.Sprintf("img%02d.jpg", i)), 1000+i)
	}
	for i := 0; i < 2; i++ {
		if _, _, err := send(t, s, tree, ""); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(p.dir, "Zdjęcia"))
	if len(entries) != 50 {
		t.Fatalf("unchanged re-send produced %d files, want 50 (no duplicates)", len(entries))
	}
	os.Remove(filepath.Join(p.dir, "Zdjęcia", "img07.jpg"))
	if _, _, err := send(t, s, tree, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p.dir, "Zdjęcia", "img07.jpg")); err != nil {
		t.Fatalf("deleted file was not sent again: %v", err)
	}
}
