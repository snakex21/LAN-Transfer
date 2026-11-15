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
// Info is read for every announcement, so a changed setting (pairing code on
// or off) is visible to others within one announce interval.
type Responder struct {
	Info func() protocol.Info
}

func (r Responder) payload() []byte {
	info := r.Info()
	payload, err := devicePayload(model.Device{
		ID: info.ID, Name: info.Name, Port: info.Port,
		Protocol: info.Protocol, Fingerprint: info.Fingerprint, NoCode: info.NoCode,
	})
	if err != nil {
		log.Println("discovery payload:", err)
	}
	return payload
}

// Run blocks forever (until the UDP socket fails to open).
func (r Responder) Run() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(bindIP()), Port: config.UDPPort})
	if err != nil {
		log.Println("discovery:", err)
		return
	}
	defer conn.Close()

	selfID := r.Info().ID
	go func() {
		announce := func() {
			payload := r.payload()
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
			_, _ = conn.WriteToUDP(r.payload(), remote)
			continue
		}
		if device, ok := parseAnnouncement(buf[:n], remote.IP, selfID); ok {
			rememberDevice(device)
		}
	}
}
