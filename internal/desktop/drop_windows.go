//go:build windows

package desktop

import (
	"reflect"
	"sync"
	"syscall"
	"time"
	"unsafe"

	webview "github.com/abemedia/go-webview"
)

const (
	cfHDrop            = 15
	dvAspectContent    = 1
	tymedHGlobal       = 1
	dropEffectNone     = 0
	dropEffectCopy     = 1
	dragQueryFileCount = 0xFFFFFFFF
	wmDropFiles        = 0x0233
	wmCopyGlobalData   = 0x0049
	msgfltAllow        = 1
	gwlpWndProc        = ^uintptr(3) // -4
	webviewUIWidget    = 1
	sOK                = 0
	sFalse             = 1
	eNoInterface       = 0x80004002
	ePointer           = 0x80004003
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type formatEtc struct {
	CfFormat uint16
	Ptd      uintptr
	DwAspect uint32
	Lindex   int32
	Tymed    uint32
}

type stgMedium struct {
	Tymed          uint32
	Data           uintptr
	PUnkForRelease uintptr
}

type dropTargetVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	DragEnter      uintptr
	DragOver       uintptr
	DragLeave      uintptr
	Drop           uintptr
}

type dropTargetObject struct {
	VTable *dropTargetVTable
	Refs   uint32
}

var (
	user32DLL                       = syscall.NewLazyDLL("user32.dll")
	shell32DLL                      = syscall.NewLazyDLL("shell32.dll")
	ole32DLL                        = syscall.NewLazyDLL("ole32.dll")
	webviewDLL                      = syscall.NewLazyDLL("webview.dll")
	procEnumChildWindows            = user32DLL.NewProc("EnumChildWindows")
	procIsWindow                    = user32DLL.NewProc("IsWindow")
	procSetWindowLongPtrW           = user32DLL.NewProc("SetWindowLongPtrW")
	procCallWindowProcW             = user32DLL.NewProc("CallWindowProcW")
	procDefWindowProcW              = user32DLL.NewProc("DefWindowProcW")
	procChangeWindowMessageFilterEx = user32DLL.NewProc("ChangeWindowMessageFilterEx")
	procDragAcceptFiles             = shell32DLL.NewProc("DragAcceptFiles")
	procDragQueryFileW              = shell32DLL.NewProc("DragQueryFileW")
	procDragFinish                  = shell32DLL.NewProc("DragFinish")
	procOleInitialize               = ole32DLL.NewProc("OleInitialize")
	procOleUninitialize             = ole32DLL.NewProc("OleUninitialize")
	procRegisterDragDrop            = ole32DLL.NewProc("RegisterDragDrop")
	procRevokeDragDrop              = ole32DLL.NewProc("RevokeDragDrop")
	procReleaseStgMedium            = ole32DLL.NewProc("ReleaseStgMedium")
	procWebviewGetNativeHandle      = webviewDLL.NewProc("webview_get_native_handle")

	iidIUnknown   = guid{Data1: 0x00000000, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidDropTarget = guid{Data1: 0x00000122, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	dropTargetTable dropTargetVTable
	dropTarget      dropTargetObject

	dropMu             sync.Mutex
	dropHandler        func([]string)
	registeredOLE      = map[uintptr]bool{}
	classicDropHooks   = map[uintptr]uintptr{}
	enumDropWndProc    = syscall.NewCallback(enumDropChild)
	classicDropWndProc = syscall.NewCallback(classicFileDropWndProc)
	lastDropSignature  string
	lastDropAt         time.Time
)

func init() {
	dropTargetTable = dropTargetVTable{
		QueryInterface: syscall.NewCallback(dropQueryInterface),
		AddRef:         syscall.NewCallback(dropAddRef),
		Release:        syscall.NewCallback(dropRelease),
		DragEnter:      syscall.NewCallback(dropDragEnter),
		DragOver:       syscall.NewCallback(dropDragOver),
		DragLeave:      syscall.NewCallback(dropDragLeave),
		Drop:           syscall.NewCallback(dropDrop),
	}
	dropTarget = dropTargetObject{VTable: &dropTargetTable, Refs: 1}
}

func installFileDrop(window webview.WebView, handler func([]string)) func() {
	dropMu.Lock()
	dropHandler = handler
	dropMu.Unlock()

	oleResult, _, _ := procOleInitialize.Call(0)
	oleInitialized := uint32(oleResult) == sOK || uint32(oleResult) == sFalse

	root := uintptr(window.Window())
	installWebViewDropTargets(window, root)

	timers := make([]*time.Timer, 0, 4)
	for _, delay := range []time.Duration{150 * time.Millisecond, 500 * time.Millisecond, 1500 * time.Millisecond, 4 * time.Second} {
		timer := time.AfterFunc(delay, func() {
			window.Dispatch(func() { installWebViewDropTargets(window, root) })
		})
		timers = append(timers, timer)
	}

	return func() {
		for _, timer := range timers {
			timer.Stop()
		}

		dropMu.Lock()
		for hwnd := range registeredOLE {
			if ok, _, _ := procIsWindow.Call(hwnd); ok != 0 {
				procRevokeDragDrop.Call(hwnd)
			}
			delete(registeredOLE, hwnd)
		}
		for hwnd, original := range classicDropHooks {
			if ok, _, _ := procIsWindow.Call(hwnd); ok != 0 {
				procDragAcceptFiles.Call(hwnd, 0)
				procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, original)
			}
			delete(classicDropHooks, hwnd)
		}
		dropHandler = nil
		dropMu.Unlock()

		if oleInitialized {
			procOleUninitialize.Call()
		}
	}
}

func installWebViewDropTargets(window webview.WebView, root uintptr) {
	if root != 0 {
		installDropTargetForWindow(root)
		procEnumChildWindows.Call(root, enumDropWndProc, 0)
	}
	if widget := webviewWidgetHandle(window); widget != 0 {
		installDropTargetForWindow(widget)
	}
}

func webviewWidgetHandle(window webview.WebView) uintptr {
	value := reflect.ValueOf(window)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return 0
	}
	// github.com/abemedia/go-webview ma w przypiętej wersji pojedyncze pole
	// `handle uintptr`. Odczytujemy uchwyt natywnej instancji i pytamy bibliotekę
	// webview bezpośrednio o HWND widgetu WebView2, zamiast zgadywać po drzewie okien.
	type webviewMirror struct {
		Handle uintptr
	}
	instance := (*webviewMirror)(unsafe.Pointer(value.Pointer()))
	if instance.Handle == 0 {
		return 0
	}
	widget, _, _ := procWebviewGetNativeHandle.Call(instance.Handle, webviewUIWidget)
	return widget
}

