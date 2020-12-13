package plugin

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/interpreter/anko"
	"project/internal/module"
	"project/internal/testsuite"

	_ "project/internal/interpreter/anko/goroot"
)

func init() {
	anko.Packages["project/internal/module"] = map[string]reflect.Value{}
	var (
		method module.Method
		value  module.Value
	)
	anko.PackageTypes["project/internal/module"] = map[string]reflect.Type{
		"Method": reflect.TypeOf(&method).Elem(),
		"Value":  reflect.TypeOf(&value).Elem(),
	}
}

type testExternal struct {
	sent bool
}

func (ext *testExternal) SendMessage(msg string) {
	fmt.Println(msg)
	ext.sent = true
}

func TestAnko(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	script, err := os.ReadFile("testdata/test.ank")
	require.NoError(t, err)

	ext := new(testExternal)
	ank, err := NewAnko(ext, os.Stdout, string(script))
	require.NoError(t, err)

	name := ank.Name()
	require.Equal(t, "anko-name", name)

	desc := ank.Description()
	require.Equal(t, "anko-description", desc)

	err = ank.Start()
	require.NoError(t, err)

	ank.Stop()

	require.False(t, ank.IsStarted())

	err = ank.Restart()
	require.NoError(t, err)

	require.True(t, ank.IsStarted())

	err = ank.Restart()
	require.NoError(t, err)

	require.True(t, ank.IsStarted())

	info := ank.Info()
	require.Equal(t, "anko-info", info)

	status := ank.Status()
	require.Equal(t, "anko-status", status)

	methods := ank.Methods()
	for i := 0; i < len(methods); i++ {
		fmt.Printf("%s\n\n", methods[i])
	}

	t.Run("call-IsStarted", func(t *testing.T) {
		ret, err := ank.Call("IsStarted")
		require.NoError(t, err)
		require.True(t, ret.(bool))
	})

	t.Run("call-Add", func(t *testing.T) {
		ret, err := ank.Call("Add", 1, 2)
		require.NoError(t, err)
		require.Equal(t, int64(3), ret.(int64))
	})

	t.Run("call-MultiReturn", func(t *testing.T) {
		ret, err := ank.Call("MultiReturn")
		require.NoError(t, err)
		rets := ret.([]interface{})
		require.Equal(t, "a", rets[0])
		require.EqualError(t, rets[1].(error), "b")
	})

	t.Run("call-UseExternal", func(t *testing.T) {
		ret, err := ank.Call("UseExternal")
		require.NoError(t, err)
		require.Nil(t, ret)
		require.True(t, ext.sent)
	})

	ank.Stop()
	require.False(t, ank.IsStarted())

	testsuite.IsDestroyed(t, ank)
}
