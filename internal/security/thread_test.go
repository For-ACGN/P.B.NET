package security

import (
	"context"
	"testing"
	"time"

	"project/internal/patch/monkey"
	"project/internal/random"
	"project/internal/testsuite"
)

func TestSchedule(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("panic", func(t *testing.T) {
		patch := func() *random.Rand {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(random.NewRand, patch)
		defer pg.Unpatch()

		ch := make(chan []byte, 5120)
		schedule(context.Background(), ch)
	})

	t.Run("ctx.Done()", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()

		ch := make(chan []byte, 16)
		schedule(ctx, ch)
	})
}

func TestSwitchThread(t *testing.T) {
	SwitchThread()
}

func TestSwitchThreadAsync(t *testing.T) {
	<-SwitchThreadAsync()
}

func TestWaitSwitchThreadAsync(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		done1 := SwitchThreadAsync()
		done2 := SwitchThreadAsync()
		WaitSwitchThreadAsync(context.Background(), done1, done2)
	})

	t.Run("interrupt", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// make sure the under channel in ctx is closed
		time.Sleep(time.Second)

		done1 := SwitchThreadAsync()
		done2 := SwitchThreadAsync()
		WaitSwitchThreadAsync(ctx, done1, done2)

		<-done1
		<-done2
	})
}

func BenchmarkSwitchThread(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SwitchThread()
	}

	b.StopTimer()
}

func BenchmarkSwitchThreadAsync(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		<-SwitchThreadAsync()
	}

	b.StopTimer()
}
