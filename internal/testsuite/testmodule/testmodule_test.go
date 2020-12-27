package testmodule

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestModule(t *testing.T) {
	mod := New()

	t.Run("Name", func(t *testing.T) {
		t.Log(mod.Name())
	})

	t.Run("Description", func(t *testing.T) {
		t.Log(mod.Description())
	})

	t.Run("Start", func(t *testing.T) {
		err := mod.Start()
		require.NoError(t, err)

		err = mod.Start()
		require.Error(t, err)

		mod.Stop()
	})

	t.Run("Stop", func(t *testing.T) {
		err := mod.Start()
		require.NoError(t, err)

		mod.Stop()
		mod.Stop()
	})

	t.Run("Restart", func(t *testing.T) {
		err := mod.Restart()
		require.NoError(t, err)

		mod.Stop()

		err = mod.Restart()
		require.NoError(t, err)
		err = mod.Restart()
		require.NoError(t, err)

		mod.Stop()
	})

	t.Run("IsStarted", func(t *testing.T) {
		mod.Stop()
		started := mod.IsStarted()
		require.False(t, started)
	})

	t.Run("Info", func(t *testing.T) {
		t.Log(mod.Info())

		err := mod.Start()
		require.NoError(t, err)
		t.Log(mod.Info())

		mod.Stop()
	})

	t.Run("Status", func(t *testing.T) {
		t.Log(mod.Status())

		err := mod.Start()
		require.NoError(t, err)
		t.Log(mod.Status())

		mod.Stop()
	})

	t.Run("Methods", func(t *testing.T) {
		t.Log(mod.Methods())
	})

	t.Run("Call", func(t *testing.T) {
		const method = "Scan"

		// common
		ret, err := mod.Call(method, "1.1.1.1")
		require.NoError(t, err)
		rets := ret.([]interface{})
		open, e := rets[0].(bool), rets[1]
		require.Nil(t, e)
		require.True(t, open)

		// empty argument
		ret, err = mod.Call(method)
		require.Error(t, err)
		require.Nil(t, ret)

		ret, err = mod.Call(method, []rune("invalid arg"))
		require.Error(t, err)
		require.Nil(t, ret)

		// empty argument
		ret, err = mod.Call(method, "")
		require.NoError(t, err)
		rets = ret.([]interface{})
		open, err = rets[0].(bool), rets[1].(error)
		require.Error(t, err)
		require.False(t, open)

		// unknown method
		ret, err = mod.Call("foo", "p1")
		require.Error(t, err)
		require.Nil(t, ret)
	})

	testsuite.IsDestroyed(t, mod)
}
