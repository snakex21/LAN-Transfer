//go:build windows

package desktop

import (
	"sync"
	"syscall"
	"unsafe"

	webview "github.com/abemedia/go-webview"

	"lantransfer/internal/config"
)

const (
	swShowMaximized       = 3
	swShowMinimized       = 2
	wpfRestoreToMaximized = 0x0002
	wmSize                = 0x0005
	wmClose               = 0x0010
	sizeRestored          = 0
)

type windowPoint struct {
	X int32
	Y int32
}

type windowRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type windowPlacement struct {
	Length         uint32
	Flags          uint32
	ShowCmd        uint32
	MinPosition    windowPoint
	MaxPosition    windowPoint
	NormalPosition windowRect
}

var (
	windowStateUser32      = syscall.NewLazyDLL("user32.dll")
	procGetWindowPlacement = windowStateUser32.NewProc("GetWindowPlacement")
	procGetClientRect      = windowStateUser32.NewProc("GetClientRect")
	procGetDpiForWindow    = windowStateUser32.NewProc("GetDpiForWindow")
	procShowWindow         = windowStateUser32.NewProc("ShowWindow")
	procIsZoomed           = windowStateUser32.NewProc("IsZoomed")

	windowStateMu           sync.Mutex
	windowStateHWND         uintptr
	windowStateOriginal     uintptr
	windowStateNormalWidth  = config.DefaultWindowWidth
	windowStateNormalHeight = config.DefaultWindowHeight
	windowStateWndProc      = syscall.NewCallback(windowStateWindowProc)
)

func restoreWindowState(window webview.WebView) func() {
	state, _ := config.LoadWindowState()

	windowStateMu.Lock()
	windowStateNormalWidth = state.Width
	windowStateNormalHeight = state.Height
	windowStateMu.Unlock()

	// SetSize operuje na obszarze klienta WebView i dzięki temu nie kumuluje
	// rozmiaru ramek Windows przy każdym ponownym uruchomieniu programu.
	window.SetSize(state.Width, state.Height, webview.HintNone)
	if state.Maximized {
		procShowWindow.Call(uintptr(window.Window()), swShowMaximized)
	}
	return func() {}
}

func installWindowStateTracking(window webview.WebView) func() {
	hwnd := uintptr(window.Window())
	if hwnd == 0 {
		return func() {}
	}

	original, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, windowStateWndProc)
	if original == 0 {
		return func() {}
	}

	windowStateMu.Lock()
	windowStateHWND = hwnd
	windowStateOriginal = original
	windowStateMu.Unlock()

	return func() {
		windowStateMu.Lock()
		trackedHWND := windowStateHWND
		trackedOriginal := windowStateOriginal
		windowStateHWND = 0
		windowStateOriginal = 0
		windowStateMu.Unlock()

		if trackedHWND != 0 && trackedOriginal != 0 {
			if ok, _, _ := procIsWindow.Call(trackedHWND); ok != 0 {
				procSetWindowLongPtrW.Call(trackedHWND, gwlpWndProc, trackedOriginal)
			}
		}
	}
}

func windowStateWindowProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == wmSize && wParam == sizeRestored {
		rememberNormalClientSize(hwnd)
	}
	if msg == wmClose {
		saveWindowStateFromHWND(hwnd)
	}

	windowStateMu.Lock()
	original := windowStateOriginal
	windowStateMu.Unlock()
	if original != 0 {
		result, _, _ := procCallWindowProcW.Call(original, hwnd, msg, wParam, lParam)
		return result
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return result
}

func rememberNormalClientSize(hwnd uintptr) {
	width, height, ok := clientSizeDIP(hwnd)
	if !ok || width < 700 || height < 520 {
		return
	}
	windowStateMu.Lock()
	windowStateNormalWidth = width
	windowStateNormalHeight = height
	windowStateMu.Unlock()
}

func saveWindowStateFromHWND(hwnd uintptr) {
	if hwnd == 0 {
		return
	}

	placement := windowPlacement{Length: uint32(unsafe.Sizeof(windowPlacement{}))}
	placementOK, _, _ := procGetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&placement)))
	zoomed, _, _ := procIsZoomed.Call(hwnd)
	maximized := zoomed != 0
	minimized := false
	if placementOK != 0 {
		maximized = maximized || placement.ShowCmd == swShowMaximized ||
			(placement.ShowCmd == swShowMinimized && placement.Flags&wpfRestoreToMaximized != 0)
		minimized = placement.ShowCmd == swShowMinimized
	}

	windowStateMu.Lock()
	width := windowStateNormalWidth
	height := windowStateNormalHeight
	windowStateMu.Unlock()

	// Gdy okno jest zwykłe, zapisujemy dokładny obszar klienta. Przy maksymalizacji
	// albo minimalizacji zachowujemy ostatni normalny rozmiar, żeby po przywróceniu
	// wrócić do tego samego wymiaru sprzed maksymalizacji.
	if !maximized && !minimized {
		if currentWidth, currentHeight, ok := clientSizeDIP(hwnd); ok && currentWidth >= 700 && currentHeight >= 520 {
			width = currentWidth
			height = currentHeight
		}
	}

	_ = config.SaveWindowState(config.WindowState{
		Width:     width,
		Height:    height,
		Maximized: maximized,
	})
}

func clientSizeDIP(hwnd uintptr) (int, int, bool) {
	rect := windowRect{}
	ok, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ok == 0 {
		return 0, 0, false
	}

	widthPixels := int(rect.Right - rect.Left)
	heightPixels := int(rect.Bottom - rect.Top)
	dpi, _, _ := procGetDpiForWindow.Call(hwnd)
	if dpi == 0 {
		dpi = 96
	}

	widthDIP := (widthPixels*96 + int(dpi)/2) / int(dpi)
	heightDIP := (heightPixels*96 + int(dpi)/2) / int(dpi)
	return widthDIP, heightDIP, true
}
