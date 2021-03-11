package watchdog

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/patch/monkey"
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

func testNewMockWorker(period, process time.Duration, onBlock OnBlock) *mockWorker {
	watchDog := New(logger.Test, "test", onBlock)
	watchDog.SetPeriod(period)
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
	go mw.taskSender()
	for i := 0; i < 4; i++ {
		mw.wg.Add(1)
		go mw.work()
	}
	mw.watchDog.Start()
}

func (mw *mockWorker) taskSender() {
	defer mw.wg.Done()
	for {
		select {
		case mw.taskCh <- 1:
		case <-mw.stopSignal:
			return
		}
	}
}

func (mw *mockWorker) work() {
	defer mw.wg.Done()
	receiver := mw.watchDog.Receiver()
	defer receiver.Stop()
	for {
		select {
		case <-mw.stopSignal:
			return
		default:
			receiver.Receive()
		}

		select {
		case <-mw.taskCh:
			time.Sleep(mw.process)
		case <-receiver.Signal:
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
			period  = 100 * time.Millisecond
			process = 10 * time.Millisecond
		)

		onBlock := func(context.Context, int32) {
			t.Fatal("watcher blocked")
		}
		worker := testNewMockWorker(period, process, onBlock)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		num := worker.watchDog.ReceiversNum()
		require.Equal(t, 4, num)
		require.Empty(t, worker.watchDog.BlockedID())

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)
	})

	t.Run("nil onBlock", func(t *testing.T) {
		const (
			period  = 100 * time.Millisecond
			process = 250 * time.Millisecond
		)

		worker := testNewMockWorker(period, process, nil)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		num := worker.watchDog.ReceiversNum()
		require.Equal(t, 4, num)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)
	})

	t.Run("block", func(t *testing.T) {
		const (
			period  = 100 * time.Millisecond
			process = 3 * time.Second
		)

		block := make(map[int32]struct{}, 4)
		blockMu := sync.Mutex{}

		onBlock := func(ctx context.Context, id int32) {
			blockMu.Lock()
			defer blockMu.Unlock()
			require.NotContainsf(t, block, id, "notice watcher %d multi times", id)
			block[id] = struct{}{}
		}
		worker := testNewMockWorker(period, process, onBlock)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		num := worker.watchDog.ReceiversNum()
		require.Equal(t, 4, num)
		require.Len(t, worker.watchDog.BlockedID(), 4)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)

		require.Len(t, block, 4)
	})

	t.Run("running", func(t *testing.T) {
		const (
			period  = 100 * time.Millisecond
			process = 250 * time.Millisecond
		)

		block := make(map[int32]int, 4)
		blockMu := sync.Mutex{}

		onBlock := func(ctx context.Context, id int32) {
			blockMu.Lock()
			defer blockMu.Unlock()
			block[id]++
		}
		worker := testNewMockWorker(period, process, onBlock)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		num := worker.watchDog.ReceiversNum()
		require.Equal(t, 4, num)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)

		require.Len(t, block, 4)
		for i := int32(0); i < 4; i++ {
			require.GreaterOrEqual(t, block[i], 3)
		}
	})
}

func TestWatchDog_SetPeriod(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("before start", func(t *testing.T) {
		watchDog := New(logger.Test, "test", nil)

		period := watchDog.GetPeriod()
		require.Equal(t, defaultPeriod, period)

		watchDog.SetPeriod(5 * time.Second)
		period = watchDog.GetPeriod()
		require.Equal(t, 5*time.Second, period)

		watchDog.SetPeriod(0)
		period = watchDog.GetPeriod()
		require.Equal(t, defaultPeriod, period)

		watchDog.SetPeriod(2 * time.Millisecond)
		period = watchDog.GetPeriod()
		require.Equal(t, defaultPeriod, period)

		watchDog.SetPeriod(10 * time.Minute)
		period = watchDog.GetPeriod()
		require.Equal(t, defaultPeriod, period)

		testsuite.IsDestroyed(t, watchDog)
	})

	t.Run("after start", func(t *testing.T) {
		const (
			period  = 500 * time.Millisecond
			process = 200 * time.Millisecond
		)

		block := make(map[int32]int, 4)
		blockMu := sync.Mutex{}

		onBlock := func(ctx context.Context, id int32) {
			blockMu.Lock()
			defer blockMu.Unlock()
			block[id]++
		}
		worker := testNewMockWorker(period, process, onBlock)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		watchDog.SetPeriod(100 * time.Millisecond)

		time.Sleep(time.Second)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)

		require.Len(t, block, 4)
		for i := int32(0); i < 4; i++ {
			require.GreaterOrEqual(t, block[i], 3)
		}
	})
}

func TestWatchDog_kickLoop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("panic in kickLoop()", func(t *testing.T) {
		var (
			wd    *WatchDog
			count int
		)
		patch := func(wd *WatchDog) time.Duration {
			count++
			if count == 2 {
				panic(monkey.Panic)
			}
			return wd.period.Load().(time.Duration)
		}
		pg := monkey.PatchInstanceMethod(wd, "GetPeriod", patch)
		defer pg.Unpatch()

		const (
			period  = 500 * time.Millisecond
			process = 200 * time.Millisecond
		)

		worker := testNewMockWorker(period, process, nil)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(2 * time.Second)

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)
	})

	t.Run("panic in onBlock", func(t *testing.T) {
		const (
			period  = 100 * time.Millisecond
			process = 250 * time.Millisecond
		)

		var blocked int32
		onBlock := func(context.Context, int32) {
			atomic.StoreInt32(&blocked, 1)
			panic("test panic")
		}
		worker := testNewMockWorker(period, process, onBlock)
		worker.Start()
		watchDog := worker.watchDog

		time.Sleep(time.Second)

		num := worker.watchDog.ReceiversNum()
		require.Equal(t, 4, num)

		require.Equal(t, int32(1), atomic.LoadInt32(&blocked))

		worker.Stop()

		testsuite.IsDestroyed(t, worker)
		testsuite.IsDestroyed(t, watchDog)
	})
}
