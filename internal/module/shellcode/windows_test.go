// +build windows

package shellcode

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestMain(m *testing.M) {
	code := m.Run()

	_, _ = exec.Command("taskkill", "-im", "calculator.exe", "/F").CombinedOutput()
	_, _ = exec.Command("taskkill", "-im", "win32calc.exe", "/F").CombinedOutput()
	fmt.Println("clean calc processes")

	os.Exit(code)
}

func testSelectShellcode(t *testing.T) []byte {
	var path string
	switch runtime.GOARCH {
	case "386":
		path = "testdata/windows_32.txt"
	case "amd64":
		path = "testdata/windows_64.txt"
	default:
		t.Skip("unsupported architecture:", runtime.GOARCH)
	}
	t.Logf("use %s shellcode\n", runtime.GOARCH)
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()
	shellcode, err := ioutil.ReadAll(hex.NewDecoder(file))
	require.NoError(t, err)
	return shellcode
}

func TestVirtualProtect(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	shellcode := testSelectShellcode(t)

	for i := 0; i < 3; i++ {
		cp := make([]byte, len(shellcode))
		copy(cp, shellcode)
		err := VirtualProtect(cp)
		require.NoError(t, err)
	}

	// empty data
	err := VirtualProtect(nil)
	require.EqualError(t, err, "empty data")
}

func TestCreateThread(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	shellcode := testSelectShellcode(t)

	for i := 0; i < 3; i++ {
		cp := make([]byte, len(shellcode))
		copy(cp, shellcode)
		err := CreateThread(cp)
		require.NoError(t, err)
	}

	// empty data
	err := CreateThread(nil)
	require.EqualError(t, err, "empty data")
}

func TestExecute(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	shellcode := testSelectShellcode(t)
	cp := make([]byte, len(shellcode))

	// must copy, because shellcode will be clean
	copy(cp, shellcode)
	err := Execute(MethodVirtualProtect, cp)
	require.NoError(t, err)

	copy(cp, shellcode)
	err = Execute(MethodCreateThread, cp)
	require.NoError(t, err)

	err = Execute("foo method", shellcode)
	require.EqualError(t, err, "unknown method: foo method")
}
