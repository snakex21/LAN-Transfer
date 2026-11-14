package model

import "time"

type Device struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	LastSeen time.Time `json:"last_seen"`
}
