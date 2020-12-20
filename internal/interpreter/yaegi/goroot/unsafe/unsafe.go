package unsafe

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Symbols stores the map of unsafe package symbols.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/traefik/yaegi/stdlib/unsafe"] = map[string]reflect.Value{
		"Symbols": reflect.ValueOf(Symbols),
	}
	Symbols["github.com/traefik/yaegi"] = map[string]reflect.Value{
		"convert": reflect.ValueOf(convert),
	}
	// type definitions
	Symbols["unsafe"] = map[string]reflect.Value{
		"Pointer": reflect.ValueOf((*unsafe.Pointer)(nil)),
	}
	// add builtin functions to unsafe.
	Symbols["unsafe"]["Sizeof"] = reflect.ValueOf(sizeof)
	Symbols["unsafe"]["Alignof"] = reflect.ValueOf(alignof)
	Symbols["unsafe"]["Offsetof"] = reflect.ValueOf(offsetof)
}

// #nosec
func convert(from, to reflect.Type) func(src, dest reflect.Value) {
	switch {
	case to.Kind() == reflect.UnsafePointer && from.Kind() == reflect.Uintptr:
		return uintptrToUnsafePtr
	case to.Kind() == reflect.UnsafePointer:
		return func(src, dest reflect.Value) {
			dest.SetPointer(unsafe.Pointer(src.Pointer()))
		}
	case to.Kind() == reflect.Uintptr && from.Kind() == reflect.UnsafePointer:
		return func(src, dest reflect.Value) {
			ptr := src.Interface().(unsafe.Pointer)
			dest.Set(reflect.ValueOf(uintptr(ptr)))
		}
	case from.Kind() == reflect.UnsafePointer:
		return func(src, dest reflect.Value) {
			ptr := src.Interface().(unsafe.Pointer)
			v := reflect.NewAt(dest.Type().Elem(), ptr)
			dest.Set(v)
		}
	default:
		return nil
	}
}

//go:nocheckptr
// #nosec
func uintptrToUnsafePtr(src, dest reflect.Value) {
	dest.SetPointer(unsafe.Pointer(src.Interface().(uintptr))) //nolint:govet
}

func sizeof(i interface{}) uintptr {
	return reflect.ValueOf(i).Type().Size()
}

func alignof(i interface{}) uintptr {
	return uintptr(reflect.ValueOf(i).Type().Align())
}

// compiler will replace origin code "unsafe.Offsetof(T{}.A)" to "unsafe.Offsetof(T{}, "A")".
func offsetof(s interface{}, f string) uintptr {
	typ := reflect.ValueOf(s).Type()
	sf, ok := typ.FieldByName(f)
	if !ok {
		panic(fmt.Sprintf("structure %s not contain field: \"%s\"", typ, f))
	}
	return sf.Offset
}
