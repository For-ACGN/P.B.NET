package task

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/looplab/fsm"
	"github.com/pkg/errors"

	"project/internal/xpanic"
)

// states about task (about FSM)
const (
	StateReady    = "ready"    // wait call Start()
	StatePrepare  = "prepare"  // prepare task like walk directory.
	StateProcess  = "process"  // doing task
	StatePause    = "pause"    // appear error or user pause progress
	StateComplete = "complete" // task finished
	StateCancel   = "cancel"   // task canceled
)

// events about task
const (
	EventStart    = "start"    // task start
	EventProcess  = "process"  // update progress
	EventPause    = "pause"    // pause update process progress
	EventContinue = "continue" // continue update process progress
	EventComplete = "complete" // task completed
	EventCancel   = "cancel"   // task canceled not update progress
)

// task internal state(different from state in FSM), "p" means private.
const (
	pStateReady int32 = iota
	pStateProcess
	pStatePause
	pStateFinish // cancel or complete
)

// Interface is the interface about task.
type Interface interface {
	Prepare(ctx context.Context) error
	Process(ctx context.Context, task *Task) error
	Progress() string // must be thread safe, usually provided like "19.99%"
	Detail() string   // must be thread safe
	Clean()           // release task internal resource
}

// Task is a task that contains all information about special task.
// It provide Pause task and get task progress.
type Task struct {
	name string
	task Interface
	fsm  *fsm.FSM

	// about control task
	state   *int32
	pauseCh chan struct{}
	mu      sync.Mutex

	startOnce  sync.Once
	cancelOnce sync.Once
	cleanOnce  sync.Once
	ctx        context.Context
	cancel     context.CancelFunc
}

// New is used to create a task with callbacks, ctx can cancel task.
func New(name string, iface Interface, callbacks fsm.Callbacks) *Task {
	// initial FSM
	cancelEvent := fsm.EventDesc{
		Name: EventCancel,
		Src:  []string{StateReady, StatePrepare, StateProcess, StatePause},
		Dst:  StateCancel,
	}
	events := []fsm.EventDesc{
		{Name: EventStart, Src: []string{StateReady}, Dst: StatePrepare},
		{Name: EventProcess, Src: []string{StatePrepare}, Dst: StateProcess},
		{Name: EventPause, Src: []string{StateProcess}, Dst: StatePause},
		{Name: EventContinue, Src: []string{StatePause}, Dst: StateProcess},
		{Name: EventComplete, Src: []string{StateProcess}, Dst: StateComplete},
		cancelEvent,
	}
	FSM := fsm.NewFSM(StateReady, events, callbacks)
	// create task
	task := Task{
		name:    name,
		task:    iface,
		fsm:     FSM,
		state:   new(int32),
		pauseCh: make(chan struct{}, 1),
	}
	task.ctx, task.cancel = context.WithCancel(context.Background())
	return &task
}

// Start is used to start current task.
func (task *Task) Start() (err error) {
	task.startOnce.Do(func() {
		defer func() {
			if err != nil {
				task.Cancel()
			} else {
				task.clean()
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				buf := xpanic.Log(r, "Task.Start")
				err = errors.New(buf.String())
			}
		}()
		if !task.checkStart() {
			err = errors.New("task canceled")
			return
		}
		err = task.prepare()
		if err != nil {
			return
		}
		if !task.checkProcess() {
			err = errors.New("task canceled")
			return
		}
		err = task.process()
	})
	return
}

func (task *Task) checkStart() bool {
	task.mu.Lock()
	defer task.mu.Unlock()
	if atomic.LoadInt32(task.state) != pStateReady {
		return false
	}
	err := task.fsm.Event(EventStart)
	if err != nil {
		internalErr(err)
	}
	return true
}

func (task *Task) prepare() error {
	err := task.task.Prepare(task.ctx)
	if err != nil {
		return errors.WithMessage(err, "failed to prepare task")
	}
	return nil
}

