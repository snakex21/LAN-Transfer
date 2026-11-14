package filebrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Browse(path string) (Listing, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return browseRoots()
	}

	clean, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return Listing{}, err
	}
	info, err := os.Stat(clean)
	if err != nil {
		return Listing{}, err
	}
	if !info.IsDir() {
		return Listing{}, fmt.Errorf("%q nie jest folderem", clean)
	}

	items, err := os.ReadDir(clean)
	if err != nil {
		return Listing{}, err
	}
	entries := make([]Entry, 0, len(items))
	for _, item := range items {
		if !item.IsDir() {
			continue
		}
		entries = append(entries, Entry{Name: item.Name(), Path: filepath.Join(clean, item.Name()), IsDir: true})
	}
	sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name) })

	parent := filepath.Dir(clean)
	if parent == clean {
		parent = ""
	}
	return Listing{Path: clean, Parent: parent, Entries: entries}, nil
}

func browseRoots() (Listing, error) {
	roots := make([]Entry, 0, 8)
	for drive := 'A'; drive <= 'Z'; drive++ {
		path := string(drive) + `:\`
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			roots = append(roots, Entry{Name: path, Path: path, IsDir: true})
		}
	}
	if len(roots) == 0 {
		root := string(filepath.Separator)
		roots = append(roots, Entry{Name: root, Path: root, IsDir: true})
	}
	return Listing{Path: "", Parent: "", Entries: roots}, nil
}
