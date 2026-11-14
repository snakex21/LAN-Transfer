package app

import "lantransfer/internal/config"

type State struct {
	ID         string
	Name       string
	ReceiveDir string
}

func NewState() (*State, error) {
	id := config.NewDeviceID()
	name, err := config.HostName()
	if err != nil {
		return nil, err
	}
	receiveDir, err := config.EnsureReceiveDir()
	if err != nil {
		return nil, err
	}
	return &State{ID: id, Name: name, ReceiveDir: receiveDir}, nil
}
