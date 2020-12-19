package brute

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"project/internal/logger"
)

// Brute is the generic brute force attacker.
// It is used to build brute module easily.
type Brute struct {
	logger logger.Logger

	// contain brute tasks
	taskID   int // auto-increment
	tasks    map[int]*task
	tasksRWM sync.RWMutex

	started bool
	rwm     sync.RWMutex
	wg      sync.WaitGroup
}

// New is used to create a new common brute module.
func New(logger logger.Logger) *Brute {
	return &Brute{
		logger: logger,
		tasks:  make(map[int]*task, 1),
	}
}

// Start is used to start scanner, it will reset task ID.
func (brute *Brute) Start() error {
	brute.rwm.Lock()
	defer brute.rwm.Unlock()
	return brute.start()
}

func (brute *Brute) start() error {
	if brute.started {
		return errors.New("brute module is started")
	}
	// reset task id
	brute.taskID = 0
	brute.started = true
	return nil
}

// Stop is used to stop scanner, it will kill all tasks.
func (brute *Brute) Stop() {
	brute.rwm.Lock()
	defer brute.rwm.Unlock()
	brute.stop()
	brute.wg.Wait()
}

func (brute *Brute) stop() {
	if !brute.started {
		return
	}
	// kill all tasks
	brute.tasksRWM.RLock()
	defer brute.tasksRWM.RUnlock()
	for _, task := range brute.tasks {
		task.Kill()
	}
	brute.started = false
}

// Restart is used to restart brute.
func (brute *Brute) Restart() error {
	brute.rwm.Lock()
	defer brute.rwm.Unlock()
	brute.stop()
	brute.wg.Wait()
	return brute.start()
}

// IsStarted is used to check brute is started.
func (brute *Brute) IsStarted() bool {
	brute.rwm.RLock()
	defer brute.rwm.RUnlock()
	return brute.started
}

// Info is used to get brute module information.
func (brute *Brute) Info() string {
	brute.tasksRWM.RLock()
	defer brute.tasksRWM.RUnlock()
	return fmt.Sprintf("total number of tasks run: %d", brute.taskID)
}

// Status is used to get brute module status.
func (brute *Brute) Status() string {
	brute.tasksRWM.RLock()
	defer brute.tasksRWM.RUnlock()
	return fmt.Sprintf("running task: %d", len(brute.tasks))
}

// Methods is used to get brute module methods.
func (brute *Brute) Methods() []string {
	return nil
}

// Call is used to call brute module extended method.
func (brute *Brute) Call(method string, args ...interface{}) (interface{}, error) {
	// check arguments first
	if len(args) != 1 {
		return nil, errors.New("invalid argument number")
	}
	id, ok := args[0].(int)
	if !ok {
		return nil, errors.New("argument 1 is not a int")
	}
	switch method {
	case "pause":
		return brute.PauseTask(id), nil
	case "continue":
		return brute.ContinueTask(id), nil
	case "kill":
		return brute.KillTask(id), nil
	default:
		return nil, errors.Errorf("unknown method: \"%s\"", method)
	}
}

// GetTask is used to get task by ID.
func (brute *Brute) GetTask(id int) (*task, error) {
	brute.tasksRWM.RLock()
	defer brute.tasksRWM.RUnlock()
	task, ok := brute.tasks[id]
	if !ok {
		return nil, errors.Errorf("task %d is not exist", id)
	}
	return task, nil
}

// PauseTask is used to pause task by ID.
func (brute *Brute) PauseTask(id int) error {
	task, err := brute.GetTask(id)
	if err != nil {
		return err
	}
	task.Pause()
	return nil
}

// ContinueTask is used to continue task by ID.
func (brute *Brute) ContinueTask(id int) error {
	task, err := brute.GetTask(id)
	if err != nil {
		return err
	}
	task.Continue()
	return nil
}

// KillTask is used to Kill task by ID.
func (brute *Brute) KillTask(id int) error {
	task, err := brute.GetTask(id)
	if err != nil {
		return err
	}
	task.Kill()
	return nil
}
