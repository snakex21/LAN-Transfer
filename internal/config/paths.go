package config

import (
	"os"
	"path/filepath"
)

func ApplicationDataDir() (string, error) {
	if dir := os.Getenv("LANTRANSFER_DATA_DIR"); dir != "" {
		return dir, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "data"), nil
}

func EnsureReceiveDir() (string, error) {
	dir := os.Getenv("LANTRANSFER_RECEIVE_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, "Downloads", "LAN Transfer")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
