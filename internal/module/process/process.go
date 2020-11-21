package process

import (
	"encoding/binary"
	"os"
	"time"
	"unsafe"
)

// Process is a module about process.
type Process interface {
	// GetList is used to get process list.
	GetList() ([]*PsInfo, error)

	// Create is used to create process with options.
	Create(name string, opts *CreateOptions) (*os.Process, error)

	// Kill is used to kill process.
	Kill(pid int) error

	// KillTree is used to kill process tree.
	KillTree(pid int) error

	// SendSignal is used to send signal to process.
	SendSignal(pid int, signal os.Signal) error

	// Close is used to close module for release resource.
	Close() error
}

// PsInfo contains information about process.
type PsInfo struct {
	// these fields must not be zero value
	Name string
	PID  int64
	PPID int64

	SessionID uint32
	Username  string

	// for calculate CPU usage
	UserModeTime   uint64
	KernelModeTime uint64

	// for calculate Memory usage
	MemoryUsed uint64

	HandleCount uint32
	ThreadCount uint32

	IOReadBytes  uint64
	IOWriteBytes uint64

	Architecture   string
	CommandLine    string
	ExecutablePath string
	CreationDate   time.Time
}

// ID is used to identified this Process.
func (info *PsInfo) ID() string {
	id := make([]byte, 16)
	binary.BigEndian.PutUint64(id, uint64(info.PID))
	binary.BigEndian.PutUint64(id[8:], uint64(info.PPID))
	return *(*string)(unsafe.Pointer(&id)) // #nosec
}

// Clone is used to clone information about this process.
func (info *PsInfo) Clone() *PsInfo {
	i := *info
	return &i
}

// CreateOptions contain options about create process.
type CreateOptions struct {
	CommandLine   string
	Directory     string
	Environment   []string
	ShowWindow    bool
	CreationFlags uint32
}
