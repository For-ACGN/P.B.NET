// +build windows

package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testProcAttachConsole = modKernel32.NewProc("AttachConsole")
)

func testAllocConsole(t *testing.T, pid int) {
	ret, _, err := testProcAttachConsole.Call(uintptr(uint32(pid)))
	if ret == 0 {
		t.Fatal(err)
	}
}

func TestClear(t *testing.T) {
	testPrintTestLine()

	handle := os.Stdout.Fd()
	if !IsTerminal(handle) {

	}

	err := Clear(handle)
	require.NoError(t, err)
}
