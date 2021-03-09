package watchdog

import (
	"sync"
	"testing"
	"time"

	"project/internal/logger"
	"project/internal/testsuite"
)

type mockWorker struct {
	process time.Duration

	taskCh   chan int
	watchDog *WatchDog

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

func testNewMockWorker(interval, process time.Duration, notice Callback) *mockWorker {
	watchDog, _ := New(logger.Test, "test", notice)
	watchDog.SetWatchInterval(interval)
	mw := mockWorker{
		process:    process,
		taskCh:     make(chan int, 1),
		watchDog:   watchDog,
		stopSignal: make(chan struct{}),
	}
	return &mw
}

func (mw *mockWorker) Start() {
	mw.wg.Add(2)
	go mw.sendTaskLoop()
	go mw.processLoop()
	mw.watchDog.Start()
}

func (mw *mockWorker) sendTaskLoop() {
	defer mw.wg.Done()
	ticker := time.NewTicker(mw.process)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case mw.taskCh <- 1:
			case <-mw.stopSignal:
				return
			}
		case <-mw.stopSignal:
			return
		}
	}
}

func (mw *mockWorker) processLoop() {
	defer mw.wg.Done()
	_, watchDog := mw.watchDog.NewWatcher()
	for {
		select {
		case <-mw.taskCh:
			time.Sleep(mw.process)
		case <-watchDog:
		case <-mw.stopSignal:
			return
		}
	}
}

func (mw *mockWorker) Stop() {
	mw.closeOnce.Do(func() {
		close(mw.stopSignal)
		mw.wg.Wait()
		mw.watchDog.Stop()
	})
}

func TestWatchDog(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		notice := func(int) {
			t.Fatal("watcher blocked")
		}
		mw := testNewMockWorker(3*time.Second, time.Second, notice)
		mw.Start()

		time.Sleep(6 * time.Second)

		mw.Stop()

		testsuite.IsDestroyed(t, mw)
	})

	t.Run("block", func(t *testing.T) {

	})
}
