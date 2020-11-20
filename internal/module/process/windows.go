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

	isARM       bool
	majorVer    uint32
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
		isARM:    isARM,
		majorVer: major,
	}
	if isARM {
		// TODO get IsWow64Process2
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
		ps.getProcessInfo(processes[i])
	}
	return processes, nil
}

func (ps *process) getProcessInfo(process *PsInfo) {
	pHandle, err := ps.openProcess(process.PID)
	if err != nil {
		return
	}
	defer api.CloseHandle(pHandle)
	if ps.opts.GetUsername {
		process.Username = getProcessUsername(pHandle)
	}
	if ps.opts.GetArchitecture {
		process.Architecture = ps.getProcessArchitecture(pHandle)
	}
}

func (ps *process) openProcess(pid int64) (windows.Handle, error) {
	var da uint32
	if ps.majorVer < 6 {
		da = windows.PROCESS_QUERY_INFORMATION
	} else {
		da = windows.PROCESS_QUERY_LIMITED_INFORMATION
	}
	return api.OpenProcess(da, false, uint32(pid))
}

func getProcessUsername(handle windows.Handle) string {
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

func (ps *process) getProcessArchitecture(handle windows.Handle) string {
	if ps.isARM {
		return ""
	}

	if ps.procIsWow64 == nil {
		return "x86"
	}
	var wow64 bool
	ret, _, _ := ps.procIsWow64.Call(uintptr(handle), uintptr(unsafe.Pointer(&wow64))) // #nosec
	if ret == 0 {
		return ""
	}
	if wow64 {
		return "x86"
	}
	return "x64"
}

func (ps *process) Create(name string, opts *CreateOptions) error {
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
		return errors.Wrap(err, "failed to create process")
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
	return nil
}

func (ps *process) Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "failed to find process %d", pid)
	}
	err = p.Kill()
	if err != nil {
		return errors.Wrapf(err, "failed to kill process %d", pid)
	}
	return nil
}

func (ps *process) KillTree(_ int) error {
	return errors.New("not implemented")
}

func (ps *process) SendSignal(pid int, signal os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "failed to find process %d", pid)
	}
	err = p.Signal(signal)
	if err != nil {
		return errors.Wrapf(err, "failed to send signal %d to process %d", signal, pid)
	}
	return nil
}

func (ps *process) Close() (err error) {
	ps.cancel()
	ps.wg.Wait()
	ps.closeOnce.Do(func() {
		if ps.modKernel32 == nil {
			handle := windows.Handle(ps.modKernel32.Handle())
			err = windows.FreeLibrary(handle)
		}
	})
	return
}
