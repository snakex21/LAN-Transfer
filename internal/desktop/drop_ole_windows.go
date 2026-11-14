//go:build windows

package desktop

// IDropTarget implemented in Go: receives CF_HDROP from Explorer with full
// paths of files and folders.

import (
	"sync/atomic"
	"syscall"
	"unsafe"
)

func dropQueryInterface(this, riid, ppv uintptr) uintptr {
	if ppv == 0 {
		return ePointer
	}
	*(*uintptr)(unsafe.Pointer(ppv)) = 0
	if riid == 0 {
		return eNoInterface
	}
	iid := *(*guid)(unsafe.Pointer(riid))
	if iid != iidIUnknown && iid != iidDropTarget {
		return eNoInterface
	}
	*(*uintptr)(unsafe.Pointer(ppv)) = this
	dropAddRef(this)
	return sOK
}

func dropAddRef(_ uintptr) uintptr {
	return uintptr(atomic.AddUint32(&dropTarget.Refs, 1))
}

func dropRelease(_ uintptr) uintptr {
	for {
		current := atomic.LoadUint32(&dropTarget.Refs)
		if current <= 1 {
			return uintptr(current)
		}
		if atomic.CompareAndSwapUint32(&dropTarget.Refs, current, current-1) {
			return uintptr(current - 1)
		}
	}
}

func dropDragEnter(_ uintptr, dataObject uintptr, _ uintptr, _ uintptr, effect uintptr) uintptr {
	setDropEffect(effect, hasFileDrop(dataObject))
	return sOK
}

func dropDragOver(_ uintptr, _ uintptr, _ uintptr, effect uintptr) uintptr {
	if effect != 0 {
		*(*uint32)(unsafe.Pointer(effect)) = dropEffectCopy
	}
	return sOK
}

func dropDragLeave(_ uintptr) uintptr {
	return sOK
}

func dropDrop(_ uintptr, dataObject uintptr, _ uintptr, _ uintptr, effect uintptr) uintptr {
	paths := pathsFromDataObject(dataObject)
	accepted := len(paths) > 0
	setDropEffect(effect, accepted)
	if accepted {
		deliverDroppedPaths(paths)
	}
	return sOK
}

func setDropEffect(effect uintptr, accepted bool) {
	if effect == 0 {
		return
	}
	if accepted {
		*(*uint32)(unsafe.Pointer(effect)) = dropEffectCopy
	} else {
		*(*uint32)(unsafe.Pointer(effect)) = dropEffectNone
	}
}

func hasFileDrop(dataObject uintptr) bool {
	if dataObject == 0 {
		return false
	}
	format := fileDropFormat()
	method := comMethod(dataObject, 5) // IDataObject::QueryGetData
	if method == 0 {
		return false
	}
	result, _, _ := syscall.SyscallN(method, dataObject, uintptr(unsafe.Pointer(&format)))
	return uint32(result) == sOK
}

func pathsFromDataObject(dataObject uintptr) []string {
	if dataObject == 0 {
		return nil
	}
	format := fileDropFormat()
	medium := stgMedium{}
	method := comMethod(dataObject, 3) // IDataObject::GetData
	if method == 0 {
		return nil
	}
	result, _, _ := syscall.SyscallN(method, dataObject, uintptr(unsafe.Pointer(&format)), uintptr(unsafe.Pointer(&medium)))
	if uint32(result) != sOK || medium.Tymed != tymedHGlobal || medium.Data == 0 {
		return nil
	}
	defer procReleaseStgMedium.Call(uintptr(unsafe.Pointer(&medium)))
	return droppedPaths(medium.Data)
}

func fileDropFormat() formatEtc {
	return formatEtc{
		CfFormat: cfHDrop,
		DwAspect: dvAspectContent,
		Lindex:   -1,
		Tymed:    tymedHGlobal,
	}
}

func comMethod(object uintptr, index uintptr) uintptr {
	if object == 0 {
		return 0
	}
	vtable := *(*uintptr)(unsafe.Pointer(object))
	if vtable == 0 {
		return 0
	}
	return *(*uintptr)(unsafe.Pointer(vtable + index*unsafe.Sizeof(uintptr(0))))
}
