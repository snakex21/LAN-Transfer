package discovery

import (
	"net"

	"lantransfer/internal/config"
	"lantransfer/internal/netutil"
)

func bindIP() string {
	if host := config.BindHost(); host != "" {
		return host
	}
	return "0.0.0.0"
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

// broadcastAddresses returns the limited broadcast address plus the directed
// broadcast address of every active, non-loopback IPv4 interface.
func broadcastAddresses() []net.IP {
	seen := map[string]bool{}
	result := make([]net.IP, 0, 4)
	add := func(ip net.IP) {
		if key := ip.String(); !seen[key] {
			seen[key] = true
			result = append(result, ip)
		}
	}
	add(net.IPv4bcast)
	for _, ipNet := range activeIPv4Networks() {
		broadcast := make(net.IP, net.IPv4len)
		for i := 0; i < net.IPv4len; i++ {
			broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
		}
		add(broadcast)
	}
	return result
}

// activeIPv4Networks lists IPv4 networks of interfaces that are up.
func activeIPv4Networks() []net.IPNet {
	var result []net.IPNet
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
			if ip == nil || len(ipNet.Mask) != net.IPv4len {
				continue
			}
			result = append(result, net.IPNet{IP: ip, Mask: ipNet.Mask})
		}
	}
	return result
}
