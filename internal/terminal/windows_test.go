// +build windows

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testProcAllocConsole  = modKernel32.NewProc("AllocConsole")
	testProcFreeConsole   = modKernel32.NewProc("FreeConsole")
	testProcAttachConsole = modKernel32.NewProc("AttachConsole")
)

func testFreeConsole(t *testing.T) {
	ret, _, err := testProcFreeConsole.Call()
	if ret == 0 {
		t.Fatal(err)
	}
}

func testAttachConsole(t *testing.T, pid int) {
	ret, _, err := testProcAttachConsole.Call(uintptr(uint32(pid)))
	if ret == 0 {
		t.Fatal(err)
	}
}

func testAllocConsole(t *testing.T) {
	ret, _, err := testProcAllocConsole.Call()
	if ret == 0 {
		t.Fatal(err)
	}
}

func TestClear(t *testing.T) {
	testPrintText()

	handle := os.Stdout.Fd()
	if !IsTerminal(handle) {
		// TODO [debt] need to find a way to test in Goland

		testFreeConsole(t)
		testAllocConsole(t)
		testFreeConsole(t)

		cmd := exec.Command("conhost.exe")

		// cmd.SysProcAttr = &syscall.SysProcAttr{
		// 	HideWindow:    false,
		// 	CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		// }
		err := cmd.Start()
		require.NoError(t, err)
		defer func() {
			err := cmd.Process.Kill()
			require.NoError(t, err)
		}()

		fmt.Println(cmd.Process.Pid)

		testAttachConsole(t, -1)
		return
	}

	err := Clear(handle)
	require.NoError(t, err)
}
