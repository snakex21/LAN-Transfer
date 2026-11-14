package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HostName is the name this computer announces. LANTRANSFER_NAME overrides
// the system host name (demos, screenshots, several instances on one PC).
func HostName() (string, error) {
	if name := strings.TrimSpace(os.Getenv("LANTRANSFER_NAME")); name != "" {
		return name, nil
	}
	return os.Hostname()
}

// NewDeviceID creates an identifier that lives only for the current process.
// This intentionally avoids persisting the ID next to the executable: the app
// is portable, so copying the program directory to another computer must never
// clone the identity of the original machine.
func NewDeviceID() string {
	removeLegacyDeviceID()

	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}

	host, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", host, time.Now().UnixNano())
}

func removeLegacyDeviceID() {
	dir, err := ApplicationDataDir()
	if err != nil {
		return
	}
	_ = os.Remove(filepath.Join(dir, "device_id"))
}
