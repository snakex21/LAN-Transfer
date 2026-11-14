package transfer

import (
	"net"
	"strings"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/i18n"
	"lantransfer/internal/netutil"
)

// ResolveTarget turns an IP, "IP:port" or a discovered device name into a
// "host:port" address.
func ResolveTarget(target, selfID string) (string, error) {
	host := target
	if h, _, err := net.SplitHostPort(target); err == nil {
		host = h
	}
	if ip := net.ParseIP(host); ip != nil {
		if !config.Loopback && netutil.IsLocalIPv4(ip) {
			return "", i18n.New("err.sameComputer")
		}
		return config.TargetAddress(target), nil
	}

	devices, err := discovery.Discover(1200*time.Millisecond, selfID)
	if err != nil {
		return "", err
	}
	for _, device := range devices {
		if strings.EqualFold(device.Name, target) {
			return device.Address(), nil
		}
	}
	return "", i18n.New("err.deviceNotFound", "name", target)
}
