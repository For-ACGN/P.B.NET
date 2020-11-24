package anko

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestNewEnv(t *testing.T) {
	env := NewEnv()
	require.NotNil(t, env)

	fmt.Println(len(Packages))
	fmt.Println(len(Types))

	v, err := env.Get("keys")
	require.NoError(t, err)
	require.NotNil(t, v)

	env.Close()

	testsuite.IsDestroyed(t, env)
}

func TestParseSrc(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		const s = `
a = 10
println(a)
`
		src := strings.Repeat(s, 1)
		stmt, err := ParseSrc(src)
		require.NoError(t, err)
		require.NotNil(t, stmt)
		require.NotEqual(t, s, src)

		testsuite.IsDestroyed(t, stmt)
	})

	t.Run("invalid", func(t *testing.T) {
		const s = `
a = 10
println(a)
a -- a
`
		src := strings.Repeat(s, 1)
		stmt, err := ParseSrc(src)
		require.Error(t, err)
		require.Nil(t, stmt)
		require.NotEqual(t, s, src)

		t.Log(err)
	})

	t.Run("empty", func(t *testing.T) {
		src := strings.Repeat("", 1)
		stmt, err := ParseSrc(src)
		require.Error(t, err)
		require.Nil(t, stmt)
	})

	t.Run("exploit", func(t *testing.T) {
		src := strings.Repeat("\ue031", 1)
		stmt, err := ParseSrc(src)
		require.Error(t, err)
		require.Nil(t, stmt)
	})
}

func testParseSrc(t *testing.T, s string) Stmt {
	src := strings.Repeat(s, 1)
	stmt, err := ParseSrc(src)
	require.NoError(t, err)
	require.NotEqual(t, s, src)
	return stmt
}

func TestRunContext(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		const src = `
a = 10
println(a)
`
		stmt := testParseSrc(t, src)

		env := NewEnv()
		val, err := Run(env, stmt)
		require.NoError(t, err)
		t.Log(val)

		env.Close()

		testsuite.IsDestroyed(t, env)
		testsuite.IsDestroyed(t, stmt)
	})

	t.Run("run one stmt twice", func(t *testing.T) {
		const src = `
a = 10
println(a)
`
		stmt := testParseSrc(t, src)

		env1 := NewEnv()
		val, err := Run(env1, stmt)
		require.NoError(t, err)
		t.Log(val)
		env1.Close()
		testsuite.IsDestroyed(t, env1)

		env2 := NewEnv()
		val, err = Run(env2, stmt)
		require.NoError(t, err)
		t.Log(val)
		env2.Close()
		testsuite.IsDestroyed(t, env2)

		testsuite.IsDestroyed(t, stmt)
	})

	t.Run("run error", func(t *testing.T) {
		const src = `
a = 10
println(a)

println(b)
`
		stmt := testParseSrc(t, src)

		env := NewEnv()
		val, err := Run(env, stmt)
		require.Error(t, err)
		t.Log(val, err)

		env.Close()

		testsuite.IsDestroyed(t, env)
		testsuite.IsDestroyed(t, stmt)
	})

	t.Run("cancel", func(t *testing.T) {
		env := NewEnv()
		err := env.Define("sleep", func() {
			time.Sleep(time.Second)
		})
		require.NoError(t, err)

		const src = `
a = 10
for {
	println(a)
	sleep()
}
`
		stmt := testParseSrc(t, src)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		val, err := RunContext(ctx, env, stmt)
		require.Error(t, err)
		t.Log(val, err)

		env.Close()

		testsuite.IsDestroyed(t, env)
		testsuite.IsDestroyed(t, stmt)
	})
}

func testRun(t *testing.T, src string, fail bool, expected interface{}) {
	stmt := testParseSrc(t, src)

	env := NewEnv()
	val, err := Run(env, stmt)
	if fail {
		require.Error(t, err)
		t.Log(val, err)
	} else {
		require.NoError(t, err)
		t.Log(val)
	}
	require.Equal(t, expected, val)

	env.Close()

	ne := env.env
	testsuite.IsDestroyed(t, env)
	testsuite.IsDestroyed(t, ne)
	testsuite.IsDestroyed(t, stmt)
}
