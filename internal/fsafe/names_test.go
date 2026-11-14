package fsafe

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCleanRelPathRejectsHostileNames(t *testing.T) {
	bad := []string{
		"", ".", "..", "../x", "a/../../x", "a/./b", `..\x`, `a\b`, "/etc/passwd", `C:\Windows\x`, "C:x",
		`\\server\share\x`, "//server/share", "plik.txt:ukryty", "plik.txt::$DATA", "CON", "con.txt", "NUL",
		"aux.tar.gz", "COM1", "lpt9.log", "CONIN$", "kropka.", "spacja ", "a/b./c", "tab\x01", "new\nline",
		"pytajnik?", "gwiazdka*", "a<b", "cudz\"ysłów", "pipe|x", strings.Repeat("x", 256), "a//b", "a/",
		strings.Repeat("d/", MaxDepth) + "f", "\xff\xfe",
	}
	for _, name := range bad {
		if _, err := CleanRelPath(name); err == nil {
			t.Errorf("CleanRelPath(%q) accepted a hostile name", name)
		}
	}
}

func TestCleanRelPathAcceptsNormalNames(t *testing.T) {
	good := map[string]string{
		"plik.txt": "plik.txt",
		"Folder ŻÓŁW 日本語/sub/zdjęcie.jpg": filepath.Join("Folder ŻÓŁW 日本語", "sub", "zdjęcie.jpg"),
		".gitignore":    ".gitignore",
		"console.log":   "console.log",
		"COM10.txt":     "COM10.txt",
		"a b  c.tar.gz": "a b  c.tar.gz",
	}
	for in, want := range good {
		got, err := CleanRelPath(in)
		if err != nil || got != want {
			t.Errorf("CleanRelPath(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

func TestUniquePathNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "raport.pdf")
	if got := UniquePath(p); got != p {
		t.Fatalf("free path changed: %s", got)
	}
	os.WriteFile(p, []byte("1"), 0o644)
	if got := UniquePath(p); got != filepath.Join(dir, "raport (1).pdf") {
		t.Fatalf("got %s", got)
	}
	os.WriteFile(filepath.Join(dir, "raport (1).pdf"), []byte("2"), 0o644)
	if got := UniquePath(p); got != filepath.Join(dir, "raport (2).pdf") {
		t.Fatalf("got %s", got)
	}
	os.Mkdir(filepath.Join(dir, ".env"), 0o755)
	if got := UniquePath(filepath.Join(dir, ".env")); got != filepath.Join(dir, ".env (1)") {
		t.Fatalf("got %s", got)
	}
}

func TestWithin(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	cases := map[string]bool{
		root:                                  true,
		filepath.Join(root, "a", "b"):         true,
		filepath.Join(root, "..", "x"):        false,
		filepath.Join(root+"x", "a"):          false,
		filepath.Join(root, "..root", "file"): true,
	}
	for target, want := range cases {
		if got := Within(root, target); got != want {
			t.Errorf("Within(%q) = %v, want %v", target, got, want)
		}
	}
	if runtime.GOOS == "windows" && Within(`C:\a`, `D:\a`) {
		t.Error("different volumes must not be Within")
	}
}

func TestFreeSpace(t *testing.T) {
	free, err := FreeSpace(t.TempDir())
	if err != nil || free == 0 {
		t.Fatalf("FreeSpace = %d, %v", free, err)
	}
}
