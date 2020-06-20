// +build windows

package goroot

import (
	"reflect"

	"github.com/mattn/anko/env"
)

func init() {
	initSyscallWindows()
}

func initSyscallWindows() {
	env.Packages["syscall"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var ()
	env.PackageTypes["syscall"] = map[string]reflect.Type{}
}
