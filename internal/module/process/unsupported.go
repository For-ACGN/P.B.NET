// +build !windows

package process

import (
	"errors"
	"runtime"
)

// Options is a padding structure.
type Options struct{}

// NewTaskList is a padding function.
func NewTaskList(opts *Options) (TaskList, error) {
	return nil, errors.New("tasklist is not implemented on " + runtime.GOOS)
}
