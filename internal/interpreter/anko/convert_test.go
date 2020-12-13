package anko

import (
	"testing"

	"project/external/anko/env"

	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestConvert(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
v = 123

// --------uint--------

cv = uint(v)
if typeOf(cv) != "uint" {
	return "not uint type"
}

cv = uint8(v)
if typeOf(cv) != "uint8" {
	return "not uint8 type"
}

cv = uint16(v)
if typeOf(cv) != "uint16" {
	return "not uint16 type"
}

cv = uint32(v)
if typeOf(cv) != "uint32" {
	return "not uint32 type"
}

cv = uint64(v)
if typeOf(cv) != "uint64" {
	return "not uint64 type"
}

// --------int--------

cv = int(v)
if typeOf(cv) != "int" {
	return "not int type"
}

cv = int8(v)
if typeOf(cv) != "int8" {
	return "not int8 type"
}

cv = int16(v)
if typeOf(cv) != "int16" {
	return "not int16 type"
}

cv = int32(v)
if typeOf(cv) != "int32" {
	return "not int32 type"
}

cv = int64(v)
if typeOf(cv) != "int64" {
	return "not int64 type"
}

// --------other--------

cv = byte(v)
if typeOf(cv) != "uint8" {
	return "not byte type"
}

cv = rune(v)
if typeOf(cv) != "int32" {
	return "not rune type"
}

cv = uintptr(v)
if typeOf(cv) != "uintptr" {
	return "not uintptr type"
}

cv = float32(v)
if typeOf(cv) != "float32" {
	return "not float32 type"
}

cv = float64(v)
if typeOf(cv) != "float64" {
	return "not float64 type"
}

// --------string--------

cv = string([]byte{97, 98, 99})
if typeOf(cv) != "string" {
	return "not string type"
}
if cv != "abc" {
	return "not abc"
}

cv = string([]rune{97, 98, 99})
if typeOf(cv) != "string" {
	return "not string type"
}
if cv != "abc" {
	return "not abc"
}

// --------byte slice--------

cv = byteSlice("abc")
if typeOf(cv) != "[]uint8" {
	return "not []byte type"
}
if cv != []byte{97, 98, 99} {
	return "not []byte abc"
}

// --------rune slice--------

cv = runeSlice("abc")
if typeOf(cv) != "[]int32" {
	return "not []rune type"
}
if cv != []rune{97, 98, 99} {
	return "not []byte abc"
}

return true
`
	testRun(t, src, false, true)
}

func TestDefineConvert(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	e := env.NewEnv()
	patch := func(interface{}, string, interface{}) error {
		return monkey.Error
	}
	pg := monkey.PatchInstanceMethod(e, "Define", patch)
	defer pg.Unpatch()

	defer testsuite.DeferForPanic(t)
	defineConvert(e)
}
