package transfer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Job and history statuses.
const (
	StatusQueued      = "queued"
	StatusRunning     = "running"
	StatusDone        = "done"
	StatusFailed      = "failed"
	StatusCancelled   = "cancelled"
	StatusInterrupted = "interrupted"
)

// maxHistory is the number of entries kept on disk.
const maxHistory = 200

// HistoryEntry describes one finished (or abandoned) transfer.
type HistoryEntry struct {
	Direction string    `json:"direction"` // "sent" or "received"
	Name      string    `json:"name"`
	Peer      string    `json:"peer"`
	Bytes     int64     `json:"bytes"`
	Files     int       `json:"files"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Started   time.Time `json:"started"`
	Finished  time.Time `json:"finished"`
	Location  string    `json:"location,omitempty"`
	JobID     string    `json:"job_id,omitempty"`
}

// History keeps recent transfers in memory and, when path is set, in a JSON file.
type History struct {
	mu      sync.Mutex
	path    string
	entries []HistoryEntry
}

// LoadHistory reads path (missing or damaged files start an empty history).
func LoadHistory(path string) *History {
	h := &History{path: path}
	if path == "" {
		return h
	}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &h.entries)
	}
	return h
}

// Add prepends an entry and persists the history.
func (h *History) Add(entry HistoryEntry) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.entries = append([]HistoryEntry{entry}, h.entries...)
	if len(h.entries) > maxHistory {
		h.entries = h.entries[:maxHistory]
	}
	entries := append([]HistoryEntry(nil), h.entries...)
	h.mu.Unlock()
	h.save(entries)
}

// Entries returns a copy, newest first.
func (h *History) Entries() []HistoryEntry {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]HistoryEntry(nil), h.entries...)
}

// Clear removes every entry.
func (h *History) Clear() {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.entries = nil
	h.mu.Unlock()
	h.save(nil)
}

func (h *History) save(entries []HistoryEntry) {
	if h.path == "" {
		return
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(h.path), 0o755); err != nil {
		return
	}
	tmp := h.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err == nil {
		if err := os.Rename(tmp, h.path); err != nil {
			_ = os.Remove(tmp)
		}
	}
}

// FormatBytes renders a size the way the UI does ("1.5 GB").
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, exp := float64(n), 0
	for value >= unit && exp < 4 {
		value /= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", value, []string{"B", "KB", "MB", "GB", "TB"}[exp])
}
