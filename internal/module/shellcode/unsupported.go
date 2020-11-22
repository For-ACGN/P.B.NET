// +build !windows

package shellcode

import (
	"errors"
	"runtime"
)

// Execute is a padding function.
func Execute(string, []byte) error {
	return errors.New("shellcode is not implemented on " + runtime.GOOS)
}
