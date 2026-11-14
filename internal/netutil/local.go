package netutil

import "net"

func IsLocalIPv4(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		var candidate net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			candidate = v.IP
		case *net.IPAddr:
			candidate = v.IP
		}
		if c4 := candidate.To4(); c4 != nil && c4.Equal(ip) {
			return true
		}
	}
	return false
}
