package transfer

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"lantransfer/internal/discovery"
	"lantransfer/internal/netutil"
)

func resolveTarget(target, selfID string) (string, error) {
	if ip := net.ParseIP(target); ip != nil {
		if netutil.IsLocalIPv4(ip) {
			return "", errors.New("nie można wysyłać plików do tego samego komputera")
		}
		return target, nil
	}

	devices, err := discovery.Discover(1200*time.Millisecond, selfID)
	if err != nil {
		return "", err
	}
	for _, device := range devices {
		if strings.EqualFold(device.Name, target) {
			return device.IP, nil
		}
	}
	return "", fmt.Errorf("nie znaleziono urządzenia %q", target)
}
