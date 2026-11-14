package discovery

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
	"lantransfer/internal/netutil"
)

type Responder struct {
	ID   string
	Name string
}

func (r Responder) Run() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: config.UDPPort})
	if err != nil {
		log.Println("discovery:", err)
		return
	}
	defer conn.Close()

	payload, err := json.Marshal(model.Device{ID: r.ID, Name: r.Name, Port: config.HTTPPort})
	if err != nil {
		log.Println("discovery payload:", err)
		return
	}

	// Poza odpowiedzią na zapytania urządzenie samo ogłasza swoją obecność.
	// Dzięki temu pojedynczy zgubiony broadcast nie powoduje znikania komputera.
	go func() {
		announce := func() {
			for _, ip := range broadcastAddresses() {
				_, _ = conn.WriteToUDP(payload, &net.UDPAddr{IP: ip, Port: config.UDPPort})
			}
		}
		announce()
		ticker := time.NewTicker(2 * time.Second)
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

		message := buf[:n]
		if string(message) == discoverMessage {
			_, _ = conn.WriteToUDP(payload, remote)
			continue
		}

		var device model.Device
		if err := json.Unmarshal(message, &device); err != nil || device.Name == "" {
			continue
		}
		if device.ID == r.ID || netutil.IsLocalIPv4(remote.IP) {
			continue
		}
		device.IP = remote.IP.String()
		rememberDevice(device)
	}
}
