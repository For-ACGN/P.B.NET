// +build windows

package process

import (
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
)

// Options is contains options about tasklist.
type Options struct {
	ShowSessionID bool
	ShowUsername  bool

	ShowUserModeTime   bool
	ShowKernelModeTime bool
	ShowMemoryUsed     bool

	ShowHandleCount bool
	ShowThreadCount bool

	ShowIOReadBytes  bool
	ShowIOWriteBytes bool

	ShowArchitecture   bool
	ShowCommandLine    bool
	ShowExecutablePath bool
	ShowCreationDate   bool
}

type taskList struct {
	opts *Options

	major       uint32
	modKernel32 *windows.LazyDLL
	isWow64     *windows.LazyProc

	closeOnce sync.Once
}

// NewTaskList is used to create a new TaskList tool.
func NewTaskList(opts *Options) (TaskList, error) {
	if opts == nil {
		opts = &Options{
			ShowSessionID:      true,
			ShowUsername:       true,
			ShowUserModeTime:   true,
			ShowKernelModeTime: true,
			ShowMemoryUsed:     true,
			ShowArchitecture:   true,
			ShowCommandLine:    true,
			ShowExecutablePath: true,
			ShowCreationDate:   true,
		}
	}
	major, _, _ := api.GetVersionNumber()
	tl := taskList{
		opts:  opts,
		major: major,
	}
	if api.IsSystem64Bit(true) {
		modKernel32 := windows.NewLazySystemDLL("kernel32.dll")
		proc := modKernel32.NewProc("IsWow64Process")
		err := proc.Find()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tl.modKernel32 = modKernel32
		tl.isWow64 = proc
	}
	return &tl, nil
}

func (tl *taskList) GetProcesses() ([]*Process, error) {
	list, err := api.GetProcessList()
	if err != nil {
		return nil, err
	}
	l := len(list)
	processes := make([]*Process, l)
	for i := 0; i < l; i++ {
		processes[i] = &Process{
			Name: list[i].Name,
			PID:  int64(list[i].PID),
			PPID: int64(list[i].PPID),
		}
		if tl.opts.ShowThreadCount {
			processes[i].ThreadCount = list[i].Threads
		}
		tl.getProcessInfo(processes[i])
	}
	return processes, nil
}

func (tl *taskList) getProcessInfo(process *Process) {
	pHandle, err := tl.openProcess(process.PID)
	if err != nil {
		return
	}
	defer api.CloseHandle(pHandle)
	if tl.opts.ShowUsername {
		process.Username = getProcessUsername(pHandle)
	}
	if tl.opts.ShowArchitecture {
		process.Architecture = tl.getProcessArchitecture(pHandle)
	}
}

func (tl *taskList) openProcess(pid int64) (windows.Handle, error) {
	var da uint32
	if tl.major < 6 {
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

func (tl *taskList) getProcessArchitecture(handle windows.Handle) string {
	if tl.isWow64 == nil {
		return "x86"
	}
	var wow64 bool
	ret, _, _ := tl.isWow64.Call(
		uintptr(handle), uintptr(unsafe.Pointer(&wow64)),
	) // #nosec
	if ret == 0 {
		return ""
	}
	if wow64 {
		return "x86"
	}
	return "x64"
}

func (tl *taskList) Close() (err error) {
	if tl.modKernel32 == nil {
		return
	}
	tl.closeOnce.Do(func() {
		handle := windows.Handle(tl.modKernel32.Handle())
		err = windows.FreeLibrary(handle)
	})
	return
}
