package process

import (
	"encoding/binary"
	"time"
	"unsafe"
)

// TaskList is used to get current process list.
type TaskList interface {
	GetProcesses() ([]*Process, error)
	Close() error
}

// Process contains information about process.
type Process struct {
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
func (p *Process) ID() string {
	id := make([]byte, 16) // PID + timestamp
	binary.BigEndian.PutUint64(id, uint64(p.PID))
	binary.BigEndian.PutUint64(id[8:], uint64(p.CreationDate.UnixNano()))
	return *(*string)(unsafe.Pointer(&id)) // #nosec
}

// for compare package
type processes []*Process

func (ps processes) Len() int {
	return len(ps)
}

func (ps processes) ID(i int) string {
	return ps[i].ID()
}
