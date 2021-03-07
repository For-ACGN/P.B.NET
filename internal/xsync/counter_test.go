package xsync

import (
	"testing"
	"time"

	"project/internal/testsuite"
)

func TestCounter(t *testing.T) {
	counter := Counter{}

	serve := func() {
		counter.Add(1)
		go func() {
			defer counter.Done()
			time.Sleep(10 * time.Millisecond)
		}()
	}
	stop := func() {
		counter.Wait()
	}
	fns := make([]func(), 101)
	for i := 0; i < 100; i++ {
		fns[i] = serve
	}
	fns[100] = stop
	testsuite.RunParallelTest(100, nil, nil, fns...)

	testsuite.IsDestroyed(t, &counter)
}

func TestCounter_Add(t *testing.T) {
	// negative counter
	counter := Counter{}

	defer testsuite.DeferForPanic(t)
	counter.Done()
}

func TestCounter_Wait(t *testing.T) {
	t.Run("max delay", func(t *testing.T) {
		counter := Counter{}

		counter.Add(1)
		go func() {
			defer counter.Done()
			time.Sleep(2 * time.Second)
		}()

		counter.Wait()

		testsuite.IsDestroyed(t, &counter)
	})

	t.Run("panic", func(t *testing.T) {
		counter := Counter{}

		counter.Add(1)
		go func() {
			defer testsuite.DeferForPanic(t)
			// negative counter
			defer counter.Add(-2)
			time.Sleep(10 * time.Millisecond)
		}()

		defer testsuite.DeferForPanic(t)
		counter.Wait()

		testsuite.IsDestroyed(t, &counter)
	})
}
