//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	findWindowW      = user32.NewProc("FindWindowW")
	sendMessageW     = user32.NewProc("SendMessageW")
	loadIconW        = user32.NewProc("LoadIconW")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	WM_SETICON = 0x0080
	ICON_SMALL = 0
	ICON_BIG   = 1
)

func SetPlatformIcon(windowTitle string) {
	titlePtr, _ := syscall.UTF16PtrFromString(windowTitle)
	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return
	}

	hInstance, _, _ := getModuleHandleW.Call(0)
	hIcon, _, _ := loadIconW.Call(hInstance, uintptr(1))
	if hIcon == 0 {
		return
	}

	sendMessageW.Call(hwnd, WM_SETICON, uintptr(ICON_SMALL), hIcon)
	sendMessageW.Call(hwnd, WM_SETICON, uintptr(ICON_BIG), hIcon)
}
