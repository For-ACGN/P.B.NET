// +build windows

package hook

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
)

func readMemory(addr uintptr, size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := api.ReadProcessMemory(windows.CurrentProcess(), addr, &data[0], uintptr(size))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writeMemory(addr uintptr, data []byte) error {
	_, err := api.WriteProcessMemory(windows.CurrentProcess(), addr, data)
	return err
}

// memory is used to write data to memory with VirtualProtect.
type memory struct {
	Addr       uintptr
	Size       uintptr
	oldProtect *uint32
}

func newMemory(addr uintptr, size int) *memory {
	return &memory{
		Addr:       addr,
		Size:       uintptr(size),
		oldProtect: new(uint32),
	}
}

func (mem *memory) Read() ([]byte, error) {
	return readMemory(mem.Addr, int(mem.Size))
}

func (mem *memory) Write(data []byte) (err error) {
	if uintptr(len(data)) != mem.Size {
		return errors.Errorf("invalid data size %d:%d", mem.Size, len(data))
	}
	err = api.VirtualProtect(mem.Addr, mem.Size, windows.PAGE_READWRITE, mem.oldProtect)
	if err != nil {
		return
	}
	defer func() {
		e := api.VirtualProtect(mem.Addr, mem.Size, *mem.oldProtect, mem.oldProtect)
		if e != nil && err == nil {
			err = e
		}
	}()
	return writeMemory(mem.Addr, data)
}

func (mem *memory) String() string {
	return fmt.Sprintf("0x%X", mem.Addr)
}
