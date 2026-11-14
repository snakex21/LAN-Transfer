package transfer

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"syscall"
	"testing"
)

func randomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func TestCopyHashingMatchesSequentialHash(t *testing.T) {
	for _, size := range []int{0, 1, chunkSize - 1, chunkSize, 5*chunkSize + 123} {
		data := randomBytes(size)
		var out bytes.Buffer
		h := sha256.New()
		var reported int64
		n, err := copyHashing(&out, h, bytes.NewReader(data), func(k int64) { reported += k })
		want := sha256.Sum256(data)
		if err != nil || n != int64(size) || reported != int64(size) || !bytes.Equal(out.Bytes(), data) || !bytes.Equal(h.Sum(nil), want[:]) {
			t.Fatalf("size %d: n=%d reported=%d err=%v", size, n, reported, err)
		}
	}
}

type failingWriter struct{ after int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.after <= 0 {
		return 0, syscall.ENOSPC
	}
	w.after--
	return len(p), nil
}

func TestCopyHashingStopsOnWriteError(t *testing.T) {
	_, err := copyHashing(&failingWriter{after: 2}, sha256.New(), bytes.NewReader(randomBytes(20*chunkSize)), nil)
	if err == nil || !isNoSpace(err) {
		t.Fatalf("expected disk-full error, got %v", err)
	}
}

func TestPrefetchReaderStreamsAndHashes(t *testing.T) {
	data := randomBytes(7*chunkSize + 77)
	h := sha256.New()
	var read int64
	r := newPrefetchReader(bytes.NewReader(data), h, func(n int64) { read += n })
	got, err := io.ReadAll(r)
	want := sha256.Sum256(data)
	if err != nil || !bytes.Equal(got, data) || !bytes.Equal(h.Sum(nil), want[:]) || read != int64(len(data)) {
		t.Fatalf("err=%v len=%d read=%d", err, len(got), read)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("dysk odłączony") }

func TestPrefetchReaderReportsReadErrorsAndCloses(t *testing.T) {
	r := newPrefetchReader(errReader{}, sha256.New(), nil)
	if _, err := io.ReadAll(r); err == nil || err.Error() != "dysk odłączony" {
		t.Fatalf("got %v", err)
	}
	r2 := newPrefetchReader(bytes.NewReader(randomBytes(30*chunkSize)), sha256.New(), nil)
	buf := make([]byte, 10)
	r2.Read(buf)
	r2.Close() // must not deadlock with the read-ahead goroutine blocked
}
