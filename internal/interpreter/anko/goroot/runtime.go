package goroot

import (
	"reflect"
	"runtime"

	"project/external/anko/env"
)

func init() {
	initRuntime()
}

func initRuntime() {
	env.Packages["runtime"] = map[string]reflect.Value{
		// define constants
		"GOOS":   reflect.ValueOf(runtime.GOOS),
		"GOARCH": reflect.ValueOf(runtime.GOARCH),

		// define variables

		// define functions
		"GC":             reflect.ValueOf(runtime.GC),
		"GOMAXPROCS":     reflect.ValueOf(runtime.GOMAXPROCS),
		"GOROOT":         reflect.ValueOf(runtime.GOROOT),
		"LockOSThread":   reflect.ValueOf(runtime.LockOSThread),
		"UnlockOSThread": reflect.ValueOf(runtime.UnlockOSThread),
		"Version":        reflect.ValueOf(runtime.Version),
	}
	var ()
	env.PackageTypes["runtime"] = map[string]reflect.Type{}
}
