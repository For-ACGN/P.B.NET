// +build !go1.9

package vm

import (
	"fmt"
	"testing"
)

func TestMakeNotGo19(t *testing.T) {
	tests := []Test{
		{Script: `make(struct { a int64 })`, RunError: fmt.Errorf("reflect.StructOf: field \"a\" is unexported but has no PkgPath")},
	}
	runTests(t, tests, &Options{Debug: false})
}
