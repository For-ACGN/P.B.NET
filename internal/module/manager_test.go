package module

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestManager_Add(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		module := new(mockModule)

		err := manager.Add("test", module)
		require.NoError(t, err)
	})

	t.Run("with empty tag", func(t *testing.T) {
		err := manager.Add("", nil)
		require.EqualError(t, err, "empty module tag")
	})

	t.Run("is already exists", func(t *testing.T) {
		const tag = "test1"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		err = manager.Add(tag, module)
		require.EqualError(t, err, "module test1 is already exists")
	})

	t.Run("add after close", func(t *testing.T) {
		manager.Close()

		err := manager.Add("test", nil)
		require.Error(t, err)
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Delete(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		err = manager.Delete(tag)
		require.NoError(t, err)
	})

	t.Run("with empty tag", func(t *testing.T) {
		err := manager.Delete("")
		require.EqualError(t, err, "empty module tag")
	})

	t.Run("is not exist", func(t *testing.T) {
		err := manager.Delete("tag")
		require.EqualError(t, err, "module tag is not exist")
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Get(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		mod, err := manager.Get(tag)
		require.NoError(t, err)
		require.NotNil(t, mod)
	})

	t.Run("with empty tag", func(t *testing.T) {
		module, err := manager.Get("")
		require.EqualError(t, err, "empty module tag")
		require.Nil(t, module)
	})

	t.Run("is not exist", func(t *testing.T) {
		module, err := manager.Get("tag")
		require.EqualError(t, err, "module tag is not exist")
		require.Nil(t, module)
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Start(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		err = manager.Start(tag)
		require.NoError(t, err)
	})

	t.Run("with empty tag", func(t *testing.T) {
		err := manager.Start("")
		require.EqualError(t, err, "empty module tag")
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Stop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		err = manager.Stop(tag)
		require.NoError(t, err)
	})

	t.Run("with empty tag", func(t *testing.T) {
		err := manager.Stop("")
		require.EqualError(t, err, "empty module tag")
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Restart(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		err = manager.Restart(tag)
		require.NoError(t, err)
	})

	t.Run("with empty tag", func(t *testing.T) {
		err := manager.Restart("")
		require.EqualError(t, err, "empty module tag")
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Info(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		info, err := manager.Info(tag)
		require.NoError(t, err)
		t.Log(info)
	})

	t.Run("with empty tag", func(t *testing.T) {
		info, err := manager.Info("")
		require.EqualError(t, err, "empty module tag")
		require.Zero(t, info)
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Status(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	t.Run("ok", func(t *testing.T) {
		const tag = "test"

		module := new(mockModule)

		err := manager.Add(tag, module)
		require.NoError(t, err)
		status, err := manager.Status(tag)
		require.NoError(t, err)
		t.Log(status)
	})

	t.Run("with empty tag", func(t *testing.T) {
		status, err := manager.Status("")
		require.EqualError(t, err, "empty module tag")
		require.Zero(t, status)
	})

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Modules(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := NewManager()

	modules := manager.Modules()
	require.Empty(t, modules)

	module := new(mockModule)
	err := manager.Add("tag", module)
	require.NoError(t, err)
	require.Len(t, manager.Modules(), 1)

	manager.Close()

	modules = manager.Modules()
	require.Empty(t, modules)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	// TODO add more add delete

	const deleteTag = "delete"

	manager := NewManager()
	module := new(mockModule)

	init := func() {
		err := manager.Add(deleteTag, module)
		require.NoError(t, err)
	}
	add := func() {
		_ = manager.Add("test", module)
	}
	del := func() {
		_ = manager.Delete(deleteTag)
	}
	get1 := func() {
		_, _ = manager.Get("test")
	}
	get2 := func() {
		module, err := manager.Get("")
		require.EqualError(t, err, "empty module tag")
		require.Nil(t, module)
	}
	modules := func() {
		modules := manager.Modules()
		require.NotNil(t, modules)
	}
	// TODO think manager.Close()
	cleanup := func() {
		_ = manager.Delete(deleteTag)
	}
	testsuite.RunParallel(100, init, cleanup,
		add, del, get1, get2, modules)

	manager.Close()

	testsuite.IsDestroyed(t, manager)
}
