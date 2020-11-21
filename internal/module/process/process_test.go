package process

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestProcess(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	process, err := New(nil)
	require.NoError(t, err)

	processes, err := process.List()
	require.NoError(t, err)

	require.NotEmpty(t, processes)
	for _, process := range processes {
		fmt.Println(process.Name, process.Architecture, process.Username)
	}

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}

func TestPsInfo_ID(t *testing.T) {
	info := PsInfo{
		PID:  0x1234567887654321,
		PPID: 0x1234567812345678,
	}
	id := string([]byte{
		0x12, 0x34, 0x56, 0x78, 0x87, 0x65, 0x43, 0x21,
		0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78,
	})
	require.Equal(t, id, info.ID())
}

func TestPsInfo_Clone(t *testing.T) {
	info := &PsInfo{
		PID:  0x1234567887654321,
		PPID: 0x1234567812345678,
	}
	infoCp := info.Clone()
	require.Equal(t, info, infoCp)
}
