// +build windows

package hook

import (
	"fmt"
	"reflect"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestHookMessageBoxTimeoutW(t *testing.T) {
	var pg *PatchGuard
	hookFn := func(hwnd windows.Handle, text, caption *uint16, uType uint, id uint32, timeout uint) int {
		// compare parameters
		originText := windows.UTF16PtrToString(text)
		originCaption := windows.UTF16PtrToString(caption)
		require.Equal(t, "text", originText)
		require.Equal(t, "caption", originCaption)
		require.Equal(t, uint(1), uType)
		require.Equal(t, uint32(0), id)
		require.Equal(t, uint(1000), timeout)

		// call original function
		hookedText := fmt.Sprintf("origin: %s, hooked!", originText)
		hookedCaption := fmt.Sprintf("origin: %s, hooked!", originCaption)
		hookedTextPtr, err := windows.UTF16PtrFromString(hookedText)
		require.NoError(t, err)
		hookedCaptionPtr, err := windows.UTF16PtrFromString(hookedCaption)
		require.NoError(t, err)
		ret, _, err := pg.Original.Call(
			uintptr(hwnd), uintptr(unsafe.Pointer(hookedTextPtr)),
			uintptr(unsafe.Pointer(hookedCaptionPtr)), uintptr(uType), 0, 1000,
		)
		require.Equal(t, uintptr(32000), ret)
		require.Equal(t, syscall.Errno(0x00), err)

		// return fake return value
		return 1234
	}
	var err error
	pg, err = NewInlineHookByName("user32.dll", "MessageBoxTimeoutW", true, hookFn)
	require.NoError(t, err)
	err = pg.Patch()
	require.NoError(t, err)

	textPtr, err := windows.UTF16PtrFromString("text")
	require.NoError(t, err)
	captionPtr, err := windows.UTF16PtrFromString("caption")
	require.NoError(t, err)

	proc := windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxTimeoutW")
	ret, _, err := proc.Call(
		0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(captionPtr)), 1, 0, 1000,
	)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x00), err)

	// after unpatch
	err = pg.UnPatch()
	require.NoError(t, err)

	ret, _, err = proc.Call(
		0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(captionPtr)), 1, 0, 1000,
	)
	require.Equal(t, uintptr(32000), ret)
	require.Equal(t, syscall.Errno(0x00), err)

	// restore patch
	err = pg.Patch()
	require.NoError(t, err)

	ret, _, err = proc.Call(
		0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(captionPtr)), 1, 0, 1000,
	)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x00), err)
}

func TestHookCryptProtectMemory(t *testing.T) {
	data := make([]byte, 16)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i)
	}

	var pg *PatchGuard
	hookFn := func(address uintptr, size, flags uint) uintptr {
		// compare parameters
		var d []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&d))
		sh.Data = address
		sh.Len = 16
		sh.Cap = 16
		require.Equal(t, data, d)
		require.Equal(t, uint(16), size)
		require.Equal(t, uint(1), flags)

		// call original function
		ret, _, err := pg.Original.Call(uintptr(unsafe.Pointer(&data[0])), 16, 1)
		require.Equal(t, uintptr(1), ret)
		require.Equal(t, syscall.Errno(0x00), err)

		// return fake return value
		return 1234
	}
	var err error
	pg, err = NewInlineHookByName("crypt32.dll", "CryptProtectMemory", true, hookFn)
	require.NoError(t, err)
	err = pg.Patch()
	require.NoError(t, err)

	proc := windows.NewLazySystemDLL("crypt32.dll").NewProc("CryptProtectMemory")
	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&data[0])), 16, 1)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x00), err)

	// after unpatch
	err = pg.UnPatch()
	require.NoError(t, err)

	ret, _, err = proc.Call(uintptr(unsafe.Pointer(&data[0])), 16, 1)
	require.Equal(t, uintptr(1), ret)
	require.Equal(t, syscall.Errno(0x00), err)

	// restore patch
	err = pg.Patch()
	require.NoError(t, err)

	ret, _, err = proc.Call(uintptr(unsafe.Pointer(&data[0])), 16, 1)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x00), err)
}

func TestHookCredReadW(t *testing.T) {
	var pg *PatchGuard
	hookFn := func(targetName *uint16, typ, flags uint, credential uintptr) uintptr {
		// compare parameters
		name := windows.UTF16PtrToString(targetName)
		require.Equal(t, "test", name)
		require.Equal(t, uint(16), typ)
		require.Equal(t, uint(1), flags)
		require.Equal(t, uintptr(0x1234), credential)

		// call original function
		ret, _, err := pg.Original.Call(uintptr(unsafe.Pointer(targetName)), 16, 1, 0x1234)
		require.Equal(t, uintptr(0), ret)
		require.Equal(t, syscall.Errno(0x57), err)

		// return fake return value
		return 1234
	}
	var err error
	pg, err = NewInlineHookByName("advapi32.dll", "CredReadW", true, hookFn)
	require.NoError(t, err)
	err = pg.Patch()
	require.NoError(t, err)

	targetName := windows.StringToUTF16Ptr("test")

	proc := windows.NewLazySystemDLL("advapi32.dll").NewProc("CredReadW")
	ret, _, err := proc.Call(uintptr(unsafe.Pointer(targetName)), 16, 1, 0x1234)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x57), err)

	// after unpatch
	err = pg.UnPatch()
	require.NoError(t, err)

	ret, _, err = proc.Call(uintptr(unsafe.Pointer(targetName)), 16, 1, 0x1234)
	require.Equal(t, uintptr(0), ret)
	require.Equal(t, syscall.Errno(0x57), err)

	// restore patch
	err = pg.Patch()
	require.NoError(t, err)

	ret, _, err = proc.Call(uintptr(unsafe.Pointer(targetName)), 16, 1, 0x1234)
	require.Equal(t, uintptr(1234), ret)
	require.Equal(t, syscall.Errno(0x57), err)
}
