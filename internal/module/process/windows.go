// +build windows

package process

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
	"project/internal/system"
	"project/internal/xpanic"
)

// Options is contains options about process module.
type Options struct {
	GetSessionID bool
	GetUsername  bool

	GetUserModeTime   bool
	GetKernelModeTime bool
	GetMemoryUsed     bool

	GetHandleCount bool
	GetThreadCount bool

	GetIOReadBytes  bool
	GetIOWriteBytes bool

	GetArchitecture   bool
	GetCommandLine    bool
	GetExecutablePath bool
	GetCreationDate   bool
}

type process struct {
	opts *Options

	majorVer    uint32
	isARM       bool
	modKernel32 *windows.LazyDLL
	procIsWow64 *windows.LazyProc

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

// New is used to create a process module.
func New(opts *Options) (Process, error) {
	if opts == nil {
		opts = &Options{
			GetSessionID:      true,
			GetUsername:       true,
			GetUserModeTime:   true,
			GetKernelModeTime: true,
			GetMemoryUsed:     true,
			GetHandleCount:    true,
			GetThreadCount:    true,
			GetIOReadBytes:    true,
			GetIOWriteBytes:   true,
			GetArchitecture:   true,
			GetCommandLine:    true,
			GetExecutablePath: true,
			GetCreationDate:   true,
		}
	}
	major, _, _ := api.GetVersionNumber()
	isARM := strings.Contains(runtime.GOARCH, "arm")
	ps := process{
		opts:     opts,
		majorVer: major,
		isARM:    isARM,
	}
	if api.IsSystem64Bit(true) {
		modKernel32 := windows.NewLazySystemDLL("kernel32.dll")
		proc := modKernel32.NewProc("IsWow64Process")
		err := proc.Find()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		ps.modKernel32 = modKernel32
		ps.procIsWow64 = proc
	}
	ps.ctx, ps.cancel = context.WithCancel(context.Background())
	return &ps, nil
}

func (ps *process) GetList() ([]*PsInfo, error) {
	list, err := api.GetProcessList()
	if err != nil {
		return nil, err
	}
	l := len(list)
	processes := make([]*PsInfo, l)
	for i := 0; i < l; i++ {
		processes[i] = &PsInfo{
			Name: list[i].Name,
			PID:  int64(list[i].PID),
			PPID: int64(list[i].PPID),
		}
		if ps.opts.GetThreadCount {
			processes[i].ThreadCount = list[i].Threads
		}
		ps.getInfo(processes[i])
	}
	return processes, nil
}

func (ps *process) getInfo(process *PsInfo) {
	pHandle, err := ps.open(process.PID)
	if err != nil {
		return
	}
	defer api.CloseHandle(pHandle)
	if ps.opts.GetUsername {
		process.Username = getUsername(pHandle)
	}
	if ps.opts.GetArchitecture {
		process.Architecture = ps.getArchitecture(pHandle)
	}
}

func (ps *process) open(pid int64) (windows.Handle, error) {
	var da uint32
	if ps.majorVer < 6 {
		da = windows.PROCESS_QUERY_INFORMATION
	} else {
		da = windows.PROCESS_QUERY_LIMITED_INFORMATION
	}
	return api.OpenProcess(da, false, uint32(pid))
}

func getUsername(handle windows.Handle) string {
	var token windows.Token
	err := windows.OpenProcessToken(handle, windows.TOKEN_QUERY, &token)
	if err != nil {
		return ""
	}
	tu, err := token.GetTokenUser()
	if err != nil {
		return ""
	}
	account, domain, _, err := tu.User.Sid.LookupAccount("")
	if err != nil {
		return ""
	}
	return domain + "\\" + account
}

func (ps *process) getArchitecture(handle windows.Handle) string {
	var arch string
	if ps.isARM {
		arch = "arm32"
	} else {
		arch = "x86"
	}
	if ps.procIsWow64 == nil {
		return arch
	}
	var isWow64 bool
	ret, _, _ := ps.procIsWow64.Call(
		uintptr(handle), uintptr(unsafe.Pointer(&isWow64)), // #nosec
	)
	if ret == 0 {
		return ""
	}
	if isWow64 {
		return arch
	}
	if ps.isARM {
		arch = "arm64"
	} else {
		arch = "x64"
	}
	return arch
}

func (ps *process) Create(name string, opts *CreateOptions) (*os.Process, error) {
	if opts == nil {
		opts = new(CreateOptions)
	}
	args := system.CommandLineToArgv(opts.CommandLine)
	cmd := exec.CommandContext(ps.ctx, name, args...)
	cmd.Dir = opts.Directory
	cmd.Env = opts.Environment
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    opts.HideWindow,
		CreationFlags: opts.CreationFlags,
	}
	err := cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create process")
	}
	ps.wg.Add(1)
	go func() {
		defer ps.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				xpanic.Log(r, "process.Create")
			}
		}()
		_ = cmd.Wait()
	}()
	return cmd.Process, nil
}

func (ps *process) Kill(pid int) error {
	hProcess, err := api.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return errors.WithMessagef(err, "failed to open process %d", pid)
	}
	defer api.CloseHandle(hProcess)
	err = api.TerminateProcess(hProcess, 1)
	if err != nil {
		return err
	}
	return nil
}

func (ps *process) KillTree(pid int) error {
	err := ps.Kill(pid)
	if err != nil {
		return err
	}
	list, err := api.GetProcessList()
	if err != nil {
		return err
	}
	for i := 0; i < len(list); i++ {
		if int(list[i].PPID) != pid {
			continue
		}
		e := ps.KillTree(int(list[i].PID))
		if e != nil && err == nil {
			err = e
		}
	}
	if err != nil {
		return errors.WithMessagef(err, "appear error when kill process tree %d", pid)
	}
	return nil
}

func (ps *process) SendSignal(pid int, signal os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "failed to find process %d", pid)
	}
	err = process.Signal(signal)
	if err != nil {
		return errors.Wrapf(err, "failed to send signal %d to process %d", signal, pid)
	}
	return nil
}

func (ps *process) Close() (err error) {
	ps.closeOnce.Do(func() {
		ps.cancel()
		ps.wg.Wait()
		if ps.modKernel32 == nil {
			return
		}
		handle := windows.Handle(ps.modKernel32.Handle())
		err = windows.FreeLibrary(handle)
	})
	return
}
