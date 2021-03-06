// +build windows

package shell

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Shell is used to run one command with system shell.
func Shell(ctx context.Context, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "cmd.exe", "/c", command) // #nosec
	attr := syscall.SysProcAttr{
		HideWindow: true,
	}
	cmd.SysProcAttr = &attr
	return cmd.CombinedOutput()
}

// https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
const createNewConsole = 0x00000010

func createCommand(path string, args []string) *exec.Cmd {
	if path == "" {
		path = "cmd.exe"
	}
	cmd := exec.Command(path, args...) // #nosec
	cmd.SysProcAttr = setSysProcAttr()
	return cmd
}

func setSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNewConsole,
	}
}

var (
	// consoleCtrlHandler is used to hook handle interrupt signal
	consoleCtrlHandler uintptr
	// signalCh is used to confirm that has receive the interrupt signal
	signalCh chan struct{}

	globalMu sync.Mutex
)

func init() {
	consoleCtrlHandler = syscall.NewCallback(handleConsoleCtrl)
	signalCh = make(chan struct{}, 1)
}

// always return true
func handleConsoleCtrl(_ uintptr) uintptr {
	signalCh <- struct{}{}
	return uintptr(1)
}

func sendInterruptSignal(cmd *exec.Cmd) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	var needAttach bool
	// Detach self console for attach the process console
	// https://docs.microsoft.com/en-us/windows/console/freeconsole
	freeConsole, err := dll.FindProc("FreeConsole")
	if err != nil {
		return err
	}
	r1, _, _ := freeConsole.Call()
	if r1 != 0 {
		needAttach = true
	}

	// Attach to the console of the process
	// https://docs.microsoft.com/en-us/windows/console/attachconsole
	attachConsole, err := dll.FindProc("AttachConsole")
	if err != nil {
		return err
	}
	pid := cmd.Process.Pid
	r1, _, err = attachConsole.Call(uintptr(pid))
	if r1 == 0 {
		return errors.Errorf("failed to call AttachConsole: %s", err)
	}
	// After call SetConsoleCtrlHandler, then call GenerateConsoleCtrlEvent
	// interrupt signal will not send to this process.
	// https://docs.microsoft.com/en-us/windows/console/setconsolectrlhandler
	setConsoleCtrlHandler, err := dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return err
	}
	r1, _, err = setConsoleCtrlHandler.Call(consoleCtrlHandler, uintptr(1))
	if r1 == 0 {
		return errors.Errorf("failed to call SetConsoleCtrlHandler: %s", err)
	}
	// Send the CTRL_C_EVENT to a console process group that shares
	// the console associated with the calling process.
	// https://docs.microsoft.com/en-us/windows/console/generateconsolectrlevent
	generateConsoleCtrlEvent, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r1, _, err = generateConsoleCtrlEvent.Call(syscall.CTRL_C_EVENT, uintptr(0))
	if r1 == 0 {
		return errors.Errorf("failed to call CTRL_C_EVENT: %s", err)
	}
	// wait receive the interrupt signal
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-signalCh:
	case <-timer.C:
		return errors.New("failed to receive the interrupt signal")
	}
	time.Sleep(3 * time.Second) // TODO think it
	// free attached console
	r1, _, err = freeConsole.Call()
	if r1 == 0 {
		return errors.Errorf("failed to call FreeConsole(self): %s", err)
	}
	if !needAttach {
		return nil
	}
	// attache self console
	ppid := os.Getppid()
	r1, _, err = attachConsole.Call(uintptr(ppid))
	if r1 == 0 {
		return errors.Errorf("failed to call AttachConsole(self): %s", err)
	}
	return nil
}
