// +build go1.10, !go1.12

package runtime

import (
	"unsafe"
)

//go:linkname reflect_mapiterelem reflect.mapiterelem
//
// From go1.12
func reflect_mapiterelem(it *hiter) unsafe.Pointer {
	return it.value
}
