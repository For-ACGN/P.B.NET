package anko

import (
	"fmt"
	"reflect"

	"project/external/anko/env"
)

var (
	typeUint      = reflect.TypeOf(uint(0))
	typeUint8     = reflect.TypeOf(uint8(0))
	typeUint16    = reflect.TypeOf(uint16(0))
	typeUint32    = reflect.TypeOf(uint32(0))
	typeUint64    = reflect.TypeOf(uint64(0))
	typeInt       = reflect.TypeOf(0)
	typeInt8      = reflect.TypeOf(int8(0))
	typeInt16     = reflect.TypeOf(int16(0))
	typeInt32     = reflect.TypeOf(int32(0))
	typeInt64     = reflect.TypeOf(int64(0))
	typeByte      = reflect.TypeOf(byte(0))
	typeRune      = reflect.TypeOf(rune(0))
	typeUintptr   = reflect.TypeOf(uintptr(0))
	typeFloat32   = reflect.TypeOf(float32(0))
	typeFloat64   = reflect.TypeOf(float64(0))
	typeString    = reflect.TypeOf("")
	typeByteSlice = reflect.TypeOf([]byte{})
	typeRuneSlice = reflect.TypeOf([]rune{})
)

func defineConvert(e *env.Env) {
	for _, item := range [...]*struct {
		symbol string
		fn     interface{}
	}{
		{"uint", convertToUint},
		{"uint8", convertToUint8},
		{"uint16", convertToUint16},
		{"uint32", convertToUint32},
		{"uint64", convertToUint64},
		{"int", convertToInt},
		{"int8", convertToInt8},
		{"int16", convertToInt16},
		{"int32", convertToInt32},
		{"int64", convertToInt64},
		{"byte", convertToByte},
		{"rune", convertToRune},
		{"uintptr", convertToUintptr},
		{"float32", convertToFloat32},
		{"float64", convertToFloat64},
		{"string", convertToString},
		{"byteSlice", convertToByteSlice},
		{"runeSlice", convertToRuneSlice},
	} {
		err := e.Define(item.symbol, item.fn)
		if err != nil {
			panic(fmt.Sprintf("anko: internal error: %s", err))
		}
	}
}

func convertToUint(v interface{}) uint {
	return reflect.ValueOf(v).Convert(typeUint).Interface().(uint)
}

func convertToUint8(v interface{}) uint8 {
	return reflect.ValueOf(v).Convert(typeUint8).Interface().(uint8)
}

func convertToUint16(v interface{}) uint16 {
	return reflect.ValueOf(v).Convert(typeUint16).Interface().(uint16)
}

func convertToUint32(v interface{}) uint32 {
	return reflect.ValueOf(v).Convert(typeUint32).Interface().(uint32)
}

func convertToUint64(v interface{}) uint64 {
	return reflect.ValueOf(v).Convert(typeUint64).Interface().(uint64)
}

func convertToInt(v interface{}) int {
	return reflect.ValueOf(v).Convert(typeInt).Interface().(int)
}

func convertToInt8(v interface{}) int8 {
	return reflect.ValueOf(v).Convert(typeInt8).Interface().(int8)
}

func convertToInt16(v interface{}) int16 {
	return reflect.ValueOf(v).Convert(typeInt16).Interface().(int16)
}

func convertToInt32(v interface{}) int32 {
	return reflect.ValueOf(v).Convert(typeInt32).Interface().(int32)
}

func convertToInt64(v interface{}) int64 {
	return reflect.ValueOf(v).Convert(typeInt64).Interface().(int64)
}

func convertToByte(v interface{}) byte {
	return reflect.ValueOf(v).Convert(typeByte).Interface().(byte)
}

func convertToRune(v interface{}) rune {
	return reflect.ValueOf(v).Convert(typeRune).Interface().(rune)
}

func convertToUintptr(v interface{}) uintptr {
	return reflect.ValueOf(v).Convert(typeUintptr).Interface().(uintptr)
}

func convertToFloat32(v interface{}) float32 {
	return reflect.ValueOf(v).Convert(typeFloat32).Interface().(float32)
}

func convertToFloat64(v interface{}) float64 {
	return reflect.ValueOf(v).Convert(typeFloat64).Interface().(float64)
}

func convertToString(v interface{}) string {
	return reflect.ValueOf(v).Convert(typeString).Interface().(string)
}

func convertToByteSlice(v interface{}) []byte {
	return reflect.ValueOf(v).Convert(typeByteSlice).Interface().([]byte)
}

func convertToRuneSlice(v interface{}) []rune {
	return reflect.ValueOf(v).Convert(typeRuneSlice).Interface().([]rune)
}
