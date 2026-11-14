package discovery

import (
	"log"
	"net"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
	"lantransfer/internal/protocol"
)

// announceInterval is how often an instance announces itself unprompted, so a
// single lost broadcast does not make a computer disappear from the list.
const announceInterval = 2 * time.Second

// Responder answers discovery probes and periodically announces this instance.
type Responder struct {
	Info protocol.Info
}

// Run blocks forever (until the UDP socket fails to open).
func (r Responder) Run() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(bindIP()), Port: config.UDPPort})
	if err != nil {
		log.Println("discovery:", err)
		return
	}
	defer conn.Close()

	payload, err := devicePayload(model.Device{
		ID: r.Info.ID, Name: r.Info.Name, Port: r.Info.Port,
		Protocol: r.Info.Protocol, Fingerprint: r.Info.Fingerprint,
	})
	if err != nil {
		log.Println("discovery payload:", err)
		return
	}

	go func() {
		announce := func() {
			for _, addr := range probeTargets() {
				_, _ = conn.WriteToUDP(payload, addr)
			}
		}
		announce()
		ticker := time.NewTicker(announceInterval)
		defer ticker.Stop()
		for range ticker.C {
			announce()
		}
	}()

	buf := make([]byte, 2048)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if string(buf[:n]) == discoverMessage {
			_, _ = conn.WriteToUDP(payload, remote)
			continue
		}
		if device, ok := parseAnnouncement(buf[:n], remote.IP, r.Info.ID); ok {
			rememberDevice(device)
		}
	}
}
