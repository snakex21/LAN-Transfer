package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"lantransfer/internal/i18n"
)

const (
	// MaxFailuresPerHost wrong codes from one address trigger a lockout.
	MaxFailuresPerHost = 5
	// LockoutDuration is how long a host must wait after too many failures.
	LockoutDuration = time.Minute
	// MaxFailuresTotal wrong codes overall rotate the pairing code, which
	// defeats slow guessing spread across many addresses.
	MaxFailuresTotal = 30
	// MaxPeers caps the number of simultaneously paired senders.
	MaxPeers = 256
)

var (
	ErrWrongCode = i18n.New("err.wrongCode")
	ErrLocked    = i18n.New("err.locked")
	// ErrCodeRequired answers a request without a code while the receiver
	// requires one. It is not counted as a wrong guess.
	ErrCodeRequired = i18n.New("err.codeRequired")
)

// Peer describes a sender that proved knowledge of the pairing code.
type Peer struct {
	DeviceID string    `json:"device_id"`
	Name     string    `json:"name"`
	Address  string    `json:"address"`
	PairedAt time.Time `json:"paired_at"`
	// NoCode marks a token issued while the code was switched off.
	NoCode bool `json:"no_code,omitempty"`
}

// Pairing issues bearer tokens to peers that present the current pairing code.
// Everything lives in memory: restarting the receiver forgets all peers.
type Pairing struct {
	mu          sync.Mutex
	code        string
	fixedCode   bool
	open        bool // code switched off: every sender gets a token
	tokens      map[string]Peer
	failures    map[string]int
	lockedUntil map[string]time.Time
	totalFails  int
	now         func() time.Time
	onRotate    func()
}

// NewPairing creates a pairing state. A non-empty fixed code (for headless
// `serve` automation) is never rotated; otherwise a random 6-digit code is used.
func NewPairing(fixed string) *Pairing {
	p := &Pairing{
		tokens:      map[string]Peer{},
		failures:    map[string]int{},
		lockedUntil: map[string]time.Time{},
		now:         time.Now,
	}
	if fixed != "" {
		p.code, p.fixedCode = fixed, true
	} else {
		p.code = randomCode()
	}
	return p
}

// Code returns the code that has to be typed on the sending computer.
func (p *Pairing) Code() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.code
}

// OnRotate registers a callback fired after the code changes.
func (p *Pairing) OnRotate(fn func()) {
	p.mu.Lock()
	p.onRotate = fn
	p.mu.Unlock()
}

// CodeRequired reports whether senders must present the pairing code.
func (p *Pairing) CodeRequired() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.open
}

// SetCodeRequired switches the pairing code on or off. With the code off any
// LAN Transfer sender on the network gets a token without a code (the 1.x
// behaviour, now over TLS). Switching it back on drops those tokens.
func (p *Pairing) SetCodeRequired(required bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if required && p.open {
		for token, peer := range p.tokens {
			if peer.NoCode {
				delete(p.tokens, token)
			}
		}
	}
	p.open = !required
}

// Pair checks code and returns a fresh token for the peer. While the code is
// switched off every request succeeds.
func (p *Pairing) Pair(host, code string, peer Peer) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	if p.open {
		peer.NoCode = subtle.ConstantTimeCompare([]byte(normalizeCode(code)), []byte(p.code)) != 1
		return p.issueLocked(host, peer, now), nil
	}
	if normalizeCode(code) == "" {
		return "", ErrCodeRequired
	}
	if until, ok := p.lockedUntil[host]; ok {
		if now.Before(until) {
			return "", ErrLocked
		}
		delete(p.lockedUntil, host)
		delete(p.failures, host)
	}
	if subtle.ConstantTimeCompare([]byte(normalizeCode(code)), []byte(p.code)) != 1 {
		p.failures[host]++
		p.totalFails++
		if p.failures[host] >= MaxFailuresPerHost {
			p.lockedUntil[host] = now.Add(LockoutDuration)
		}
		if p.totalFails >= MaxFailuresTotal && !p.fixedCode {
			p.rotateLocked()
		}
		return "", ErrWrongCode
	}
	delete(p.failures, host)
	peer.NoCode = false
	return p.issueLocked(host, peer, now), nil
}

func (p *Pairing) issueLocked(host string, peer Peer, now time.Time) string {
	if len(p.tokens) >= MaxPeers {
		p.evictOldestLocked()
	}
	token := randomToken()
	peer.Address = host
	peer.PairedAt = now
	p.tokens[token] = peer
	return token
}

// Authorize returns the peer that owns token.
func (p *Pairing) Authorize(token string) (Peer, bool) {
	if token == "" {
		return Peer{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	peer, ok := p.tokens[token]
	return peer, ok
}

// Rotate generates a new code and forgets every paired peer.
func (p *Pairing) Rotate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rotateLocked()
}

// Peers lists currently paired senders.
func (p *Pairing) Peers() []Peer {
	p.mu.Lock()
	defer p.mu.Unlock()
	peers := make([]Peer, 0, len(p.tokens))
	for _, peer := range p.tokens {
		peers = append(peers, peer)
	}
	return peers
}

func (p *Pairing) rotateLocked() {
	if !p.fixedCode {
		p.code = randomCode()
	}
	p.tokens = map[string]Peer{}
	p.failures = map[string]int{}
	p.totalFails = 0
	if p.onRotate != nil {
		go p.onRotate()
	}
}

func (p *Pairing) evictOldestLocked() {
	var oldestToken string
	var oldest time.Time
	for token, peer := range p.tokens {
		if oldestToken == "" || peer.PairedAt.Before(oldest) {
			oldestToken, oldest = token, peer.PairedAt
		}
	}
	delete(p.tokens, oldestToken)
}

func normalizeCode(code string) string {
	out := make([]byte, 0, len(code))
	for i := 0; i < len(code); i++ {
		if c := code[i]; c != ' ' && c != '-' {
			out = append(out, c)
		}
	}
	return string(out)
}

func randomCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
