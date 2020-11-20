// +build !windows

package process

import (
	"errors"
	"runtime"
)

// Options is a padding structure.
type Options struct{}

// New is a padding function.
func New(opts *Options) (Process, error) {
	return nil, errors.New("process is not implemented on " + runtime.GOOS)
}
