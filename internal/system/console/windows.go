// +build windows

package console

import (
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// reference:
// https://docs.microsoft.com/zh-cn/windows/console/allocconsole

var (
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procFillConsoleOutputCharacter = modKernel32.NewProc("FillConsoleOutputCharacterA")
	procFillConsoleOutputAttribute = modKernel32.NewProc("FillConsoleOutputAttribute")
)

func isTerminal(handle uintptr) bool {
	var mode uint32
	err := windows.GetConsoleMode(windows.Handle(handle), &mode)
	return err == nil
}

func clear(handle uintptr) error {
	hConsole := windows.Handle(handle)
	// get current buffer size
	var ci windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(hConsole, &ci)
	if err != nil {
		return errors.Wrap(err, "failed to get console screen buffer information")
	}
	// calculate length that need fill
	cp := ci.CursorPosition
	length := uintptr(int(cp.Y)*int(ci.Size.X) + int(cp.X))
	// fill the entire screen with blanks
	pos := uintptr(uint32(0))
	var charsWritten uint32
	ret, _, err := procFillConsoleOutputCharacter.Call(
		handle, uintptr(uint8(' ')), length, pos,
		uintptr(unsafe.Pointer(&charsWritten)),
	)
	if ret == 0 {
		return errors.Wrap(err, "failed to fill console output character")
	}
	// get the current text attribute
	err = windows.GetConsoleScreenBufferInfo(hConsole, &ci)
	if err != nil {
		return errors.Wrap(err, "failed to get console screen buffer information")
	}
	// set the buffer's attributes accordingly
	ret, _, err = procFillConsoleOutputAttribute.Call(
		handle, uintptr(ci.Attributes), length, pos,
		uintptr(unsafe.Pointer(&charsWritten)),
	)
	if ret == 0 {
		return errors.Wrap(err, "failed to fill console output attribute")
	}
	// set cursor at (0, 0)
	err = windows.SetConsoleCursorPosition(hConsole, windows.Coord{})
	if err != nil {
		return errors.Wrap(err, "failed to set console cursor position")
	}
	return nil
}
