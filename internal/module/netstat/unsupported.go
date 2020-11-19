// +build !windows

package netstat

import (
	"errors"
	"runtime"
)

// Options is a padding structure.
type Options struct{}

// NewNetstat is a padding function.
func NewNetstat(*Options) (Netstat, error) {
	return nil, errors.New("netstat is not implemented on " + runtime.GOOS)
}

// GetTCPConnState is is a padding function.
func GetTCPConnState(uint8) string {
	return "padding"
}
