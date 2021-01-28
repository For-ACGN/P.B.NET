package testsuite

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkGoroutine(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	ch := make(chan struct{})
	go func() {
		<-ch
	}()
	close(ch)
}

func TestMarkGoroutine_Leak(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	ch := make(chan struct{})
	go func() {
		<-ch
	}()

	delta := gm.compare()
	require.Equal(t, 1, delta)

	close(ch)
}

func TestIsDestroyed(t *testing.T) {
	t.Run("destroyed", func(t *testing.T) {
		obj := 1

		n, err := fmt.Fprintln(io.Discard, obj)
		require.Equal(t, n, 2)
		require.NoError(t, err)

		if !Destroyed(&obj) {
			t.Fatal("doesn't destroyed")
		}
	})

	t.Run("doesn't destroyed", func(t *testing.T) {
		obj := 2

		if Destroyed(&obj) {
			t.Fatal("destroyed")
		}

		n, err := fmt.Fprintln(io.Discard, obj)
		require.Equal(t, n, 2)
		require.NoError(t, err)
	})

	t.Run("is destroyed", func(t *testing.T) {
		obj := 3

		n, err := fmt.Fprintln(io.Discard, obj)
		require.Equal(t, n, 2)
		require.NoError(t, err)

		IsDestroyed(t, &obj)
	})
}

func TestMarkMemory(t *testing.T) {
	mm := MarkMemory(t)
	defer mm.Compare()
}

func TestMarkMemory_Leak(t *testing.T) {
	mm := MarkMemory(t)
	defer mm.Compare()
}

func TestTestMainCheckError(t *testing.T) {
	defer DeferForPanic(t)

	TestMainCheckError(errors.New("foo error"))
}

func TestTestMainGoroutineLeaks(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		leaks := testMainGoroutineLeaks(5)
		require.False(t, leaks)
	})

	t.Run("leaks", func(t *testing.T) {
		ch := make(chan struct{})
		RunGoroutines(func() {
			<-ch
		})

		leaks := testMainGoroutineLeaks(5)
		require.True(t, leaks)

		close(ch)
		leaks = testMainGoroutineLeaks(5)
		require.False(t, leaks)
	})

	// for code coverage
	TestMainGoroutineLeaks()
}
