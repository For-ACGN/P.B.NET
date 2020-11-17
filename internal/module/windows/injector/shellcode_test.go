// +build windows

package injector

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestMain(m *testing.M) {
	code := m.Run()

	_, _ = exec.Command("taskkill", "-im", "calculator.exe", "/F").CombinedOutput()
	_, _ = exec.Command("taskkill", "-im", "win32calc.exe", "/F").CombinedOutput()
	fmt.Println("clean calc processes.")

	os.Exit(code)
}

func testSelectShellcode(t *testing.T) []byte {
	var (
		file *os.File
		err  error
	)
	switch runtime.GOARCH {
	case "386":
		file, err = os.Open("../../shellcode/testdata/windows_32.txt")
		require.NoError(t, err)
	case "amd64":
		file, err = os.Open("../../shellcode/testdata/windows_64.txt")
		require.NoError(t, err)
	default:
		t.Skip("unsupported architecture:", runtime.GOARCH)
	}
	t.Logf("use %s shellcode\n", runtime.GOARCH)
	defer func() { _ = file.Close() }()
	shellcode, err := ioutil.ReadAll(hex.NewDecoder(file))
	require.NoError(t, err)
	return shellcode
}

func TestInjectShellcode(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	cmd := exec.Command("notepad.exe")
	err := cmd.Start()
	require.NoError(t, err)

	pid := uint32(cmd.Process.Pid)
	t.Log("notepad.exe process id:", pid)

	shellcode := testSelectShellcode(t)
	cp := make([]byte, len(shellcode))

	t.Run("wait and clean", func(t *testing.T) {
		copy(cp, shellcode)

		err = InjectShellcode(pid, cp, 0, false, true, true)
		require.NoError(t, err)
	})

	t.Run("bypass session isolation", func(t *testing.T) {
		copy(cp, shellcode)

		err = InjectShellcode(pid, cp, 0, true, true, true)
		require.NoError(t, err)
	})

	t.Run("wait", func(t *testing.T) {
		copy(cp, shellcode)

		err = InjectShellcode(pid, cp, 8, false, true, false)
		require.NoError(t, err)
	})

	t.Run("not wait", func(t *testing.T) {
		copy(cp, shellcode)

		err = InjectShellcode(pid, cp, 16, false, false, false)
		require.NoError(t, err)

		time.Sleep(3 * time.Second)
	})

	err = cmd.Process.Kill()
	require.NoError(t, err)

	// exit status 1
	err = cmd.Wait()
	require.Error(t, err)
}
