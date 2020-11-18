package pauser

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestPauser(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pauser := New(context.Background())
	require.Equal(t, StateRunning, pauser.State())

	pauser.Pause()
	require.Equal(t, StatePaused, pauser.State())

	go func() {
		time.Sleep(2 * time.Second)
		pauser.Continue()
	}()

	now := time.Now()
	pauser.Paused()
	require.True(t, time.Since(now) > time.Second)
	require.Equal(t, StateRunning, pauser.State())
}

func TestPauser_Continue(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	ctx := context.Background()

	t.Run("continue but not paused", func(t *testing.T) {
		pauser := New(ctx)
		pauser.Continue()
	})

	t.Run(" simulate continue too fast", func(t *testing.T) {
		pauser := New(ctx)
		fakeState := StatePaused
		pauser.state = &fakeState

		pauser.Continue()
		pauser.Continue()
	})
}

func TestPauser_Pause(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	ctx := context.Background()

	t.Run("not paused", func(t *testing.T) {
		pauser := New(ctx)

		now := time.Now()
		pauser.Paused()
		require.True(t, time.Since(now) < time.Second)
		require.Equal(t, StateRunning, pauser.State())
	})

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		pauser := New(ctx)
		pauser.Pause()
		require.Equal(t, StatePaused, pauser.State())

		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		now := time.Now()
		pauser.Paused()
		require.True(t, time.Since(now) > time.Second)
		require.Equal(t, StateCancel, pauser.State())

		pauser.Paused()
	})
}
