// Package fsafe validates file names that arrive from the network and picks
// collision-free destination paths.
package fsafe

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxNameBytes limits a single path component (NTFS/ext4 limit is 255 units).
const MaxNameBytes = 255

// MaxDepth limits how deep a received folder tree may be.
const MaxDepth = 64

// ErrUnsafeName is wrapped by every rejection so callers can map it to 400.
var ErrUnsafeName = errors.New("niebezpieczna nazwa pliku")

var reservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true, "CONIN$": true, "CONOUT$": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"COM¹": true, "COM²": true, "COM³": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	"LPT¹": true, "LPT²": true, "LPT³": true,
}

// CheckName validates one path component received from a peer. It rejects
// anything that could escape the destination folder or address something other
// than a regular file on Windows: separators, "..", drive letters, alternate
// data streams (":"), device names (CON, NUL, COM1…), control characters and
// trailing dots/spaces that Windows silently strips.
func CheckName(name string) error {
	switch {
	case name == "" || name == "." || name == "..":
		return fmt.Errorf("%w: %q", ErrUnsafeName, name)
	case len(name) > MaxNameBytes:
		return fmt.Errorf("%w: nazwa dłuższa niż %d bajtów", ErrUnsafeName, MaxNameBytes)
	case !utf8.ValidString(name):
		return fmt.Errorf("%w: nieprawidłowe UTF-8", ErrUnsafeName)
	case strings.ContainsAny(name, `<>:"/\|?*`):
		return fmt.Errorf("%w: %q zawiera znak niedozwolony w Windows", ErrUnsafeName, name)
	case strings.HasSuffix(name, ".") || strings.HasSuffix(name, " "):
		return fmt.Errorf("%w: %q kończy się kropką lub spacją", ErrUnsafeName, name)
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f || unicode.IsControl(r) {
			return fmt.Errorf("%w: %q zawiera znak sterujący", ErrUnsafeName, name)
		}
	}
	stem := name
	if dot := strings.IndexByte(stem, '.'); dot >= 0 {
		stem = stem[:dot]
	}
	if reservedNames[strings.ToUpper(strings.TrimRight(stem, " "))] {
		return fmt.Errorf("%w: %q to zarezerwowana nazwa urządzenia Windows", ErrUnsafeName, name)
	}
	return nil
}

// CleanRelPath validates a slash-separated relative path received from a peer
// ("Folder/sub/file.txt") and returns it in OS form. Absolute paths, UNC
// paths, drive letters and any ".." component are rejected rather than
// silently rewritten, so a malicious name can never land outside the
// destination folder.
func CleanRelPath(rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("%w: pusta ścieżka", ErrUnsafeName)
	}
	if strings.Contains(rel, `\`) {
		return "", fmt.Errorf("%w: używaj '/' jako separatora", ErrUnsafeName)
	}
	if path.IsAbs(rel) || filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" {
		return "", fmt.Errorf("%w: ścieżka bezwzględna %q", ErrUnsafeName, rel)
	}
	parts := strings.Split(rel, "/")
	if len(parts) > MaxDepth {
		return "", fmt.Errorf("%w: zbyt głęboka struktura folderów", ErrUnsafeName)
	}
	for _, part := range parts {
		if err := CheckName(part); err != nil {
			return "", err
		}
	}
	return filepath.Join(parts...), nil
}

// Within reports whether target is root itself or lies inside root.
func Within(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}

// UniquePath returns p when nothing exists there, otherwise the first free
// "name (n).ext" variant. It never overwrites an existing file or folder.
func UniquePath(p string) string {
	if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) {
		return p
	}
	dir, base := filepath.Split(p)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == "" { // ".env"
		stem, ext = base, ""
	}
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if _, err := os.Lstat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}
