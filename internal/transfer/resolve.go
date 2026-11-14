package transfer

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
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
			return "", errors.New("nie można wysyłać plików do tego samego komputera")
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
	return "", fmt.Errorf("nie znaleziono urządzenia %q", target)
}
