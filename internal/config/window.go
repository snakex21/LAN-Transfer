package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	DefaultWindowWidth  = 1100
	DefaultWindowHeight = 760
)

type WindowState struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

func LoadWindowState() (WindowState, bool) {
	state := WindowState{Width: DefaultWindowWidth, Height: DefaultWindowHeight}
	dir, err := ApplicationDataDir()
	if err != nil {
		return state, false
	}

	data, err := os.ReadFile(filepath.Join(dir, "window_state.json"))
	if err != nil {
		return state, false
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return WindowState{Width: DefaultWindowWidth, Height: DefaultWindowHeight}, false
	}
	if state.Width < 700 || state.Height < 520 {
		return WindowState{Width: DefaultWindowWidth, Height: DefaultWindowHeight}, false
	}
	return state, true
}

func SaveWindowState(state WindowState) error {
	if state.Width < 700 || state.Height < 520 {
		return errors.New("invalid window size")
	}

	dir, err := ApplicationDataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "window_state.json")
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}
