// Package security holds the TLS identity of this instance and the pairing
// logic that decides which peers may send files or browse folders.
package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"lantransfer/internal/i18n"
)

// Identity is an ephemeral, self-signed TLS certificate generated at start-up.
// It is never written to disk, in line with the per-process device ID: copying
// the portable program folder never clones a machine's identity.
type Identity struct {
	Certificate tls.Certificate
	Fingerprint string // hex SHA-256 of the DER certificate
}

// NewIdentity creates a fresh ECDSA P-256 certificate valid for 30 days.
func NewIdentity(commonName string) (*Identity, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "LAN Transfer " + commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(30 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(der)
	return &Identity{
		Certificate: tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key},
		Fingerprint: hex.EncodeToString(sum[:]),
	}, nil
}

// ServerConfig returns the TLS configuration for the transfer server.
func (id *Identity) ServerConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{id.Certificate},
		MinVersion:   tls.VersionTLS13,
	}
}

// ShortFingerprint formats the first 8 hex digits as "ABCD-EF12" so people can
// compare the receiver's identity shown on both screens.
func ShortFingerprint(fingerprint string) string {
	if len(fingerprint) < 8 {
		return strings.ToUpper(fingerprint)
	}
	f := strings.ToUpper(fingerprint[:8])
	return f[:4] + "-" + f[4:]
}

// ErrFingerprintMismatch means the peer presented a different certificate than
// the one recorded when it was paired (possible man-in-the-middle, or the peer
// simply restarted).
var ErrFingerprintMismatch = i18n.New("err.fingerprintMismatch")

// ClientConfig returns a TLS configuration that accepts the self-signed
// certificate of a peer. When expected is non-empty the peer's certificate
// must have exactly that fingerprint (trust on first use, pinned at pairing).
// The observed fingerprint is reported through seen when it is not nil.
func ClientConfig(expected string, seen func(string)) *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // self-signed; verified by fingerprint below
		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return i18n.New("err.noCertificate")
			}
			sum := sha256.Sum256(state.PeerCertificates[0].Raw)
			actual := hex.EncodeToString(sum[:])
			if seen != nil {
				seen(actual)
			}
			if expected != "" && !strings.EqualFold(expected, actual) {
				return i18n.Wrap(ErrFingerprintMismatch, "err.fingerprintChanged",
					"actual", ShortFingerprint(actual), "expected", ShortFingerprint(expected))
			}
			return nil
		},
	}
}
