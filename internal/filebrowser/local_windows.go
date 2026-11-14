//go:build windows

package filebrowser

import (
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"lantransfer/internal/i18n"
)

const (
	ofnExplorer         = 0x00080000
	ofnAllowMultiSelect = 0x00000200
	ofnFileMustExist    = 0x00001000
	ofnPathMustExist    = 0x00000800
	ofnHideReadOnly     = 0x00000004

	bifReturnOnlyFSDirs = 0x0001
	bifEditBox          = 0x0010
	bifNewDialogStyle   = 0x0040
)

var (
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")

	shell32                  = syscall.NewLazyDLL("shell32.dll")
	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")

	ole32             = syscall.NewLazyDLL("ole32.dll")
	procCoTaskMemFree = ole32.NewProc("CoTaskMemFree")
)

type openFileName struct {
	StructSize      uint32
	Owner           uintptr
	Instance        uintptr
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustomData      uintptr
	Hook            uintptr
	TemplateName    *uint16
	Reserved        uintptr
	Reserved2       uint32
	FlagsEx         uint32
}

type browseInfo struct {
	Owner       uintptr
	Root        uintptr
	DisplayName *uint16
	Title       *uint16
	Flags       uint32
	Callback    uintptr
	Param       uintptr
	Image       int32
}

func PickFiles() ([]string, error) {
	buffer := make([]uint16, 65536)
	lang := i18n.Current()
	title, _ := syscall.UTF16PtrFromString(i18n.T(lang, "dialog.pickFiles"))
	// Filter: description, NUL, "*.*", NUL, NUL.
	filter := append(utf16.Encode([]rune(i18n.T(lang, "dialog.allFiles"))), 0, '*', '.', '*', 0, 0)

	dialog := openFileName{
		StructSize:  uint32(unsafe.Sizeof(openFileName{})),
		Filter:      &filter[0],
		FilterIndex: 1,
		File:        &buffer[0],
		MaxFile:     uint32(len(buffer)),
		Title:       title,
		Flags:       ofnExplorer | ofnAllowMultiSelect | ofnFileMustExist | ofnPathMustExist | ofnHideReadOnly,
	}

	result, _, callErr := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&dialog)))
	if result == 0 {
		// Anulowanie dialogu też zwraca 0. Dla GUI traktujemy to jako brak wyboru.
		if callErr != syscall.Errno(0) {
			return nil, callErr
		}
		return []string{}, nil
	}

	parts := splitUTF16MultiString(buffer)
	if len(parts) == 0 {
		return []string{}, nil
	}
	if len(parts) == 1 {
		return []string{parts[0]}, nil
	}

	base := parts[0]
	paths := make([]string, 0, len(parts)-1)
	for _, name := range parts[1:] {
		paths = append(paths, filepath.Join(base, name))
	}
	return paths, nil
}

func PickFolder() (string, error) {
	title, _ := syscall.UTF16PtrFromString(i18n.T(i18n.Current(), "dialog.pickFolder"))
	displayName := make([]uint16, 260)
	info := browseInfo{
		DisplayName: &displayName[0],
		Title:       title,
		Flags:       bifReturnOnlyFSDirs | bifEditBox | bifNewDialogStyle,
	}

	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if pidl == 0 {
		return "", nil
	}
	defer procCoTaskMemFree.Call(pidl)

	pathBuffer := make([]uint16, 32768)
	ok, _, callErr := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&pathBuffer[0])))
	if ok == 0 {
		if callErr != syscall.Errno(0) {
			return "", callErr
		}
		return "", i18n.New("err.pickedFolderUnreadable")
	}
	return syscall.UTF16ToString(pathBuffer), nil
}

func splitUTF16MultiString(buffer []uint16) []string {
	parts := make([]string, 0, 4)
	start := 0
	for i, value := range buffer {
		if value != 0 {
			continue
		}
		if i == start {
			break
		}
		parts = append(parts, syscall.UTF16ToString(buffer[start:i]))
		start = i + 1
	}
	return parts
}
