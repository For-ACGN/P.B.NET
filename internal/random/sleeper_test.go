package random

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestSleeper(t *testing.T) {
	t.Run("SleepSecond", func(t *testing.T) {
		now := time.Now()

		done, sleeper := SleepSecond(2, 2)
		defer sleeper.Stop()
		<-done

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 2*time.Second)
		require.True(t, d < 5*time.Second)
	})

	t.Run("SleepMillisecond", func(t *testing.T) {
		now := time.Now()

		done, sleeper := SleepMillisecond(100, 200)
		defer sleeper.Stop()
		<-done

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 100*time.Millisecond)
		require.True(t, d < 500*time.Millisecond)
	})

	t.Run("zero value", func(t *testing.T) {
		now := time.Now()

		done, sleeper := SleepSecond(0, 0)
		defer sleeper.Stop()
		<-done

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 1*time.Second)
		require.True(t, d < 2*time.Second)
	})

	t.Run("not read", func(t *testing.T) {
		sleeper := NewSleeper()

		sleeper.SleepSecond(0, 0)
		time.Sleep(time.Second + 100*time.Millisecond)
		sleeper.SleepSecond(0, 0)
	})

	t.Run("max duration", func(t *testing.T) {
		sleeper := NewSleeper()

		d := sleeper.calculateDuration(3600*1000, 3600*1000)
		require.Equal(t, MaxSleepTime, d)
	})
}

func TestSleep(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		now := time.Now()

		Sleep(2, 2)

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 2*time.Second)
		require.True(t, d < 5*time.Second)
	})

	t.Run("timeout", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(time.Duration) *time.Timer {
			pg.Unpatch()
			defer pg.Restore()
			return time.NewTimer(2 * time.Second)
		}
		pg = monkey.Patch(time.NewTimer, patch)
		defer pg.Unpatch()

		now := time.Now()

		Sleep(100, 100)

		// maybe CPU is full
		d := time.Since(now)

		require.True(t, d > 2*time.Second)
		require.True(t, d < 5*time.Second)
	})
}
