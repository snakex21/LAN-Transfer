// Package discovery finds other LAN Transfer instances with UDP broadcast.
//
// Discovery is only a hint: datagrams are unauthenticated and can be spoofed
// by anyone on the network. Trust comes from pairing (code + pinned TLS
// certificate), never from what a device announces about itself.
package discovery

import (
	"encoding/json"
	"errors"
	"net"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
)

const (
	discoverMessage = "LANTRANSFER_DISCOVER_V1"
	deviceTTL       = 15 * time.Second
	// maxDevices bounds the cache so a flood of fake announcements cannot
	// exhaust memory or bury real devices in the UI.
	maxDevices    = 64
	maxNameLength = 64
)

var cache = struct {
	sync.Mutex
	devices map[string]model.Device
}{devices: map[string]model.Device{}}

// Discover broadcasts a probe, collects answers for `wait` and returns every
// device seen recently (including unsolicited announcements).
func Discover(wait time.Duration, selfID string) ([]model.Device, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(bindIP()), Port: 0})
	if err != nil {
		return cachedDevices(), err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(wait)); err != nil {
		return cachedDevices(), err
	}
	sendProbes(conn)
	// A second probe helps with packet loss and adapters waking up slowly.
	if wait >= 300*time.Millisecond {
		time.Sleep(80 * time.Millisecond)
		sendProbes(conn)
	}

	buf := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				break
			}
			return cachedDevices(), err
		}
		if device, ok := parseAnnouncement(buf[:n], addr.IP, selfID); ok {
			rememberDevice(device)
		}
	}

	devices := cachedDevices()
	if len(devices) == 0 && !config.Loopback {
		scanLocalSubnets(selfID)
		devices = cachedDevices()
	}
	return devices, nil
}

func sendProbes(conn *net.UDPConn) {
	for _, addr := range probeTargets() {
		_, _ = conn.WriteToUDP([]byte(discoverMessage), addr)
	}
}

// parseAnnouncement validates a device datagram from ip.
func parseAnnouncement(data []byte, ip net.IP, selfID string) (model.Device, bool) {
	var device model.Device
	if err := json.Unmarshal(data, &device); err != nil {
		return device, false
	}
	if device.Name == "" || len(device.Name) > maxNameLength || !utf8.ValidString(device.Name) ||
		len(device.ID) > 64 || len(device.Fingerprint) > 64 || device.Port < 0 || device.Port > 65535 {
		return device, false
	}
	if (selfID != "" && device.ID == selfID) || isOwnAddress(ip) {
		return device, false
	}
	device.IP = ip.String()
	return device, true
}

func rememberDevice(device model.Device) {
	if device.Name == "" || device.IP == "" {
		return
	}
	device.LastSeen = time.Now()
	key := deviceKey(device)
	cache.Lock()
	defer cache.Unlock()
	if _, known := cache.devices[key]; !known && len(cache.devices) >= maxDevices {
		evictOldestLocked()
	}
	cache.devices[key] = device
}

func evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	for key, device := range cache.devices {
		if oldestKey == "" || device.LastSeen.Before(oldest) {
			oldestKey, oldest = key, device.LastSeen
		}
	}
	delete(cache.devices, oldestKey)
}

func deviceKey(device model.Device) string {
	if device.ID != "" {
		return device.ID
	}
	return device.IP
}

func cachedDevices() []model.Device {
	now := time.Now()
	cache.Lock()
	defer cache.Unlock()

	devices := make([]model.Device, 0, len(cache.devices))
	for key, device := range cache.devices {
		if now.Sub(device.LastSeen) > deviceTTL {
			delete(cache.devices, key)
			continue
		}
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Name == devices[j].Name {
			return devices[i].IP < devices[j].IP
		}
		return devices[i].Name < devices[j].Name
	})
	return devices
}
