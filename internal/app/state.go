// Package app wires together the long-lived pieces of one running instance.
package app

import (
	"os"
	"path/filepath"

	"lantransfer/internal/config"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/transfer"
)

// State is shared by the GUI, the CLI and the HTTPS service.
type State struct {
	ID         string
	Name       string
	ReceiveDir string
	Language   LanguageInfo

	Identity *security.Identity
	Pairing  *security.Pairing
	History  *transfer.History
	Incoming *transfer.IncomingTracker
	Sender   *transfer.Manager
}

// NewState creates the per-process identity (device ID, TLS certificate,
// pairing code) and loads the transfer history.
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
	identity, err := security.NewIdentity(name)
	if err != nil {
		return nil, err
	}
	historyPath := ""
	if dir, err := config.ApplicationDataDir(); err == nil {
		historyPath = filepath.Join(dir, "history.json")
	}
	history := transfer.LoadHistory(historyPath)
	state := &State{
		ID: id, Name: name, ReceiveDir: receiveDir,
		Identity: identity,
		Pairing:  security.NewPairing(os.Getenv("LANTRANSFER_PIN")),
		History:  history,
		Incoming: transfer.NewIncomingTracker(history),
	}
	state.Sender = transfer.NewManager(state.Info(), history)
	return state, nil
}

// Info is the public description served at /api/info and in discovery.
func (s *State) Info() protocol.Info {
	return protocol.Info{
		ID: s.ID, Name: s.Name, Port: config.HTTPPort, Protocol: protocol.Version,
		Fingerprint: s.Identity.Fingerprint,
	}
}
