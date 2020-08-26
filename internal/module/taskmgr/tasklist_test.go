package taskmgr

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestTaskList(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tasklist, err := NewTaskList()
	require.NoError(t, err)

	processes, err := tasklist.GetProcesses()
	require.NoError(t, err)

	require.NotEmpty(t, processes)
	for _, process := range processes {
		spew.Dump(process)
	}

	tasklist.Close()

	testsuite.IsDestroyed(t, tasklist)
}

func TestProcess_ID(t *testing.T) {
	process := Process{
		PID:          0x1234567887654321,
		CreationDate: time.Unix(123, 123),
	}
	id := string([]byte{
		0x12, 0x34, 0x56, 0x78, 0x87, 0x65, 0x43, 0x21,
		0x00, 0x00, 0x00, 0x1C, 0xA3, 0x5F, 0x0E, 0x7B,
	})
	require.Equal(t, id, process.ID())
}
