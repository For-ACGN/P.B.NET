// +build windows

package process

import (
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
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

	closeOnce sync.Once
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
	ps := process{
		opts:     opts,
		isARM:    strings.Contains(runtime.GOARCH, "arm"),
		majorVer: major,
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
	return &ps, nil
}

func (ps *process) GetProcesses() ([]*PsInfo, error) {
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
		if ps.opts.ShowThreadCount {
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
	if ps.opts.ShowArchitecture {
		process.Architecture = ps.getProcessArchitecture(pHandle)
	}
}

func (ps *process) openProcess(pid int64) (windows.Handle, error) {
	var da uint32
	if ps.major < 6 {
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
	ret, _, _ := ps.isWow64.Call(uintptr(handle), uintptr(unsafe.Pointer(&wow64))) // #nosec
	if ret == 0 {
		return ""
	}
	if wow64 {
		return "x86"
	}
	return "x64"
}

func (ps *process) Close() (err error) {
	if ps.modKernel32 == nil {
		return
	}
	ps.closeOnce.Do(func() {
		handle := windows.Handle(ps.modKernel32.Handle())
		err = windows.FreeLibrary(handle)
	})
	return
}
