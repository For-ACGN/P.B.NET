package goroot

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"project/external/anko/env"
)

func init() {
	initUnsafe()
}

func initUnsafe() {
	env.Packages["unsafe"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Convert":         reflect.ValueOf(convert),
		"ConvertWithType": reflect.ValueOf(convertWithType),
		"Sizeof":          reflect.ValueOf(sizeOf),
		"Alignof":         reflect.ValueOf(alignOf),
		"Offsetof":        reflect.ValueOf(offsetOf),
	}
	var (
		pointer unsafe.Pointer
	)
	env.PackageTypes["unsafe"] = map[string]reflect.Type{
		"Pointer": reflect.TypeOf(&pointer).Elem(),
	}
}

// convert is used to force convert like
// n := *(*uint32)(unsafe.Pointer(&Float32))
//
// you can use these code in anko script:
// val = 256
// dstType = make(struct {
//     A int32,
//     B int32
// })
// p = unsafe.Convert(&val, dstType)
//
// println(p.A, p.B)
//
// newVal must the same type with typ
// see more information in TestUnsafe()
func convert(pointer *interface{}, typ interface{}) interface{} {
	return convertWithType(pointer, reflect.TypeOf(typ))
}

//go:nocheckptr
func convertWithType(pointer *interface{}, typ reflect.Type) interface{} {
	address := reflect.ValueOf(pointer).Elem().InterfaceData()[1]
	ptr := reflect.NewAt(typ, unsafe.Pointer(address)).Interface() // #nosec
	runtime.KeepAlive(pointer)
	return ptr
}

func sizeOf(i interface{}) uintptr {
	return reflect.ValueOf(i).Type().Size()
}

func alignOf(i interface{}) uintptr {
	return uintptr(reflect.ValueOf(i).Type().Align())
}

func offsetOf(s interface{}, f string) uintptr {
	typ := reflect.ValueOf(s).Type()
	sf, ok := typ.FieldByName(f)
	if !ok {
		panic(fmt.Sprintf("structure %s not contain field: \"%s\"", typ, f))
	}
	return sf.Offset
}
