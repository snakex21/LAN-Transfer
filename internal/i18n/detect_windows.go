//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

var procGetUserPreferredUILanguages = syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserPreferredUILanguages")

// SystemLanguages returns the Windows display languages of the user, most
// preferred first ("de-DE", "en-US"), from GetUserPreferredUILanguages.
func SystemLanguages() []string {
	const muiLanguageName = 0x8
	var count, size uint32
	if r, _, _ := procGetUserPreferredUILanguages.Call(muiLanguageName, uintptr(unsafe.Pointer(&count)), 0, uintptr(unsafe.Pointer(&size))); r == 0 || size == 0 {
		return nil
	}
	buffer := make([]uint16, size)
	if r, _, _ := procGetUserPreferredUILanguages.Call(muiLanguageName, uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size))); r == 0 {
		return nil
	}
	var result []string
	start := 0
	for i, c := range buffer {
		if c != 0 {
			continue
		}
		if i == start {
			break
		}
		result = append(result, syscall.UTF16ToString(buffer[start:i]))
		start = i + 1
	}
	return result
}
