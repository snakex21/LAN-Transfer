package preview

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	maxInlineBytes = 20 * 1024 * 1024
	maxTextBytes   = 512 * 1024
	maxFolderItems = 40
)

type Info struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	MIME       string   `json:"mime"`
	Size       int64    `json:"size"`
	Modified   string   `json:"modified"`
	DataURL    string   `json:"data_url,omitempty"`
	Text       string   `json:"text,omitempty"`
	Truncated  bool     `json:"truncated,omitempty"`
	EntryCount int      `json:"entry_count,omitempty"`
	Entries    []string `json:"entries,omitempty"`
	Message    string   `json:"message,omitempty"`
}

func Load(path string) (Info, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || clean == "" {
		return Info{}, errors.New("brak ścieżki do podglądu")
	}

	stat, err := os.Stat(clean)
	if err != nil {
		return Info{}, err
	}

	info := Info{
		Path:     clean,
		Name:     stat.Name(),
		Size:     stat.Size(),
		Modified: stat.ModTime().Format("2006-01-02 15:04:05"),
	}

	if stat.IsDir() {
		return loadFolder(clean, info)
	}
	return loadFile(clean, info)
}

func loadFolder(path string, info Info) (Info, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return Info{}, err
	}
	info.Kind = "folder"
	info.MIME = "inode/directory"
	info.EntryCount = len(entries)

	names := make([]string, 0, min(len(entries), maxFolderItems))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += string(os.PathSeparator)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > maxFolderItems {
		names = names[:maxFolderItems]
		info.Truncated = true
	}
	info.Entries = names
	return info, nil
}

func loadFile(path string, info Info) (Info, error) {
	file, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer file.Close()

	head := make([]byte, 512)
	n, readErr := file.Read(head)
	if readErr != nil && n == 0 {
		return Info{}, readErr
	}
	head = head[:n]

	mime := http.DetectContentType(head)
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".svg" {
		mime = "image/svg+xml"
	} else if ext == ".pdf" {
		mime = "application/pdf"
	}
	info.MIME = mime

	switch {
	case strings.HasPrefix(mime, "image/"):
		info.Kind = "image"
		if info.Size > maxInlineBytes {
			info.Message = "Plik obrazu jest za duży na szybki podgląd w aplikacji."
			return info, nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Info{}, err
		}
		info.DataURL = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
		return info, nil

	case mime == "application/pdf":
		info.Kind = "pdf"
		if info.Size > maxInlineBytes {
			info.Message = "PDF jest za duży na szybki podgląd w aplikacji."
			return info, nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Info{}, err
		}
		info.DataURL = "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(data)
		return info, nil

	case isTextFile(ext, mime, head):
		info.Kind = "text"
		data, truncated, err := readText(path)
		if err != nil {
			return Info{}, err
		}
		info.Text = data
		info.Truncated = truncated
		return info, nil

	default:
		info.Kind = "file"
		info.Message = "Dla tego typu pliku dostępne są informacje, ale nie podgląd zawartości."
		return info, nil
	}
}

func readText(path string) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	buffer := make([]byte, maxTextBytes+1)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return "", false, err
	}
	truncated := n > maxTextBytes
	if truncated {
		n = maxTextBytes
	}
	text := string(buffer[:n])
	if !utf8.ValidString(text) {
		return "", false, fmt.Errorf("plik nie zawiera tekstu UTF-8")
	}
	return text, truncated, nil
}

func isTextFile(ext, mime string, head []byte) bool {
	if strings.HasPrefix(mime, "text/") {
		return true
	}
	textExt := map[string]bool{
		".txt": true, ".md": true, ".log": true, ".csv": true, ".json": true, ".xml": true,
		".html": true, ".htm": true, ".css": true, ".js": true, ".ts": true, ".go": true,
		".py": true, ".java": true, ".c": true, ".h": true, ".cpp": true, ".hpp": true,
		".rs": true, ".toml": true, ".yaml": true, ".yml": true, ".ini": true, ".cfg": true,
		".bat": true, ".cmd": true, ".ps1": true, ".sh": true, ".sql": true,
	}
	if textExt[ext] {
		return true
	}
	if len(head) == 0 {
		return false
	}
	return utf8.Valid(head) && !containsNUL(head)
}

func containsNUL(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
