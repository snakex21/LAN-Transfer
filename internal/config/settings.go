package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings are the user preferences kept in data/settings.json.
type Settings struct {
	// Language is the UI language code chosen in the settings dialog.
	Language string `json:"language,omitempty"`
	// LanguageSource is "manual" (picked by the user) or "auto" (follow the
	// Windows display language on every start).
	LanguageSource string `json:"language_source,omitempty"`
}

func settingsPath() (string, error) {
	dir, err := ApplicationDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// LoadSettings reads data/settings.json; a missing or damaged file gives
// the defaults.
func LoadSettings() Settings {
	var settings Settings
	path, err := settingsPath()
	if err != nil {
		return settings
	}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	return settings
}

// SaveSettings writes data/settings.json atomically.
func SaveSettings(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}