func enumDropChild(hwnd, _ uintptr) uintptr {
	installDropTargetForWindow(hwnd)
	return 1
}

func installDropTargetForWindow(hwnd uintptr) {
	if hwnd == 0 {
		return
	}

	// WebView2 rejestruje własny OLE drop target. Zdejmujemy go i rejestrujemy
	// nasz IDropTarget, który odbiera CF_HDROP z Eksploratora i zachowuje pełne
	// ścieżki plików oraz folderów.
	procRevokeDragDrop.Call(hwnd)
	result, _, _ := procRegisterDragDrop.Call(hwnd, uintptr(unsafe.Pointer(&dropTarget)))
	if uint32(result) == sOK {
		dropMu.Lock()
		registeredOLE[hwnd] = true
		dropMu.Unlock()
	}

	// Fallback dla klasycznego WM_DROPFILES (oraz dla części konfiguracji z UIPI).
	procDragAcceptFiles.Call(hwnd, 1)
	procChangeWindowMessageFilterEx.Call(hwnd, wmDropFiles, msgfltAllow, 0)
	procChangeWindowMessageFilterEx.Call(hwnd, wmCopyGlobalData, msgfltAllow, 0)

	dropMu.Lock()
	_, alreadyHooked := classicDropHooks[hwnd]
	dropMu.Unlock()
	if alreadyHooked {
		return
	}
	original, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, classicDropWndProc)
	if original == 0 {
		return
	}
	dropMu.Lock()
	classicDropHooks[hwnd] = original
	dropMu.Unlock()
}

func deliverDroppedPaths(paths []string) {
	if len(paths) == 0 {
		return
	}
	signature := ""
	for _, path := range paths {
		signature += "\x00" + path
	}

	dropMu.Lock()
	if signature == lastDropSignature && time.Since(lastDropAt) < 350*time.Millisecond {
		dropMu.Unlock()
		return
	}
	lastDropSignature = signature
	lastDropAt = time.Now()
	handler := dropHandler
	dropMu.Unlock()

	if handler != nil {
		handler(paths)
	}
}
