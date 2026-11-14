package transfer

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"lantransfer/internal/config"

	"lantransfer/internal/discovery"
	"lantransfer/internal/netutil"
)

func resolveTarget(target, selfID string) (string, error) {
	host := target
	if h, _, err := net.SplitHostPort(target); err == nil {
		host = h
	}
	if ip := net.ParseIP(host); ip != nil {
		if !config.Loopback && netutil.IsLocalIPv4(ip) {
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
			if device.Port == 0 {
				return device.IP, nil
			}
			return net.JoinHostPort(device.IP, strconv.Itoa(device.Port)), nil
		}
	}
	return "", fmt.Errorf("nie znaleziono urządzenia %q", target)
}
