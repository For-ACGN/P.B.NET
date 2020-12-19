package brute

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/mohae/deepcopy"

	"project/external/ranger"

	"project/internal/logger"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

// ErrInvalidCred is a error about login, if Login returned error
// is not this, brute will stop current task.
var ErrInvalidCred = fmt.Errorf("invalid username or password")

// TaskConfig contains brute task configuration.
// If service don't need username or password, brute instance can
// input a fake but not zero []string for pass Apply check, and
// in login callback don't use it.
type TaskConfig struct {
	Targets       []string      `range:"nonzero"`
	Username      []string      `range:"nonzero"`
	Password      []string      `range:"nonzero"`
	Worker        int           `range:"min=1" default:"4"`
	Timeout       time.Duration `range:"min=1" default:"30s"`
	Interval      time.Duration `range:"min=1" default:"10ms"`
	StopOnSuccess bool          `default:"true"`
}

// Apply is used to apply default value and check value range.
func (cfg *TaskConfig) Apply() (*TaskConfig, error) {
	cp := deepcopy.Copy(cfg).(*TaskConfig)
	err := defaults.Set(cp)
	if err != nil {
		return nil, err
	}
	err = ranger.Validate(cp)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// Task is the brute task.
type Task struct {
	pauser *pauser.Pauser

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Run is used to create a brute task and running it.
func (brute *Brute) Run(cfg *TaskConfig) {

}

// Pause is used to pause this brute task.
func (task *Task) Pause() {
	task.pauser.Pause()
}

// Continue is used to continue this brute task.
func (task *Task) Continue() {
	task.pauser.Continue()
}

// Kill is used to kill this brute task.
func (task *Task) Kill() {
	task.pauser.Close()
}

// Progress is used to get the current task progress.
// "15.22%|current/total|128 MB/s"
func (task *Task) Progress() {

}

func (brute *Brute) logf(lv logger.Level, format string, log ...interface{}) {
	brute.logger.Printf(lv, "worker", format, log...)
}

func (brute *Brute) log(lv logger.Level, log ...interface{}) {
	brute.logger.Println(lv, "worker", log...)
}

func (brute *Brute) worker(id int) {
	brute.logf(logger.Debug, "worker %d started", id)
	defer brute.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			brute.log(logger.Fatal, xpanic.Printf(r, "Scanner.worker-%d", id))
			// restart
			time.Sleep(time.Second)
			brute.wg.Add(1)
			go brute.worker(id)
		}
	}()
}
