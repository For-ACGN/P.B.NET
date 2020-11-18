package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const testTaskName = "mock task"

type mockTask struct {
	Pause       bool
	PrepareErr  bool
	PrepareSlow bool
	ProcessSlow bool

	progress float64
	detail   string
	rwm      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

func testNewMockTask() *mockTask {
	task := mockTask{}
	task.ctx, task.cancel = context.WithCancel(context.Background())
	return &task
}

func (mt *mockTask) Prepare(context.Context) error {
	if mt.PrepareErr {
		return errors.New("mock task prepare error")
	}
	if mt.PrepareSlow {
		// select {
		// case <-time.After(3 * time.Second):
		// case <-ctx.Done():
		// 	return ctx.Err()
		// }
		time.Sleep(3 * time.Second)
	}
	go mt.watcher()
	return nil
}

func (mt *mockTask) Process(ctx context.Context, task *Task) error {
	if task.Canceled() {
		return ctx.Err()
	}

	if mt.ProcessSlow {
		time.Sleep(3 * time.Second)
		return nil
	}

	// do something
	for i := 0; i < 5; i++ {
		// if task is paused, it will block here
		// if task canceled return process at once.
		if task.Canceled() {
			return ctx.Err()
		}

		// self call Pause
		if mt.Pause && i == 3 {
			task.Pause()

			// UI block, wait user interact
			time.Sleep(time.Second)

			task.Continue()
		}

		// do
		select {
		case <-time.After(200 * time.Millisecond):
			fmt.Printf("process %d\n", i)
			mt.updateProgress()
			mt.updateDetail(fmt.Sprintf("mock task detail: %d", i))
		case <-mt.ctx.Done():
			return mt.ctx.Err()
		}
	}
	return nil
}

func (mt *mockTask) Progress() string {
	mt.rwm.RLock()
	defer mt.rwm.RUnlock()
	return strconv.FormatFloat(mt.progress, 'f', -1, 64)
}

func (mt *mockTask) updateProgress() {
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	mt.progress += 0.2
}

func (mt *mockTask) Detail() string {
	mt.rwm.RLock()
	defer mt.rwm.RUnlock()
	return mt.detail
}

func (mt *mockTask) updateDetail(detail string) {
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	mt.detail = detail
}

func (mt *mockTask) watcher() {
	for {
		select {
		case <-time.After(time.Second):
			fmt.Println("watcher is alive")
		case <-mt.ctx.Done():
			return
		}
	}
}

func (mt *mockTask) Clean() {
	mt.cancel()
}

func TestTask(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	mt := testNewMockTask()
	mt.Pause = true
	task := New(testTaskName, mt, nil)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)

		task.Pause()

		time.Sleep(time.Second)

		task.Continue()

		t.Log("name:", task.Name())
		t.Log("state", task.State())
		t.Log("progress:", task.Progress())
		t.Log("detail", task.Detail())
		// prevent data race, not call "task.Task()"
		t.Log("task:", task.Task().(*mockTask).Progress())
	}()

	err := task.Start()
	require.NoError(t, err)

	wg.Wait()

	testsuite.IsDestroyed(t, task)
	testsuite.IsDestroyed(t, mt)
}

func TestTask_Start(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		err := task.Start()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("cancel before start", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		task.Cancel()

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("failed to prepare", func(t *testing.T) {
		mt := testNewMockTask()
		mt.PrepareErr = true
		task := New(testTaskName, mt, nil)

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("cancel before checkProcess", func(t *testing.T) {
		mt := testNewMockTask()
		mt.PrepareSlow = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(time.Second)
			task.Cancel()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("panic in checkStart", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		// set invalid state
		task.fsm.SetState(StateCancel)

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("panic in checkProcess", func(t *testing.T) {
		mt := testNewMockTask()
		mt.PrepareSlow = true
		task := New(testTaskName, mt, nil)

		// set invalid state
		go func() {
			time.Sleep(time.Second)
			task.fsm.SetState(StateCancel)
		}()

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("cancel before event complete", func(t *testing.T) {
		mt := testNewMockTask()
		mt.ProcessSlow = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(time.Second)
			task.Cancel()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("invalid pState", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		// set invalid state
		go func() {
			time.Sleep(100 * time.Millisecond)
			atomic.StoreInt32(task.state, pStateReady)
		}()

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("panic in process", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		// set invalid state
		go func() {
			time.Sleep(300 * time.Millisecond)
			task.fsm.SetState(StateCancel)
		}()

		err := task.Start()
		require.Error(t, err)

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}

func TestTask_Pause(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			time.Sleep(time.Second)

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("pause after cancel", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Cancel()

			task.Pause()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("pause twice", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()
			task.Pause()

			time.Sleep(time.Second)

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("pause in prepare", func(t *testing.T) {
		mt := testNewMockTask()
		mt.PrepareSlow = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			time.Sleep(4 * time.Second)

			task.Pause()

			time.Sleep(time.Second)

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("event failed", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			// mock
			err := task.fsm.Event(EventComplete)
			require.NoError(t, err)

			defer testsuite.DeferForPanic(t)
			task.Pause()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}

func TestTask_Continue(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			time.Sleep(time.Second)

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("continue after cancel", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Cancel()

			task.Continue()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("continue twice", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			time.Sleep(100 * time.Millisecond)

			task.Continue()

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("continue in prepare", func(t *testing.T) {
		mt := testNewMockTask()
		mt.PrepareSlow = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Continue()

			time.Sleep(4 * time.Second)

			task.Pause()

			time.Sleep(time.Second)

			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}

func TestTask_Cancel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(600 * time.Millisecond)
			task.Cancel()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("cancel after complete", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		err := task.Start()
		require.NoError(t, err)

		task.Cancel()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}

func TestTask_Paused(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("canceled", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			// make sure select ctx
			time.Sleep(300 * time.Millisecond)
			task.cancel()

			time.Sleep(300 * time.Millisecond)
			task.Cancel()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("continue and finish", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			// wait task run Paused()
			time.Sleep(time.Second)

			// mock finish
			atomic.StoreInt32(task.state, pStateProcess)

			// mock continue
			task.pauseCh <- struct{}{}
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("event failed", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			task.Pause()

			// mock
			task.fsm.SetState(StateComplete)

			task.Continue()
		}()

		err := task.Start()
		require.Error(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}

func TestTask_Pause_Continue(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("brute", func(t *testing.T) {
		mt := testNewMockTask()
		mt.Pause = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			for i := 0; i < 100; i++ {
				task.Pause()

				task.Continue()
			}
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("brute without pause", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			for i := 0; i < 100; i++ {
				task.Pause()

				task.Continue()
			}
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("brute with continue", func(t *testing.T) {
		mt := testNewMockTask()
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			for i := 0; i < 100; i++ {
				task.Pause()

				task.Continue()
			}

			time.Sleep(3 * time.Second)
			task.Continue()
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("a lot of operations", func(t *testing.T) {
		mt := testNewMockTask()
		mt.Pause = true
		task := New(testTaskName, mt, nil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			for i := 0; i < 100; i++ {
				task.Pause()

				time.Sleep(20 * time.Millisecond)

				task.Continue()
			}
		}()

		err := task.Start()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})

	t.Run("pause continue after complete", func(t *testing.T) {
		mt := testNewMockTask()
		mt.Pause = true
		task := New(testTaskName, mt, nil)

		err := task.Start()
		require.NoError(t, err)

		for i := 0; i < 100; i++ {
			task.Pause()

			task.Continue()
		}

		testsuite.IsDestroyed(t, task)
		testsuite.IsDestroyed(t, mt)
	})
}
