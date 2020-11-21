// +build windows

package process

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
	"project/internal/module/windows/privilege"
	"project/internal/patch/monkey"
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

	t.Run("common", func(t *testing.T) {
		ps, err := process.Create("notepad.exe", nil)
		require.NoError(t, err)

		err = ps.Kill()
		require.NoError(t, err)
	})

	t.Run("failed to start", func(t *testing.T) {
		ps, err := process.Create("foo.acg", nil)
		require.Error(t, err)
		require.Nil(t, ps)
	})

	t.Run("panic in wait", func(t *testing.T) {
		var cmd *exec.Cmd
		patch := func(interface{}) error {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(cmd, "Wait", patch)
		defer pg.Unpatch()

		ps, err := process.Create("notepad.exe", nil)
		require.NoError(t, err)

		time.Sleep(time.Second)

		err = ps.Kill()
		require.NoError(t, err)
	})

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}

func TestProcess_Kill(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	process, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		ps, err := process.Create("notepad.exe", nil)
		require.NoError(t, err)

		err = process.Kill(ps.Pid)
		require.NoError(t, err)
	})

	t.Run("failed to find process", func(t *testing.T) {
		err = process.Kill(12345678)
		require.Error(t, err)
	})

	t.Run("failed to kill process", func(t *testing.T) {
		patch := func(uint32, bool, uint32) (windows.Handle, error) {
			return 0, nil
		}
		pg := monkey.Patch(api.OpenProcess, patch)
		defer pg.Unpatch()

		err = process.Kill(0)
		require.Error(t, err)
	})

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}

func TestProcess_KillTree(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	process, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		cmd := exec.Command("cmd.exe")
		r, w, err := os.Pipe()
		require.NoError(t, err)
		cmd.Stdin = r

		err = cmd.Start()
		require.NoError(t, err)
		_, err = w.WriteString("start\n")
		require.NoError(t, err)
		// wait start sub process
		time.Sleep(time.Second)

		var pg *monkey.PatchGuard
		patch := func(p Process, pid int) error {
			pg.Unpatch()
			defer pg.Restore()
			// check process name is cmd.exe
			// because kill conhost.exe maybe failed
			name, err := api.GetProcessNameByPID(uint32(pid))
			require.NoError(t, err)
			if name == "cmd.exe" {
				err = p.Kill(pid)
				require.NoError(t, err)
			}
			if pid != cmd.Process.Pid {
				return monkey.Error
			}
			return nil
		}
		pg = monkey.PatchInstanceMethod(process, "Kill", patch)
		defer pg.Unpatch()

		err = process.KillTree(cmd.Process.Pid)
		require.NoError(t, err)

		// exit status 1
		err = cmd.Wait()
		require.Error(t, err)

		err = w.Close()
		require.NoError(t, err)
	})

	t.Run("failed to get process list", func(t *testing.T) {
		patch := func() ([]*api.ProcessBasicInfo, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(api.GetProcessList, patch)
		defer pg.Unpatch()

		err = process.KillTree(0)
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("failed to kill sub process", func(t *testing.T) {
		cmd := exec.Command("cmd.exe")
		r, w, err := os.Pipe()
		require.NoError(t, err)
		cmd.Stdin = r

		err = cmd.Start()
		require.NoError(t, err)
		_, err = w.WriteString("start\n")
		require.NoError(t, err)
		// wait start sub process
		time.Sleep(time.Second)

		var pg *monkey.PatchGuard
		patch := func(p Process, pid int) error {
			pg.Unpatch()
			defer pg.Restore()
			// check process name is cmd.exe
			// because kill conhost.exe maybe failed
			name, err := api.GetProcessNameByPID(uint32(pid))
			require.NoError(t, err)
			if name == "cmd.exe" {
				err = p.Kill(pid)
				require.NoError(t, err)
			}
			if pid != cmd.Process.Pid {
				return monkey.Error
			}
			return nil
		}
		pg = monkey.PatchInstanceMethod(process, "Kill", patch)
		defer pg.Unpatch()

		err = process.KillTree(cmd.Process.Pid)
		monkey.IsExistMonkeyError(t, err)

		// exit status 1
		err = cmd.Wait()
		require.Error(t, err)

		err = w.Close()
		require.NoError(t, err)
	})

	err = process.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, process)
}

func TestProcess_SendSignal(t *testing.T) {

}

func TestProcess_Close(t *testing.T) {

}
