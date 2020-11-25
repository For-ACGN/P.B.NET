package psmon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/module/process"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestMonitor(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	handler := func(_ context.Context, event uint8, data interface{}) {
		switch event {
		case EventProcessCreated:
			testMonitorPrintCreatedProcesses(data.([]*process.PsInfo))
		case EventProcessTerminated:
			testMonitorPrintTerminatedProcesses(data.([]*process.PsInfo))
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)
	monitor.Start()

	monitor.SetInterval(50 * time.Millisecond)

	err = monitor.SetOptions(nil)
	require.NoError(t, err)

	monitor.Pause()
	time.Sleep(3 * time.Second)
	monitor.Continue()

	time.Sleep(3 * time.Second)

	monitor.GetProcesses()

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)
}

func testMonitorPrintCreatedProcesses(processes []*process.PsInfo) {
	for _, ps := range processes {
		fmt.Printf("create process PID: %d Name: %s\n", ps.PID, ps.Name)
	}
}

func testMonitorPrintTerminatedProcesses(processes []*process.PsInfo) {
	for _, ps := range processes {
		fmt.Printf("terminate process PID: %d Name: %s\n", ps.PID, ps.Name)
	}
}

func TestMonitor_EventProcessCreated(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	path, err := os.Executable()
	require.NoError(t, err)
	name := filepath.Base(path)

	var created bool

	handler := func(_ context.Context, event uint8, data interface{}) {
		if event != EventProcessCreated {
			return
		}
		for _, ps := range data.([]*process.PsInfo) {
			if ps.Name == name {
				created = true
			}
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)
	monitor.Start()

	// wait first auto refresh
	time.Sleep(2 * defaultRefreshInterval)

	// create process
	cmd := exec.Command(path)
	err = cmd.Start()
	require.NoError(t, err)

	// wait refresh
	time.Sleep(2 * defaultRefreshInterval)

	// terminate process
	err = cmd.Process.Kill()
	require.NoError(t, err)
	err = cmd.Process.Release()
	require.NoError(t, err)

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)

	require.True(t, created, "not find expected created process")
}

func TestMonitor_EventProcessTerminated(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	path, err := os.Executable()
	require.NoError(t, err)
	name := filepath.Base(path)

	// create process
	cmd := exec.Command(path)
	err = cmd.Start()
	require.NoError(t, err)

	var terminated bool

	handler := func(_ context.Context, event uint8, data interface{}) {
		if event != EventProcessTerminated {
			return
		}
		for _, ps := range data.([]*process.PsInfo) {
			if ps.Name == name {
				terminated = true
			}
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)
	monitor.Start()

	// wait first auto refresh
	time.Sleep(2 * defaultRefreshInterval)

	// terminate process
	err = cmd.Process.Kill()
	require.NoError(t, err)
	err = cmd.Process.Release()
	require.NoError(t, err)

	// wait refresh
	time.Sleep(2 * defaultRefreshInterval)

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)

	require.True(t, terminated, "not find expected terminated process")
}

func TestMonitor_refreshLoop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to refresh", func(t *testing.T) {
		monitor, err := New(logger.Test, nil, nil)
		require.NoError(t, err)

		var m *Monitor
		patch := func(interface{}) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(m, "Refresh", patch)
		defer pg.Unpatch()

		monitor.Start()

		// wait restart
		time.Sleep(3 * time.Second)

		err = monitor.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, monitor)
	})

	t.Run("panic", func(t *testing.T) {
		monitor, err := New(logger.Test, nil, nil)
		require.NoError(t, err)

		var m *Monitor
		patch := func(interface{}) error {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(m, "Refresh", patch)
		defer pg.Unpatch()

		monitor.Start()

		// wait restart
		time.Sleep(3 * time.Second)

		err = monitor.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, monitor)
	})
}
