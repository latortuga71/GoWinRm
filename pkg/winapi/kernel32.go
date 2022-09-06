package winapi

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	pModWKernel32 = syscall.NewLazyDLL("kernel32.dll")
	pCreateEventW = pModWKernel32.NewProc("CreateEventW")
	pSetEvent     = pModWKernel32.NewProc("SetEvent")
)

func SetEvent(handle uintptr) error {
	res, _, err := pSetEvent.Call(handle)
	if res == 0 {
		return err
	}
	return nil
}

func CreateEventW(lpSecAttr *windows.SecurityAttributes, bManualReset uint32, bInitialState uint32, name string) (uintptr, error) {
	n, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	res, _, err := pCreateEventW.Call(uintptr(unsafe.Pointer(lpSecAttr)), uintptr(bManualReset), uintptr(bInitialState), uintptr(unsafe.Pointer(n)))
	if res == 0 {
		return 0, err
	}
	return res, nil
}
