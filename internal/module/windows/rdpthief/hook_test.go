// +build windows

package rdpthief

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const (
	testCredHostname = "hostname"
	testCredUsername = "username"
	testCredPassword = "test"
)

func testCreateCredential(t *testing.T) {
	hostname := windows.StringToUTF16Ptr(testCredHostname)
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

	usernamePtr := windows.StringToUTF16Ptr(testCredUsername)
	proc = windows.NewLazySystemDLL("advapi32.dll").NewProc("CredIsMarshaledCredentialW")
	ret, _, _ = proc.Call(uintptr(unsafe.Pointer(usernamePtr)))
	require.Equal(t, uintptr(0), ret)
}

func TestHook(t *testing.T) {
	var hooked bool
	hook, err := NewHook(func(cred *Credential) {
		require.Equal(t, testCredHostname, cred.Hostname)
		require.Equal(t, testCredUsername, cred.Username)
		require.Equal(t, testCredPassword, cred.Password)
		hooked = true
	})
	require.NoError(t, err)

	err = hook.Install()
	require.NoError(t, err)

	testCreateCredential(t)
	require.True(t, hooked)

	// not record the same credential
	hooked = false
	testCreateCredential(t)
	require.False(t, hooked)

	// install && uninstall
	hook.lastCredHash = [32]byte{}

	err = hook.Uninstall()
	require.NoError(t, err)

	hooked = false
	testCreateCredential(t)
	require.False(t, hooked)

	err = hook.Install()
	require.NoError(t, err)

	testCreateCredential(t)
	require.True(t, hooked)

	err = hook.Uninstall()
	require.NoError(t, err)
}
