package security

import (
	"errors"
	"testing"
	"time"
)

func TestPairingIssuesTokenOnlyForCorrectCode(t *testing.T) {
	p := NewPairing("")
	code := p.Code()
	if len(code) != 6 {
		t.Fatalf("code %q is not 6 digits", code)
	}
	if _, err := p.Pair("10.0.0.2", "000000x", Peer{Name: "x"}); !errors.Is(err, ErrWrongCode) {
		t.Fatalf("wrong code accepted: %v", err)
	}
	token, err := p.Pair("10.0.0.2", code[:3]+"-"+code[3:], Peer{Name: "Laptop"})
	if err != nil || len(token) != 64 {
		t.Fatalf("Pair = %q, %v", token, err)
	}
	peer, ok := p.Authorize(token)
	if !ok || peer.Name != "Laptop" || peer.Address != "10.0.0.2" {
		t.Fatalf("Authorize = %+v, %v", peer, ok)
	}
	if _, ok := p.Authorize("deadbeef"); ok {
		t.Fatal("unknown token authorized")
	}
	if _, ok := p.Authorize(""); ok {
		t.Fatal("empty token authorized")
	}
}

func TestPairingLocksOutAfterRepeatedFailures(t *testing.T) {
	p := NewPairing("123456")
	now := time.Now()
	p.now = func() time.Time { return now }
	for i := 0; i < MaxFailuresPerHost; i++ {
		p.Pair("h", "999999", Peer{})
	}
	if _, err := p.Pair("h", "123456", Peer{}); !errors.Is(err, ErrLocked) {
		t.Fatalf("expected lockout, got %v", err)
	}
	if _, err := p.Pair("other", "123456", Peer{}); err != nil {
		t.Fatalf("other host should not be locked: %v", err)
	}
	now = now.Add(LockoutDuration + time.Second)
	if _, err := p.Pair("h", "123456", Peer{}); err != nil {
		t.Fatalf("lockout should expire: %v", err)
	}
}

func TestPairingRotatesCodeAfterManyFailures(t *testing.T) {
	p := NewPairing("")
	first := p.Code()
	token, _ := p.Pair("a", first, Peer{})
	for i := 0; i < MaxFailuresTotal; i++ {
		p.Pair(string(rune('A'+i%26))+"host", "x", Peer{})
	}
	if p.Code() == first {
		t.Fatal("code was not rotated") // 1e-6 chance of a false failure
	}
	if _, ok := p.Authorize(token); ok {
		t.Fatal("tokens must be dropped on rotation")
	}
}

func TestIdentityFingerprintPinning(t *testing.T) {
	id, err := NewIdentity("test")
	if err != nil {
		t.Fatal(err)
	}
	if len(id.Fingerprint) != 64 {
		t.Fatalf("fingerprint %q", id.Fingerprint)
	}
	if got := ShortFingerprint("abcdef0123"); got != "ABCD-EF01" {
		t.Fatalf("ShortFingerprint = %q", got)
	}
}
