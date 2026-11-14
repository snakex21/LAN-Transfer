package discovery

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
	"lantransfer/internal/netutil"
)

// scanInterval throttles the fallback subnet scan. The UI refreshes the device
// list every few seconds; without the throttle an empty network was probed
// with 254 connections per subnet on every refresh.
const scanInterval = time.Minute

var lastScan = struct {
	sync.Mutex
	at time.Time
}{}

// scanLocalSubnets is a fallback for networks that drop broadcasts (some
// Wi-Fi access points): it asks every host of the local private /24 networks
// for /api/info.
func scanLocalSubnets(selfID string) {
	lastScan.Lock()
	if time.Since(lastScan.at) < scanInterval {
		lastScan.Unlock()
		return
	}
	lastScan.at = time.Now()
	lastScan.Unlock()

	prefixes := localPrivatePrefixes()
	if len(prefixes) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			// Only the public /api/info is read here; identity is verified
			// later by pairing, so the self-signed certificate is accepted.
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS13},
			DisableKeepAlives: true,
		},
	}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				probeInfo(client, ip, selfID)
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

func probeInfo(client *http.Client, ip, selfID string) {
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/api/info", ip, config.HTTPPort))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	if device, ok := parseAnnouncement(data, net.ParseIP(ip), selfID); ok {
		rememberDevice(device)
	}
}

func localPrivatePrefixes() [][3]byte {
	seen := map[[3]byte]bool{}
	result := make([][3]byte, 0, 2)
	for _, ipNet := range activeIPv4Networks() {
		if !ipNet.IP.IsPrivate() {
			continue
		}
		prefix := [3]byte{ipNet.IP[0], ipNet.IP[1], ipNet.IP[2]}
		if seen[prefix] {
			continue
		}
		seen[prefix] = true
		result = append(result, prefix)
		if len(result) >= 4 {
			break
		}
	}
	return result
}

// devicePayload is the JSON announced over UDP.
func devicePayload(d model.Device) ([]byte, error) {
	return json.Marshal(d)
}
