package discovery

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"lantransfer/internal/model"
)

func TestParseAnnouncementValidates(t *testing.T) {
	ip := net.IPv4(10, 1, 2, 3)
	device, ok := parseAnnouncement([]byte(`{"id":"abc","name":"Laptop","port":45455,"protocol":2,"fingerprint":"ff"}`), ip, "self")
	if !ok || device.IP != "10.1.2.3" || device.Protocol != 2 || device.Address() != "10.1.2.3:45455" {
		t.Fatalf("got %+v %v", device, ok)
	}
	bad := []string{
		`not json`,
		`{"name":""}`,
		`{"name":"` + strings.Repeat("x", 65) + `"}`,
		`{"name":"a","port":70000}`,
		`{"name":"a","id":"self"}`,
		`{"name":"a","id":"` + strings.Repeat("i", 65) + `"}`,
	}
	for _, payload := range bad {
		if _, ok := parseAnnouncement([]byte(payload), ip, "self"); ok {
			t.Errorf("accepted %s", payload)
		}
	}
}

func TestDeviceCacheIsBounded(t *testing.T) {
	cache.Lock()
	cache.devices = map[string]model.Device{}
	cache.Unlock()
	for i := 0; i < maxDevices*3; i++ {
		rememberDevice(model.Device{ID: fmt.Sprint(i), Name: "fake", IP: "10.0.0.1"})
	}
	if n := len(cachedDevices()); n != maxDevices {
		t.Fatalf("cache holds %d devices, want %d", n, maxDevices)
	}
}

func TestBroadcastAddressesIncludeLimitedBroadcast(t *testing.T) {
	addrs := broadcastAddresses()
	if len(addrs) == 0 || !addrs[0].Equal(net.IPv4bcast) {
		t.Fatalf("got %v", addrs)
	}
}
