// +build windows

package process

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/module/windows/privilege"
	"project/internal/testsuite"
)

func TestTaskList(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	err := privilege.EnableDebug()
	require.NoError(t, err)

	tasklist, err := NewTaskList(nil)
	require.NoError(t, err)

	processes, err := tasklist.GetProcesses()
	require.NoError(t, err)

	require.NotEmpty(t, processes)
	for _, process := range processes {
		fmt.Println(process.Name, process.Architecture, process.Username)
	}

	err = tasklist.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tasklist)
}
