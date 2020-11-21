// +build windows

package process

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/module/windows/privilege"
	"project/internal/testsuite"
)

func TestProcess_GetList(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	err := privilege.EnableDebug()
	require.NoError(t, err)

	process, err := New(nil)
	require.NoError(t, err)

	processes, err := process.GetList()
	require.NoError(t, err)

	require.NotEmpty(t, processes)
	for _, process := range processes {
		fmt.Println(process.Name, process.Architecture, process.Username)
	}

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}

func TestProcess_Create(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	process, err := New(nil)
	require.NoError(t, err)

	p, err := process.Create("notepad.exe", nil)
	require.NoError(t, err)

	err = p.Kill()
	require.NoError(t, err)

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}
