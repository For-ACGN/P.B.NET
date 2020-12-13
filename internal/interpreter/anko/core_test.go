package anko

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"project/external/anko/env"

	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestCoreType(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
// --------uint--------

v = new(uint)
if typeOf(v) != "*uint" {
	return "not *uint type"
}

v = new(uint8)
if typeOf(v) != "*uint8" {
	return "not *uint8 type"
}

v = new(uint16)
if typeOf(v) != "*uint16" {
	return "not *uint16 type"
}

v = new(uint32)
if typeOf(v) != "*uint32" {
	return "not *uint32 type"
}

v = new(uint64)
if typeOf(v) != "*uint64" {
	return "not *uint64 type"
}

// --------int--------

v = new(int)
if typeOf(v) != "*int" {
	return "not *int type"
}

v = new(int8)
if typeOf(v) != "*int8" {
	return "not *int8 type"
}

v = new(int16)
if typeOf(v) != "*int16" {
	return "not *int16 type"
}

v = new(int32)
if typeOf(v) != "*int32" {
	return "not *int32 type"
}

v = new(int64)
if typeOf(v) != "*int64" {
	return "not *int64 type"
}

// --------other--------

v = new(byte)
if typeOf(v) != "*uint8" {
	return "not *uint8 type"
}

v = new(rune)
if typeOf(v) != "*int32" {
	return "not *int32 type"
}

v = new(uintptr)
if typeOf(v) != "*uintptr" {
	return "not *uintptr type"
}

v = new(float32)
if typeOf(v) != "*float32" {
	return "not *float32 type"
}

v = new(float64)
if typeOf(v) != "*float64" {
	return "not *float64 type"
}

return true
`
	testRun(t, src, false, true)
}

func TestCoreKeys(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
m = {"foo": "bar", "bar": "baz"}
println(keys(m))
return true
`
	testRun(t, src, false, true)
}

func TestCoreRange(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("no parameter", func(t *testing.T) {
		const src = `range()`
		testRun(t, src, true, nil)
	})

	t.Run("1p", func(t *testing.T) {
		const src = `range(3)`
		testRun(t, src, false, []int64{0, 1, 2})
	})

	t.Run("2p", func(t *testing.T) {
		const src = `range(1, 3)`
		testRun(t, src, false, []int64{1, 2})
	})

	t.Run("3p", func(t *testing.T) {
		const src = `range(1, 10, 2)`
		testRun(t, src, false, []int64{1, 3, 5, 7, 9})
	})

	t.Run("3p-zero step", func(t *testing.T) {
		const src = `range(1, 10, 0)`
		testRun(t, src, true, nil)
	})

	t.Run("4p", func(t *testing.T) {
		const src = `range(1, 2, 3, 4)`
		testRun(t, src, true, nil)
	})
}

func TestCoreInstance(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
sa = make(type sa, make(struct{
 A string,
 B string
}))

i1 = instance(sa)
i1.A = "acg"

i2 = instance(make(sa))
i2.A = "abc"
i2.B = "bbb"

if i1.A != "acg" {
	return "invalid i1.A"
}
if i2.A != "abc" {
	return "invalid i2.A"
}
if i2.B != "bbb" {
	return "invalid i2.B"
}
return true
`
	testRun(t, src, false, true)
}

func TestCoreArrayType(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
typ = arrayType(make(int8), 4)
if typ.String() != "[4]int8" {
	return "invalid type"
}

// i1 = make(typ) undefined
// i1 = new(typ) undefined
i1 = instance(typ)
println(i1)

return true
`
	testRun(t, src, false, true)
}

func TestCoreArray(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
a = array(make(int8), 4)
if typeOf(a) != "*[4]int8" {
	return "not *[4]int8 type"
}
a = *a

p1 = &a
printf("%p", p1)

a[1] = 123
if a[1] != 123 {
	return "invalid array value"
}

p2 = &a
printf("%p", p2)

if p1 != p2 {
	return "address changed"
}
return true
`
	testRun(t, src, false, true)
}

func TestCoreSlice(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
s = slice(array(make(int8), 4))
if typeOf(s) != "[]int8" {
	return "not []int8 type"
}
println(s)
return true
`
	testRun(t, src, false, true)
}

