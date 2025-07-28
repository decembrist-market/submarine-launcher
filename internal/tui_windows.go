//go:build windows
// +build windows

package internal

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows API для включения поддержки цветов
var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
)

const (
	enableVirtualTerminalProcessing = 0x0004
	stdOutputHandle                 = ^uintptr(10) // -11 в правильном формате для Windows
)

// EnableWindowsColors включает поддержку ANSI цветов в Windows терминале
func EnableWindowsColors() error {
	handle, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if handle == 0 {
		return fmt.Errorf("не удалось получить handle консоли")
	}

	var mode uint32
	ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return fmt.Errorf("не удалось получить режим консоли")
	}

	mode |= enableVirtualTerminalProcessing
	ret, _, _ = procSetConsoleMode.Call(handle, uintptr(mode))
	if ret == 0 {
		return fmt.Errorf("не удалось установить режим консоли")
	}

	return nil
}
