package anko

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/testsuite"

	_ "project/internal/interpreter/anko/goroot"
	_ "project/internal/interpreter/anko/thirdparty"
)

func TestEnv(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	const src = `
fmt = import("fmt")
msgpack = import("github.com/vmihailenco/msgpack/v5")

data, err = msgpack.Marshal("acg")
if err != nil {
return err
}
fmt.Println(data)

io = import("io")
println(io.ErrUnexpectedEOF)

acg = "acg"
func bb() {
	println(acg)
	acg = "acg2"
}
bb()
if acg != "acg2" {
	println(acg)
	return acg
}

for i in []byte{1,2,3} {
	println(i)
}

return true
`
	testRun(t, src, false, true)
}

func TestEnv_defined(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("defined", func(t *testing.T) {
		const src = `
a = "out"
return defined("a")
`
		testRun(t, src, false, true)
	})

	t.Run("not defined", func(t *testing.T) {
		const src = `
return defined("a")
`
		testRun(t, src, false, false)
	})
}

func TestEnv_definedType(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("defined", func(t *testing.T) {
		const src = `
make(type a, int8(0))
return definedType("a")
`
		testRun(t, src, false, true)
	})

	t.Run("not defined", func(t *testing.T) {
		const src = `
return definedType("a")
`
		testRun(t, src, false, false)
	})
}

func TestEnv_eval(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		const src = `
a = "out"
eval('a = "in eval"')
println(a)
return true
`
		testRun(t, src, false, true)
	})

	t.Run("invalid source", func(t *testing.T) {
		const src = `
val, err = eval("a -- a")
if err == nil {
	return "no error"
}
println("val:", val)
println("error:", err)
return true
`
		testRun(t, src, false, true)
	})

	t.Run("eval with error", func(t *testing.T) {
		const src = `
src = "a = 10\n"
src += "println(a)\n"
src += "println(b)\n"
val, err = eval(src)
if err == nil {
	return "no error"
}
println("val:", val)
println("error:", err)
return true
`
		testRun(t, src, false, true)
	})

	t.Run("nest", func(t *testing.T) {
		const src = `
a = "out"
eval("eval('a = \"nest eval\"')")
println(a)
return true
`
		testRun(t, src, false, true)
	})
}

func TestEnv_Define(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	env := NewEnv()

	// common
	err := env.Define("out_println", fmt.Println)
	require.NoError(t, err)

	// reflect.Value
	err = env.Define("out_print", reflect.ValueOf(fmt.Print))
	require.NoError(t, err)

	// nil value
	err = env.Define("out_nil", nil)
	require.NoError(t, err)

	// invalid symbol
	err = env.Define("out_nil.out", nil)
	require.Error(t, err)

	const src = `
out_println("println")
out_print("print\n")
out_println(out_nil)
return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(env, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	env.Close()

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestEnv_Get(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	env := NewEnv()

	const src = `
inner = "test"
return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(env, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	t.Run("common", func(t *testing.T) {
		inner, err := env.Get("inner")
		require.NoError(t, err)
		require.Equal(t, "test", inner)
	})

	t.Run("is not exist", func(t *testing.T) {
		inner, err := env.Get("foo")
		require.Error(t, err)
		require.Nil(t, inner)
	})

	env.Close()

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestEnv_GetValue(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	env := NewEnv()

	const src = `
inner = "test"
pointer = new(string)
return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(env, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	t.Run("common", func(t *testing.T) {
		inner, err := env.GetValue("inner")
		require.NoError(t, err)
		require.Equal(t, "test", inner.Interface())

		// pointer
		pointer, err := env.GetValue("pointer")
		require.NoError(t, err)
		require.Equal(t, "", pointer.Elem().Interface())

		pointer.Elem().SetString("set")

		pointer, err = env.GetValue("pointer")
		require.NoError(t, err)
		require.Equal(t, "set", pointer.Elem().Interface())
	})

	t.Run("is not exist", func(t *testing.T) {
		inner, err := env.GetValue("foo")
		require.Error(t, err)
		require.Equal(t, NilValue.Interface(), inner.Interface())
	})

	env.Close()

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestEnv_DefineType(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	type out struct {
		A string
		B string
		c string
	}

	env := NewEnv()

	o := out{
		c: "out",
	}

	// common
	err := env.DefineType("out", o)
	require.NoError(t, err)

	// reflect.Value
	err = env.DefineType("out2", reflect.TypeOf(o))
	require.NoError(t, err)

	// nil value
	err = env.DefineType("out_nil", nil)
	require.NoError(t, err)

	// invalid symbol
	err = env.DefineType("out_nil.out", nil)
	require.Error(t, err)

	const src = `
out = new(out)
out.A = "acg"
out.B = "bbb"
// out.c = "ccc" // can't access
println(out)

out2 = new(out2)
out2.A = "acg"
out2.B = "bbb"
println(out2)

// nn = new(out_nil) // error

return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(env, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	env.Close()

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestEnv_Type(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	type out struct {
		A string
		B string
		c string
	}

	env := NewEnv()

	o := out{
		c: "out",
	}

	err := env.DefineType("out", o)
	require.NoError(t, err)

	// ok
	typ, err := env.Type("out")
	require.NoError(t, err)
	require.Equal(t, "anko.out", typ.String())

	// not exist
	typ, err = env.Type("foo")
	require.Error(t, err)
	require.Equal(t, NilType, typ)
}

func TestEnv_SetOutput(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	output := bytes.NewBuffer(make([]byte, 0, 1024))

	env := NewEnv()
	env.SetOutput(output)

	const src = `
printf("%s\n","printf")
print("print\n")
println("println")
return true
`
	stmt := testParseSrc(t, src)

	val, err := Run(env, stmt)
	require.NoError(t, err)
	require.Equal(t, true, val)

	env.Close()

	fmt.Println(output)

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}

func TestNewRuntime(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	var rt *runtime
	patch := func(interface{}, string, interface{}) error {
		return monkey.Error
	}
	pg := monkey.PatchInstanceMethod(rt, "DefineValue", patch)
	defer pg.Unpatch()

	defer testsuite.DeferForPanic(t)
	newRuntime(NewEnv())
}