func TestCoreTypeOf(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
v = 10
if typeOf(v) != "int64"{
	return "not int64 type"
}
return true
`
	testRun(t, src, false, true)
}

func TestCoreKindOf(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("int64", func(t *testing.T) {
		const src = `
v = 10
if kindOf(v) != "int64" {
	return "not int64 kind"
}
return true
`
		testRun(t, src, false, true)
	})

	t.Run("nil", func(t *testing.T) {
		const src = `
v = nil
if kindOf(v) != "nil kind" {
	return "not nil kind"
}
return true
`
		testRun(t, src, false, true)
	})
}

func TestDefineCoreType(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	e := env.NewEnv()
	patch := func(interface{}, string, interface{}) error {
		return monkey.Error
	}
	pg := monkey.PatchInstanceMethod(e, "DefineType", patch)
	defer pg.Unpatch()

	defer testsuite.DeferForPanic(t)
	defineCoreType(e)
}

func TestDefineCoreFunc(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	e := env.NewEnv()
	patch := func(interface{}, string, interface{}) error {
		return monkey.Error
	}
	pg := monkey.PatchInstanceMethod(e, "Define", patch)
	defer pg.Unpatch()

	defer testsuite.DeferForPanic(t)
	defineCoreFunc(e)
}

// function in the module can only set the root module, instance will bug.
// reference:
// https://github.com/mattn/anko/issues/315

func TestModule(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	e := NewEnv()

	const src = `
module a {
	b = 1

    func Input(v1) {
		b = v1
		return
	}

	func Output() {
		return b
	}
}
c = a
d = a

d.Input(2)

println(c) // 1
println(d) // except 2 but is 1
println(a) // except 1 but is 2

a.Input(3)
println(c)
println(d)
println(a)

return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(e, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	e.Close()

	ne := e.env
	testsuite.IsDestroyed(t, e)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestAnkoModule(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	e := NewEnv()

	const src = `
module Test {
	inner1 = "acg"
	inner2 = 1

	func Input(v1, v2) {
		inner1 = v1
		inner2 = v2
		return
	}

	func Output() {
		return inner1, inner2
	}
}
func New() {
	var mod = Test
	return mod
}
return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(e, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	// get New function
	newFn, err := e.Get("New")
	require.NoError(t, err)
	create := newFn.(func(context.Context) (reflect.Value, reflect.Value))

	ctx := context.Background()

	f := func() {
		// create instance
		mod, re := create(ctx)
		require.True(t, re.IsNil())
		module := mod.Interface().(*env.Env)

		// get output function
		outputFn, err := module.Get("Output")
		require.NoError(t, err)
		output := outputFn.(func(context.Context) (reflect.Value, reflect.Value))

		// call output
		ret, re := output(ctx)
		require.True(t, re.IsNil())
		i1 := ret.Index(0).Interface()
		i2 := ret.Index(1).Interface()

		// require.Equal(t, "acg", i1)
		// require.Equal(t, int64(1), i2)

		fmt.Println(i1)
		fmt.Println(i2)

		// get input function
		inputFn, err := module.Get("Input")
		require.NoError(t, err)
		input := inputFn.(func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value))

		// call input
		ret, re = input(ctx, reflect.ValueOf("aaa"), reflect.ValueOf(int64(2)))
		require.True(t, re.IsNil())
		require.True(t, ret.IsNil())

		ret, re = output(ctx)
		require.True(t, re.IsNil())
		i1 = ret.Index(0).Interface()
		i2 = ret.Index(1).Interface()

		// require.Equal(t, "aaa", i1)
		// require.Equal(t, int64(2), i2)

		fmt.Println(i1)
		fmt.Println(i2)
	}
	f()
	f()

	e.Close()

	ne := e.env
	testsuite.IsDestroyed(t, e)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

type FooStruct struct {
	function  func(string)
	Pointer   *int
	Slice     []string
	Map       map[string]string
	Channel   chan string
	Function  func(string)
	Interface interface{}
	Transport http.RoundTripper
	Str2      FooStruct2
	Str2p     *FooStruct2
}

type FooStruct2 struct {
	Pointer *int
}

// Println is used to check structure fields are Zero Value.
func (f *FooStruct) Println() {
	fmt.Println("func(unexported):", f.function == nil)
	fmt.Println("pointer:", f.Pointer == nil)
	fmt.Println("slice:", f.Slice == nil)
	fmt.Println("map:", f.Map == nil)
	fmt.Println("chan:", f.Channel == nil)
	fmt.Println("func:", f.Function == nil)
	fmt.Println("interface{}:", f.Interface == nil)
	fmt.Println("interface:", f.Transport == nil)
	fmt.Println("str2:", f.Str2.Pointer == nil)
	fmt.Println("str2p:", f.Str2p == nil)
	fmt.Println()
}

func TestAnkoMakeStruct(t *testing.T) {
	// Zero Value
	fs1 := new(FooStruct)
	fs1.Println()
	fs2 := FooStruct{}
	fs2.Println()

	// some fields not Zero Value
	e := NewEnv()
	err := e.DefineType("FooStruct", reflect.TypeOf(fs1).Elem())
	require.NoError(t, err)

	const src = `
fs1 = new(FooStruct)
fs1.Println()

fs2 = make(FooStruct)
fs2.Println()
`
	stmt := testParseSrc(t, src)
	_, err = Run(e, stmt)
	require.NoError(t, err)

	f1, err := e.Get("fs1")
	require.NoError(t, err)
	f1v := f1.(*FooStruct)
	testCheckAnkoStruct(t, f1v)

	f2, err := e.Get("fs2")
	require.NoError(t, err)
	f2v := f2.(FooStruct)
	testCheckAnkoStruct(t, &f2v)
}

func testCheckAnkoStruct(t *testing.T, f *FooStruct) {
	require.Nil(t, f.function)
	require.Nil(t, f.Pointer)
	require.Nil(t, f.Slice)
	require.Nil(t, f.Map)
	require.Nil(t, f.Channel)
	require.Nil(t, f.Function)
	require.Nil(t, f.Interface)
	require.Nil(t, f.Transport)
	require.Nil(t, f.Str2.Pointer)
	require.Nil(t, f.Str2p)
}
