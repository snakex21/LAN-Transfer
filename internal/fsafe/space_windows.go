//go:build windows

package fsafe

import (
	"syscall"
	"unsafe"
)

var procGetDiskFreeSpaceExW = syscall.NewLazyDLL("kernel32.dll").NewProc("GetDiskFreeSpaceExW")

// FreeSpace returns the number of bytes available to the current user on the
// volume that holds dir.
func FreeSpace(dir string) (uint64, error) {
	ptr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return 0, err
	}
	var available uint64
	ok, _, callErr := procGetDiskFreeSpaceExW.Call(uintptr(unsafe.Pointer(ptr)), uintptr(unsafe.Pointer(&available)), 0, 0)
	if ok == 0 {
		return 0, callErr
	}
	return available, nil
}
