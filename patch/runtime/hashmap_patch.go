// +build go1.10 !go1.11

package runtime

import (
	"unsafe"
)

//go:linkname reflect_mapiterelem reflect.mapiterelem
func reflect_mapiterelem(it *hiter) unsafe.Pointer {
	return it.value
}
