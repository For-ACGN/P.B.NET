package security

import (
	"bytes"
	"context"
	"runtime"
	"sync"
	"time"

	"project/internal/random"
	"project/internal/xpanic"
)

func schedule(ctx context.Context, ch chan []byte) {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "schedule")
		}
	}()
	rand := random.NewRand()
	n := 100 + rand.Intn(100)
	for i := 0; i < n; i++ {
		buf := bytes.Buffer{}
		buf.Write(random.Bytes(16 + rand.Intn(1024)))
		select {
		case ch <- buf.Bytes():
		case <-ctx.Done():
			return
		}
		runtime.Gosched()
	}
}

// SwitchThread is used to create a lot of goroutine to call "select"
// that can split syscall to random threads to call.
func SwitchThread() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rand := random.NewRand()
	// must > n * (n in schedule)
	bc := make(chan []byte, 5120)
	n := 8 + rand.Intn(8)
	for i := 0; i < n; i++ {
		go schedule(ctx, bc)
	}
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
read:
	for {
		timer.Reset(250 * time.Millisecond)
		select {
		case b := <-bc:
			b[0] = byte(rand.Int63())
		case <-timer.C:
			break read
		}
	}
	random.SleepMillisecond(5, 50)
}

// SwitchThreadAsync like SwitchThread, but will not wait goroutine run finish.
func SwitchThreadAsync() <-chan struct{} {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// must > n * (n in schedule)
	bc := make(chan []byte, 5120)
	n := 8 + random.NewRand().Intn(8)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			schedule(ctx, bc)
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		cancel()
	}()
	return done
}

// WaitSwitchThreadAsync is used to wait all goroutine done.
func WaitSwitchThreadAsync(ctx context.Context, d ...<-chan struct{}) {
	for _, done := range d {
		select {
		case <-done:
		case <-ctx.Done():
		}
	}
}
