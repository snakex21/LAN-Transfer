package config

import (
	"net"
	"os"
	"strconv"
	"strings"
)

// Default network ports used on a real LAN.
const (
	DefaultUDPPort  = 45454
	DefaultHTTPPort = 45455
)

// Runtime settings. They default to the production values and can be
// overridden through environment variables, which lets tests and CI run two
// independent instances on one machine using only the loopback interface.
var (
	// UDPPort is the discovery port (LANTRANSFER_UDP_PORT).
	UDPPort = envInt("LANTRANSFER_UDP_PORT", DefaultUDPPort)
	// HTTPPort is the transfer/API port (LANTRANSFER_PORT).
	HTTPPort = envInt("LANTRANSFER_PORT", DefaultHTTPPort)
	// Loopback confines the instance to 127.0.0.1 (LANTRANSFER_LOOPBACK=1):
	// it listens only on loopback, never broadcasts or scans the LAN and
	// allows sending to 127.0.0.1. Intended for tests and demos.
	Loopback = envBool("LANTRANSFER_LOOPBACK")
	// LoopbackPeerUDPPorts lists discovery ports of other loopback instances
	// (LANTRANSFER_PEER_UDP_PORTS=45464,45474).
	LoopbackPeerUDPPorts = envInts("LANTRANSFER_PEER_UDP_PORTS")
	// SkipFirewall disables the one-time firewall helper (LANTRANSFER_NO_FIREWALL=1).
	SkipFirewall = envBool("LANTRANSFER_NO_FIREWALL") || Loopback
)

// BindHost returns the interface address the services listen on.
func BindHost() string {
	if Loopback {
		return "127.0.0.1"
	}
	return ""
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 || value > 65535 {
		return fallback
	}
	return value
}

func envBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func envInts(name string) []int {
	var result []int
	for _, field := range strings.Split(os.Getenv(name), ",") {
		if value, err := strconv.Atoi(strings.TrimSpace(field)); err == nil && value > 0 && value <= 65535 {
			result = append(result, value)
		}
	}
	return result
}

// TargetAddress turns "host" or "host:port" into "host:port", defaulting to
// this instance's HTTP port.
func TargetAddress(target string) string {
	if _, _, err := net.SplitHostPort(target); err == nil {
		return target
	}
	return net.JoinHostPort(target, strconv.Itoa(HTTPPort))
}
