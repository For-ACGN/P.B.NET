package anko

import (
	"fmt"
	"reflect"

	"project/external/anko/env"
)

func defineCore(e *env.Env) {
	defineCoreType(e)
	defineCoreFunc(e)
}

func defineCoreType(e *env.Env) {
	for _, item := range [...]*struct {
		symbol string
		typ    interface{}
	}{
		{"int8", int8(1)},
		{"int16", int16(1)},
		{"uint8", uint8(1)},
		{"uint16", uint16(1)},
		{"uintptr", uintptr(1)},
	} {
		err := e.DefineType(item.symbol, item.typ)
		if err != nil {
			panic(fmt.Sprintf("anko: internal error: %s", err))
		}
	}
}

// defineCoreFunc is used to add core function.
// core.Import() with leaks, so we implement it self.
func defineCoreFunc(e *env.Env) {
	for _, item := range [...]*struct {
		symbol string
		fn     interface{}
	}{
		{"keys", coreKeys},
		{"range", coreRange},
		{"instance", coreInstance},
		{"arrayType", coreArrayType},
		{"array", coreArray},
		{"slice", coreSlice},
		{"typeOf", coreTypeOf},
		{"kindOf", coreKindOf},
	} {
		err := e.Define(item.symbol, item.fn)
		if err != nil {
			panic(fmt.Sprintf("anko: internal error: %s", err))
		}
	}
}

func coreKeys(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	keysValue := rv.MapKeys()
	keys := make([]interface{}, len(keysValue))
	for i := 0; i < len(keysValue); i++ {
		keys[i] = keysValue[i].Interface()
	}
	return keys
}

func coreRange(args ...int64) []int64 {
	var (
		start int64
		stop  int64
	)
	step := int64(1)
	switch len(args) {
	case 0:
		panic("range expected at least 1 argument, got 0")
	case 1:
		stop = args[0]
	case 2:
		start = args[0]
		stop = args[1]
	case 3:
		start = args[0]
		stop = args[1]
		step = args[2]
		if step == 0 {
			panic("range argument 3 must not be zero")
		}
	default:
		panic(fmt.Sprintf("range expected at most 3 arguments, got %d", len(args)))
	}
	var val []int64
	for i := start; (step > 0 && i < stop) || (step < 0 && i > stop); i += step {
		val = append(val, i)
	}
	return val
}

// coreInstance is used to new object with type.
// if is global type, use new or make. if type is created in script, use this.
func coreInstance(typ interface{}) interface{} {
	var reflectType reflect.Type
	var ok bool
	reflectType, ok = typ.(reflect.Type)
	if !ok {
		reflectType = reflect.TypeOf(typ)
	}
	return reflect.New(reflectType).Interface()
}

// coreArrayType is used to create a array type like [8]byte.
func coreArrayType(typ interface{}, size int) reflect.Type {
	return reflect.ArrayOf(size, reflect.TypeOf(typ))
}

// coreArray is used to create a array like [8]byte{}.
func coreArray(typ interface{}, size int) interface{} {
	return coreInstance(coreArrayType(typ, size))
}

// coreSlice is used to convert array to slice like [8]byte[:]
// must input address about array.
func coreSlice(ptr interface{}) interface{} {
	array := reflect.ValueOf(ptr).Elem()
	return array.Slice(0, array.Len()).Interface()
}

func coreTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}

func coreKindOf(v interface{}) string {
	typeOf := reflect.TypeOf(v)
	if typeOf == nil {
		return "nil kind"
	}
	return typeOf.Kind().String()
}
