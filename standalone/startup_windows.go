//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// showStartupError shows a blocking message box. The Windows build has no
// console (-H=windowsgui), so stderr output would be invisible.
func showStartupError(title, msg string) {
	const mbOK, mbIconError = 0x0, 0x10
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbOK|mbIconError)
}
