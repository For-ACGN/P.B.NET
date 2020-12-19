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
	return &Brute{logger: logger}
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
	// initialize
	brute.taskID = 0
	brute.tasks = make(map[int]*task, 1)
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
	switch method {
	case "pause":

	case "continue":

	case "kill":

	}
	return nil, nil
}
