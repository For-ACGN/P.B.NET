package goroot

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"project/internal/interpreter/anko"
	"project/internal/testsuite"
)

func TestUnsafeAboutStruct(t *testing.T) {
	type s struct {
		A int32
		B int32
	}
	val := int64(256)

	aa := (*s)(unsafe.Pointer(&val))
	fmt.Println(aa.A)
	fmt.Println(aa.B)

	n := *(*[8]byte)(unsafe.Pointer(&val))
	fmt.Println(n)
}

func testRun(t *testing.T, s string, fail bool, expected interface{}) {
	src := strings.Repeat(s, 1)
	stmt, err := anko.ParseSrc(src)
	require.NoError(t, err)
	require.NotEqual(t, s, src)

	env := anko.NewEnv()
	val, err := anko.Run(env, stmt)
	if fail {
		require.Error(t, err)
		t.Log(val, err)
	} else {
		require.NoError(t, err)
		t.Log(val)
	}
	require.Equal(t, expected, val)

	env.Close()

	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, stmt)
}

func TestUnsafe(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("sizeof and alignof", func(t *testing.T) {
		const src = `
unsafe = import("unsafe")

val = 256

size = unsafe.Sizeof(val)
if size != 8 {
	return size
}

align = unsafe.Alignof(val)
if align != 8 {
	return align
}

return true
`
		testRun(t, src, false, true)
	})

	t.Run("offsetof", func(t *testing.T) {
		const src = `
unsafe = import("unsafe")

s = make(struct {
	A int64,
	B int64
})

offset = unsafe.Offsetof(s, "B")
if offset != 8 {
	return offset
}
return true
`
		testRun(t, src, false, true)
	})

	t.Run("offsetof not exist", func(t *testing.T) {
		const src = `
unsafe = import("unsafe")

s = make(struct {
	A int64,
	B int64
})

unsafe.Offsetof(s, "C")
`
		testRun(t, src, true, nil)
	})

	t.Run("convert to struct", func(t *testing.T) {
		// convert to struct, like these golang code
		// p := (*testStruct)(unsafe.Pointer(&Int64))
		const src = `
// 16777217 = []byte{0x01, 0x00, 0x00, 0x01}
// 72057598349672449 = []byte{0x01, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01}

unsafe = import("unsafe")
reflect = import("reflect")

val = 256
valPtr = &val

dstType = make(struct {
	A int32,
	B int32
})
p = unsafe.Convert(&val, dstType)

println(p.A, p.B)

// byte order
if !(p.A == 256 || p.B == 256) {
	return val
}

// cover memory
p.A = 16777217
p.B = 16777217

if val != 72057598349672449 {
	return val
}

if valPtr != &val {
	return "val address is changed"
}

return true
`
		testRun(t, src, false, true)
	})

	t.Run("convert to array", func(t *testing.T) {
		// make [8]byte and test ConvertWithType, like these golang code
		// p := (*[8]byte)(unsafe.Pointer(&Int64))
		const src = `
// 72057594037927937 is []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

unsafe = import("unsafe")
reflect = import("reflect")

val = 256
valPtr = &val

dstTyp = arrayType(make(byte), 8) // [8]byte
p = unsafe.ConvertWithType(&val, dstTyp)
p = *p

println(reflect.TypeOf(p))

// cover memory
for i = 0; i < 8; i++ {
	p[i] = 0
}
p[0] = 1
p[7] = 1
if val != 72057594037927937 {
	return val
}

if valPtr != &val {
	return "val address is changed"
}

return true
`
		testRun(t, src, false, true)
	})
}
