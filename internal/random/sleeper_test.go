package random

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestSleeper_Sleep(t *testing.T) {
	sleeper := NewSleeper()
	defer sleeper.Stop()

	now := time.Now()

	sleeper.Sleep(2, 2)

	// maybe CPU is full
	d := time.Since(now)

	require.True(t, d > 2*time.Second)
	require.True(t, d < 5*time.Second)
}

func TestSleeper_SleepSecond(t *testing.T) {
	sleeper := NewSleeper()
	defer sleeper.Stop()

	now := time.Now()

	done := sleeper.SleepSecond(2, 2)
	select {
	case <-done:
	case <-time.After(time.Minute):
		t.Fatal("timeout")
	}

	// maybe CPU is full
	d := time.Since(now)

	require.True(t, d > 2*time.Second)
	require.True(t, d < 5*time.Second)
}

func TestSleeper_SleepMillisecond(t *testing.T) {
	sleeper := NewSleeper()
	defer sleeper.Stop()

	now := time.Now()

	done := sleeper.SleepMillisecond(100, 200)
	select {
	case <-done:
	case <-time.After(time.Minute):
		t.Fatal("timeout")
	}

	// maybe CPU is full
	d := time.Since(now)

	require.True(t, d > 100*time.Millisecond)
	require.True(t, d < 500*time.Millisecond)
}

func TestSleeper(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		sleeper := NewSleeper()
		defer sleeper.Stop()

		now := time.Now()

		done := sleeper.SleepSecond(0, 0)
		select {
		case <-done:
		case <-time.After(time.Minute):
			t.Fatal("timeout")
		}

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 1*time.Second)
		require.True(t, d < 2*time.Second)
	})

	t.Run("timeout", func(t *testing.T) {
		sleeper := NewSleeper()
		defer sleeper.Stop()

		var pg *monkey.PatchGuard
		patch := func(time.Duration) *time.Timer {
			pg.Unpatch()
			defer pg.Restore()
			return time.NewTimer(3 * time.Second)
		}
		pg = monkey.Patch(time.NewTimer, patch)
		defer pg.Unpatch()

		now := time.Now()

		sleeper.Sleep(100, 100)

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 2*time.Second)
		require.True(t, d < 5*time.Second)
	})

	t.Run("not read done channel", func(t *testing.T) {
		sleeper := NewSleeper()
		defer sleeper.Stop()

		sleeper.SleepSecond(0, 0)
		time.Sleep(time.Second + 100*time.Millisecond)
		sleeper.SleepSecond(0, 0)
	})

	t.Run("max duration", func(t *testing.T) {
		sleeper := NewSleeper()
		defer sleeper.Stop()

		d := sleeper.calculateTime(3600*1000, 3600*1000)
		require.Equal(t, MaxSleepTime, d)
	})
}

func TestSleep(t *testing.T) {
	now := time.Now()

	Sleep(2, 2)

	// maybe CPU is full
	d := time.Since(now)

	require.True(t, d > 2*time.Second)
	require.True(t, d < 5*time.Second)
}

func TestSleepMillisecond(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		now := time.Now()

		SleepMillisecond(100, 200)

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 100*time.Millisecond)
		require.True(t, d < 500*time.Millisecond)
	})

	t.Run("timeout", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(time.Duration) *time.Timer {
			pg.Unpatch()
			defer pg.Restore()
			return time.NewTimer(3 * time.Second)
		}
		pg = monkey.Patch(time.NewTimer, patch)
		defer pg.Unpatch()

		now := time.Now()

		SleepMillisecond(20000, 20000)

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 2*time.Second)
		require.True(t, d < 5*time.Second)
	})
}
