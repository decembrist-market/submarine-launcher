//go:build !windows
// +build !windows

package internal

// EnableWindowsColors для non-Windows систем (Linux, macOS, etc.)
// На этих системах ANSI цвета обычно поддерживаются по умолчанию
func EnableWindowsColors() error {
	// На Unix-системах цвета работают из коробки
	return nil
}
