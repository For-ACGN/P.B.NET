package xreflect

import (
	"reflect"
	"strings"
)

// GetStructName is used to get structure name, it will not contain package name.
func GetStructName(v interface{}) string {
	// [package name].[structure name]
	s := reflect.TypeOf(v).String()
	ss := strings.Split(s, ".")
	return ss[len(ss)-1]
}

// StructToMap is used to convert structure to a string map.
func StructToMap(v interface{}, tag string) map[string]interface{} {
	typ, val := structToMap(v)
	n := val.NumField()
	m := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		key := typ.Field(i).Tag.Get(tag)
		m[key] = val.Field(i).Interface()
	}
	return m
}

// StructToMapExceptZero is used to convert structure to a string map,
// it not contain zero value like 0, false, "" and nil.
func StructToMapExceptZero(v interface{}, tag string) map[string]interface{} {
	typ, val := structToMap(v)
	n := val.NumField()
	m := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		key := typ.Field(i).Tag.Get(tag)
		field := val.Field(i)
		if !field.IsZero() {
			m[key] = field.Interface()
		}
	}
	return m
}

func structToMap(v interface{}) (reflect.Type, reflect.Value) {
	typ := reflect.TypeOf(v)
	var val reflect.Value
	if typ.Kind() == reflect.Ptr {
		val = reflect.ValueOf(v).Elem()
		typ = val.Type()
	} else {
		val = reflect.ValueOf(v)
	}
	return typ, val
}
