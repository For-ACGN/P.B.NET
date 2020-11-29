package pauser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestPauser(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pauser := New()
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

	pauser.Close()
	require.Equal(t, StateClosed, pauser.State())
}

func TestPauser_Pause(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("not paused", func(t *testing.T) {
		pauser := New()

		now := time.Now()
		pauser.Paused()
		require.True(t, time.Since(now) < time.Second)
		require.Equal(t, StateRunning, pauser.State())
	})

	t.Run("closed", func(t *testing.T) {
		pauser := New()

		pauser.Pause()
		require.Equal(t, StatePaused, pauser.State())

		go func() {
			time.Sleep(2 * time.Second)
			pauser.Close()
		}()

		now := time.Now()
		pauser.Paused()
		require.True(t, time.Since(now) > time.Second)
		require.Equal(t, StateClosed, pauser.State())

		pauser.Paused()
	})
}

func TestPauser_Continue(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("continue but not paused", func(t *testing.T) {
		pauser := New()
		pauser.Continue()
	})

	t.Run("simulate continue too fast", func(t *testing.T) {
		pauser := New()
		fakeState := StatePaused
		pauser.state = &fakeState

		pauser.Continue()
		pauser.Continue()
	})
}
