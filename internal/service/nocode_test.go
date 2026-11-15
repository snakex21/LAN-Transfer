package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"lantransfer/internal/i18n"
	"lantransfer/internal/transfer"
)

// With "Require pairing code" off, a sender connects with an empty code and
// sends; the connection is still TLS with the receiver's certificate.
func TestCodeOffAllowsTransferWithoutCode(t *testing.T) {
	p := startPeer(t, nil)
	p.state.Pairing.SetCodeRequired(false)

	info, err := transfer.FetchInfo(context.Background(), p.address)
	if err != nil || !info.NoCode || info.Fingerprint != p.state.Identity.Fingerprint {
		t.Fatalf("/api/info: %+v %v", info, err)
	}
	session, err := transfer.Pair(context.Background(), p.address, "", senderInfo())
	if err != nil {
		t.Fatalf("pair without code: %v", err)
	}
	defer session.Close()
	if session.Fingerprint != p.state.Identity.Fingerprint {
		t.Fatal("certificate fingerprint not pinned")
	}
	src := filepath.Join(t.TempDir(), "bez-kodu.bin")
	data := writeRandom(t, src, 300_000)
	if _, _, err := send(t, session, src, ""); err != nil {
		t.Fatalf("send without code: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(p.dir, "bez-kodu.bin"))
	if err != nil || sha(got) != sha(data) {
		t.Fatalf("received file differs: %v", err)
	}
}

// With the code on (the default) an empty code is refused with
// "code_required", nothing can be uploaded, and switching the code back on
// revokes sessions opened without it.
func TestCodeOnRequiresCode(t *testing.T) {
	p := startPeer(t, nil)
	info, err := transfer.FetchInfo(context.Background(), p.address)
	if err != nil || info.NoCode {
		t.Fatalf("/api/info must not announce no_code: %+v %v", info, err)
	}
	_, err = transfer.Pair(context.Background(), p.address, "", senderInfo())
	if !transfer.IsCodeRequired(err) || i18n.MsgOf(err).Key != "err.codeRequired" {
		t.Fatalf("empty code: %v (%+v)", err, i18n.MsgOf(err))
	}
	q := url.Values{"transfer": {"0123456789abcdef"}, "path": {"x.txt"}, "size": {"2"}}
	if status, _ := rawUpload(t, p, "", q, []byte("hi"), sha([]byte("hi"))); status != http.StatusUnauthorized {
		t.Fatalf("upload without token: %d", status)
	}

	// Off: connect without a code; on again: that session stops working.
	p.state.Pairing.SetCodeRequired(false)
	session, err := transfer.Pair(context.Background(), p.address, "", senderInfo())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	p.state.Pairing.SetCodeRequired(true)
	src := filepath.Join(t.TempDir(), "x.txt")
	writeRandom(t, src, 10)
	if _, _, err := send(t, session, src, ""); !errors.Is(err, transfer.ErrNotPaired) {
		t.Fatalf("session opened without a code survived switching the code on: %v", err)
	}
	// The code still works.
	if _, err := transfer.Pair(context.Background(), p.address, testCode, senderInfo()); err != nil {
		t.Fatalf("pair with code: %v", err)
	}
}
