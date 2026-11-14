package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
	"lantransfer/internal/netutil"
)

const (
	discoverMessage = "LANTRANSFER_DISCOVER_V1"
	deviceTTL       = 15 * time.Second
)

var cache = struct {
	sync.Mutex
	devices map[string]model.Device
}{devices: map[string]model.Device{}}

func Discover(wait time.Duration, selfID string) ([]model.Device, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return cachedDevices(), err
	}
	defer conn.Close()

	deadline := time.Now().Add(wait)
	if err := conn.SetDeadline(deadline); err != nil {
		return cachedDevices(), err
	}

	for _, addr := range probeTargets() {
		_, _ = conn.WriteToUDP([]byte(discoverMessage), addr)
	}
	// A second probe helps with packet loss and adapters waking up slowly.
	if wait >= 300*time.Millisecond {
		time.Sleep(80 * time.Millisecond)
		for _, addr := range probeTargets() {
			_, _ = conn.WriteToUDP([]byte(discoverMessage), addr)
		}
	}

	buf := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				break
			}
			return cachedDevices(), err
		}

		var device model.Device
		if err := json.Unmarshal(buf[:n], &device); err != nil || device.Name == "" {
			continue
		}
		if (selfID != "" && device.ID == selfID) || isOwnAddress(addr.IP) {
			continue
		}
		device.IP = addr.IP.String()
		rememberDevice(device)
	}

	devices := cachedDevices()
	if len(devices) == 0 && !config.Loopback {
		scanLocalSubnets(selfID)
		devices = cachedDevices()
	}
	return devices, nil
}

func rememberDevice(device model.Device) {
	if device.Name == "" || device.IP == "" {
		return
	}
	device.LastSeen = time.Now()
	cache.Lock()
	cache.devices[deviceKey(device)] = device
	cache.Unlock()
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

func scanLocalSubnets(selfID string) {
	prefixes := localPrivatePrefixes()
	if len(prefixes) == 0 {
		return
	}

	client := &http.Client{Timeout: 220 * time.Millisecond}
	jobs := make(chan string)
	var wg sync.WaitGroup
	workers := 64
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/info", ip, config.HTTPPort))
				if err != nil {
					continue
				}
				var device model.Device
				err = json.NewDecoder(resp.Body).Decode(&device)
				_ = resp.Body.Close()
				if err != nil || resp.StatusCode != http.StatusOK || device.Name == "" {
					continue
				}
				if selfID != "" && device.ID == selfID {
					continue
				}
				device.IP = ip
				rememberDevice(device)
			}
		}()
	}

	for _, prefix := range prefixes {
		for host := 1; host <= 254; host++ {
			ip := net.IPv4(prefix[0], prefix[1], prefix[2], byte(host))
			if netutil.IsLocalIPv4(ip) {
				continue
			}
			jobs <- ip.String()
		}
	}
	close(jobs)
	wg.Wait()
}

func localPrivatePrefixes() [][3]byte {
	seen := map[[3]byte]bool{}
	result := make([][3]byte, 0, 2)
	interfaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || !ip.IsPrivate() {
				continue
			}
			prefix := [3]byte{ip[0], ip[1], ip[2]}
			if seen[prefix] {
				continue
			}
			seen[prefix] = true
			result = append(result, prefix)
			if len(result) >= 4 {
				return result
			}
		}
	}
	return result
}

// probeTargets returns where discovery datagrams are sent. In loopback mode
// only explicitly configured local instances are contacted.
func probeTargets() []*net.UDPAddr {
	targets := make([]*net.UDPAddr, 0, 4)
	if config.Loopback {
		for _, port := range config.LoopbackPeerUDPPorts {
			if port != config.UDPPort {
				targets = append(targets, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
			}
		}
		return targets
	}
	for _, ip := range broadcastAddresses() {
		targets = append(targets, &net.UDPAddr{IP: ip, Port: config.UDPPort})
	}
	return targets
}

// isOwnAddress filters this computer's own datagrams on a real LAN. In
// loopback mode every peer lives on 127.0.0.1 and is told apart by its ID.
func isOwnAddress(ip net.IP) bool {
	return !config.Loopback && netutil.IsLocalIPv4(ip)
}

func broadcastAddresses() []net.IP {
	seen := map[string]bool{}
	add := func(ip net.IP, out *[]net.IP) {
		if ip == nil {
			return
		}
		key := ip.String()
		if seen[key] {
			return
		}
		seen[key] = true
		*out = append(*out, ip)
	}

	result := make([]net.IP, 0, 4)
	add(net.IPv4bcast, &result)

	interfaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			mask := ipNet.Mask
			if ip == nil || len(mask) != net.IPv4len {
				continue
			}
			broadcast := make(net.IP, net.IPv4len)
			for i := 0; i < net.IPv4len; i++ {
				broadcast[i] = ip[i] | ^mask[i]
			}
			add(broadcast, &result)
		}
	}
	return result
}