func (task *Task) checkProcess() bool {
	task.mu.Lock()
	defer task.mu.Unlock()
	if atomic.LoadInt32(task.state) != pStateReady {
		return false
	}
	err := task.fsm.Event(EventProcess)
	if err != nil {
		internalErr(err)
	}
	atomic.StoreInt32(task.state, pStateProcess)
	return true
}

func (task *Task) process() error {
	err := task.task.Process(task.ctx, task)
	if err != nil {
		return err
	}
	// maybe call Paused() between Process and mu.Lock()
	task.mu.Lock()
	defer task.mu.Unlock()
	switch state := atomic.LoadInt32(task.state); state {
	case pStateProcess:
	case pStatePause: // if paused, we need "continue"
		task.fsm.SetState(StateProcess)
	case pStateFinish:
		return errors.New("task canceled")
	default:
		panic(fmt.Sprintf("task: invalid pState: %d", state))
	}
	err = task.fsm.Event(EventComplete)
	if err != nil {
		internalErr(err)
	}
	atomic.StoreInt32(task.state, pStateFinish)
	return nil
}

// clean will be call once after task finish(include Cancel before Start)
func (task *Task) clean() {
	task.cleanOnce.Do(task.task.Clean)
}

// Pause is used to pause process progress.
func (task *Task) Pause() {
	task.mu.Lock()
	defer task.mu.Unlock()
	if atomic.LoadInt32(task.state) != pStateProcess {
		return
	}
	err := task.fsm.Event(EventPause)
	if err != nil {
		internalErr(err)
	}
	atomic.StoreInt32(task.state, pStatePause)
}

// Continue is used to continue current task.
func (task *Task) Continue() {
	task.mu.Lock()
	defer task.mu.Unlock()
	if atomic.LoadInt32(task.state) != pStatePause {
		return
	}
	select {
	case task.pauseCh <- struct{}{}:
	default:
	}
}

// Cancel is used to cancel current task.
func (task *Task) Cancel() {
	task.cancelOnce.Do(func() {
		defer task.clean()
		defer func() {
			if r := recover(); r != nil {
				xpanic.Log(r, "Task.Cancel")
			}
		}()
		task.mu.Lock()
		defer task.mu.Unlock()
		if atomic.LoadInt32(task.state) == pStateFinish {
			return
		}
		close(task.pauseCh)
		task.cancel()
		err := task.fsm.Event(EventCancel)
		if err != nil {
			internalErr(err)
		}
		atomic.StoreInt32(task.state, pStateFinish)
	})
}

// Paused is used to check current task is paused in process function.
func (task *Task) Paused() {
	if atomic.LoadInt32(task.state) != pStatePause {
		return
	}
	// wait continue signal
	select {
	case <-task.pauseCh:
	case <-task.ctx.Done():
		return
	}
	// set process state
	task.mu.Lock()
	defer task.mu.Unlock()
	if atomic.LoadInt32(task.state) != pStatePause {
		return
	}
	err := task.fsm.Event(EventContinue)
	if err != nil {
		internalErr(err)
	}
	atomic.StoreInt32(task.state, pStateProcess)
}

// Canceled is used to check current task is canceled.
// if Task paused it will block until continue or cancel.
func (task *Task) Canceled() bool {
	task.Paused()
	select {
	case <-task.ctx.Done():
		return true
	default:
		return false
	}
}

// Name is used to get the name of current task.
func (task *Task) Name() string {
	return task.name
}

// State is used to get the state about current task.
func (task *Task) State() string {
	return task.fsm.Current()
}

// Progress is used to get the progress about current task.
func (task *Task) Progress() string {
	return task.task.Progress()
}

// Detail is used to get the detail about current task.
func (task *Task) Detail() string {
	return task.task.Detail()
}

// Task is used to get raw task Interface.
func (task *Task) Task() Interface {
	return task.task
}

func internalErr(err error) {
	panic(fmt.Sprintf("task: internal error: %s", err))
}
