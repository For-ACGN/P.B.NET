package xreflect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct{}

func TestGetStructName(t *testing.T) {
	name := GetStructName(&testStruct{})
	require.Equal(t, "testStruct", name)
	nest := struct {
		a int
		b struct {
			c int
			d int
		}
	}{}
	name = GetStructName(nest)
	expected := "struct { a int; b struct { c int; d int } }"
	require.Equal(t, expected, name)
}

func TestStructToMap(t *testing.T) {
	s := struct {
		Name string `msgpack:"name"`
		Host string `msgpack:"host"`
	}{
		Name: "aaa",
		Host: "bbb",
	}
	// point
	m := StructToMap(&s, "msgpack")
	require.Equal(t, "aaa", m["name"])
	require.Equal(t, "bbb", m["host"])
	// value
	m = StructToMap(s, "msgpack")
	require.Equal(t, "aaa", m["name"])
	require.Equal(t, "bbb", m["host"])
}

func TestStructToMapExceptZero(t *testing.T) {
	s := struct {
		Name string `msgpack:"name"`
		Host string `msgpack:"host"`
	}{
		Name: "aaa",
		Host: "",
	}
	// point
	m := StructToMapExceptZero(&s, "msgpack")
	require.Len(t, m, 1)
	require.Equal(t, "aaa", m["name"])
	// value
	m = StructToMapExceptZero(s, "msgpack")
	require.Len(t, m, 1)
	require.Equal(t, "aaa", m["name"])
}
