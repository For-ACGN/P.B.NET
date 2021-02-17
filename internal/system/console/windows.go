// +build windows

package console

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procFillConsoleOutputCharacter = modKernel32.NewProc("FillConsoleOutputCharacterA")
	procFillConsoleOutputAttribute = modKernel32.NewProc("FillConsoleOutputAttribute")
)

// IsTerminal is used to check is in terminal.
func IsTerminal(handle uintptr) bool {
	var mode uint32
	err := windows.GetConsoleMode(windows.Handle(handle), &mode)
	return err == nil
}

// Clear is used to clean console buffer.
func Clear(handle uintptr) error {
	hConsole := windows.Handle(handle)
	// get current buffer size
	var ci windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(hConsole, &ci)
	if err != nil {
		return err
	}
	bufferSize := uintptr(ci.Size.X) * uintptr(ci.Size.Y)
	// fill the entire screen with blanks
	pos := uintptr(uint32(0))
	var charsWritten uint32
	ret, _, err := procFillConsoleOutputCharacter.Call(
		handle, uintptr(uint8(' ')), bufferSize, pos,
		uintptr(unsafe.Pointer(&charsWritten)),
	)
	if ret == 0 {
		return err
	}
	// get the current text attribute
	err = windows.GetConsoleScreenBufferInfo(hConsole, &ci)
	if err != nil {
		return err
	}
	// set the buffer's attributes accordingly
	ret, _, err = procFillConsoleOutputAttribute.Call(
		handle, uintptr(ci.Attributes), bufferSize, pos,
		uintptr(unsafe.Pointer(&charsWritten)),
	)
	if ret == 0 {
		return err
	}
	// set cursor at (0, 0)
	return windows.SetConsoleCursorPosition(hConsole, windows.Coord{})
}
