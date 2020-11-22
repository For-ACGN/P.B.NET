// +build !windows

package meterpreter

import (
	"errors"
	"net"
	"runtime"
)

func reverseTCP(*net.TCPConn, []byte, string) error {
	return errors.New("meterpreter is not implemented on " + runtime.GOOS)
}
