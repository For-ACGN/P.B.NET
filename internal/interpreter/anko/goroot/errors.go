package goroot

import (
	"errors"
	"reflect"

	"project/external/anko/env"
)

func init() {
	initErrors()
}

func initErrors() {
	env.Packages["errors"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"New": reflect.ValueOf(errors.New),
	}
	var ()
	env.PackageTypes["errors"] = map[string]reflect.Type{}
}
