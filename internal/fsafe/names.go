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

	"lantransfer/internal/i18n"
)

// MaxNameBytes limits a single path component (NTFS/ext4 limit is 255 units).
const MaxNameBytes = 255

// MaxDepth limits how deep a received folder tree may be.
const MaxDepth = 64

// ErrUnsafeName is wrapped by every rejection so callers can map it to 400.
var ErrUnsafeName = i18n.New("err.unsafeName")

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
		return i18n.Wrap(ErrUnsafeName, "err.name.invalid", "name", name)
	case len(name) > MaxNameBytes:
		return i18n.Wrap(ErrUnsafeName, "err.name.tooLong", "max", MaxNameBytes)
	case !utf8.ValidString(name):
		return i18n.Wrap(ErrUnsafeName, "err.name.badUTF8")
	case strings.ContainsAny(name, `<>:"/\|?*`):
		return i18n.Wrap(ErrUnsafeName, "err.name.forbiddenChar", "name", name)
	case strings.HasSuffix(name, ".") || strings.HasSuffix(name, " "):
		return i18n.Wrap(ErrUnsafeName, "err.name.trailingDot", "name", name)
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f || unicode.IsControl(r) {
			return i18n.Wrap(ErrUnsafeName, "err.name.control", "name", fmt.Sprintf("%q", name))
		}
	}
	stem := name
	if dot := strings.IndexByte(stem, '.'); dot >= 0 {
		stem = stem[:dot]
	}
	if reservedNames[strings.ToUpper(strings.TrimRight(stem, " "))] {
		return i18n.Wrap(ErrUnsafeName, "err.name.reserved", "name", name)
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
		return "", i18n.Wrap(ErrUnsafeName, "err.path.empty")
	}
	if strings.Contains(rel, `\`) {
		return "", i18n.Wrap(ErrUnsafeName, "err.path.backslash")
	}
	if path.IsAbs(rel) || filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" {
		return "", i18n.Wrap(ErrUnsafeName, "err.path.absolute", "path", rel)
	}
	parts := strings.Split(rel, "/")
	if len(parts) > MaxDepth {
		return "", i18n.Wrap(ErrUnsafeName, "err.path.tooDeep", "max", MaxDepth)
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
