// +build linux

package pkg

import (
	"reflect"
)

func f5() {
	reflect.ValueOf(nil)
}
