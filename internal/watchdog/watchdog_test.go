package watchdog

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

func testNewMockWorker(watch, process time.Duration, notice Callback) *mockWorker {
	watchDog, _ := New(logger.Test, "test", notice)
	watchDog.SetWatchInterval(watch)
	mw := mockWorker{
		process:    process,
		taskCh:     make(chan int, 16),
		watchDog:   watchDog,
		stopSignal: make(chan struct{}),
	}
	return &mw
}

func (mw *mockWorker) Start() {
	mw.wg.Add(1)
	go mw.sendTaskLoop()
	for i := 0; i < 4; i++ {
		mw.wg.Add(1)
		go mw.work()
	}
	mw.watchDog.Start()
}

func (mw *mockWorker) sendTaskLoop() {
	defer mw.wg.Done()
	select {
	case mw.taskCh <- 1:
	case <-mw.stopSignal:
		return
	}
}

func (mw *mockWorker) work() {
	defer mw.wg.Done()
	_, watcher := mw.watchDog.NewWatcher()
	for {
		select {
		case <-mw.taskCh:
			time.Sleep(mw.process)
		case <-watcher:
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
		const (
			watch   = 3 * time.Second
			process = time.Second
		)

		notice := func(int) {
			t.Fatal("watcher blocked")
		}
		worker := testNewMockWorker(watch, process, notice)
		worker.Start()

		time.Sleep(6 * time.Second)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
	})

	t.Run("block", func(t *testing.T) {
		const (
			watch   = time.Second
			process = 3 * time.Second
		)

		block := make(map[int]struct{}, 4)
		blockMu := sync.Mutex{}

		notice := func(id int) {
			blockMu.Lock()
			defer blockMu.Unlock()
			require.NotContainsf(t, block, id, "notice watcher %d multi times", id)
			block[id] = struct{}{}
		}
		worker := testNewMockWorker(watch, process, notice)
		worker.Start()

		time.Sleep(6 * time.Second)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)

		require.Len(t, block, 4)
	})
}
