// +build go1.10 !go1.11

package runtime

import (
	"unsafe"
)

// reflectlite_resolveNameOff resolves a name offset from a base pointer.
//go:linkname reflectlite_resolveNameOff internal/reflectlite.resolveNameOff
func reflectlite_resolveNameOff(ptrInModule unsafe.Pointer, off int32) unsafe.Pointer {
	return unsafe.Pointer(resolveNameOff(ptrInModule, nameOff(off)).bytes)
}

// reflectlite_resolveTypeOff resolves an *rtype offset from a base type.
//go:linkname reflectlite_resolveTypeOff internal/reflectlite.resolveTypeOff
func reflectlite_resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer {
	return unsafe.Pointer((*_type)(rtype).typeOff(typeOff(off)))
}
