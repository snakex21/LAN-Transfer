//go:build windows

package desktop

// Classic WM_DROPFILES fallback (some UIPI configurations never deliver OLE
// drag and drop to the WebView window).

import (
	"syscall"
	"unsafe"
)

func classicFileDropWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == wmDropFiles {
		paths := droppedPaths(wParam)
		procDragFinish.Call(wParam)
		if len(paths) > 0 {
			deliverDroppedPaths(paths)
		}
		return 0
	}

	dropMu.Lock()
	original := classicDropHooks[hwnd]
	dropMu.Unlock()
	if original != 0 {
		result, _, _ := procCallWindowProcW.Call(original, hwnd, msg, wParam, lParam)
		return result
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return result
}

func droppedPaths(hdrop uintptr) []string {
	count, _, _ := procDragQueryFileW.Call(hdrop, dragQueryFileCount, 0, 0)
	paths := make([]string, 0, int(count))
	for i := uintptr(0); i < count; i++ {
		length, _, _ := procDragQueryFileW.Call(hdrop, i, 0, 0)
		if length == 0 {
			continue
		}
		buffer := make([]uint16, int(length)+1)
		procDragQueryFileW.Call(hdrop, i, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
		path := syscall.UTF16ToString(buffer)
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}
