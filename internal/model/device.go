package model

import (
	"net"
	"strconv"
	"time"

	"lantransfer/internal/config"
)

// Device is a peer found by discovery.
type Device struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	Port        int       `json:"port"`
	Protocol    int       `json:"protocol"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	NoCode      bool      `json:"no_code,omitempty"` // receiver needs no pairing code
	LastSeen    time.Time `json:"last_seen"`
}

// Address returns "ip:port".
func (d Device) Address() string {
	port := d.Port
	if port <= 0 || port > 65535 {
		port = config.HTTPPort
	}
	return net.JoinHostPort(d.IP, strconv.Itoa(port))
}
