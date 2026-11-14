package discovery

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"lantransfer/internal/config"
	"lantransfer/internal/model"
)

func bindIP() string {
	if host := config.BindHost(); host != "" {
		return host
	}
	return "0.0.0.0"
}

type Responder struct {
	ID   string
	Name string
}

func (r Responder) Run() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(bindIP()), Port: config.UDPPort})
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
			for _, addr := range probeTargets() {
				_, _ = conn.WriteToUDP(payload, addr)
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
		if device.ID == r.ID || isOwnAddress(remote.IP) {
			continue
		}
		device.IP = remote.IP.String()
		rememberDevice(device)
	}
}
