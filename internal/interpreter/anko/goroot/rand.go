package goroot

import (
	"crypto/rand"
	"reflect"

	"project/external/anko/env"
)

func init() {
	initCryptoRand()
}

func initCryptoRand() {
	env.Packages["crypto/rand"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Reader": reflect.ValueOf(rand.Reader),

		// define functions
		"Int":   reflect.ValueOf(rand.Int),
		"Prime": reflect.ValueOf(rand.Prime),
		"Read":  reflect.ValueOf(rand.Read),
	}
	var ()
	env.PackageTypes["crypto/rand"] = map[string]reflect.Type{}
}
