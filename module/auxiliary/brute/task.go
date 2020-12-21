package brute

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/mohae/deepcopy"

	"project/external/ranger"

	"project/internal/logger"
	"project/internal/task"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

// ErrInvalidCred is a error about login, if Login function returned
// error is not this, brute will stop current task.
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

type bruteTask struct {
	ctx *Brute

	cfg *TaskConfig

	// about progress, speed and detail
	current *big.Float
	total   *big.Float
	speed   uint64
	speeds  [10]uint64
	full    bool

	targets []string // current targets
	detail  string
	rwm     sync.RWMutex

	// control speed watcher
	stopSignal chan struct{}
	wg         sync.WaitGroup
}

func newBruteTask(ctx *Brute) *task.Task {
	bt := bruteTask{
		ctx:        ctx,
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New("brute", &bt, nil)
}

func (bt *bruteTask) Prepare(context.Context) error {
	cfg, err := bt.cfg.Apply()
	if err != nil {
		return err
	}
	bt.cfg = cfg

	return nil
}

func (bt *bruteTask) Process(ctx context.Context, task *task.Task) error {

	return nil
}

// Progress is used to
func (bt *bruteTask) Progress() string {
	return ""
}

// Detail is used to get detail about brute task.
//
// current target: 1.1.1.1:22, 1.1.1.2:22
func (bt *bruteTask) Detail() string {
	bt.rwm.RLock()
	defer bt.rwm.RUnlock()
	return bt.detail
}

func (bt *bruteTask) updateCurrentTargets() {
	bt.rwm.Lock()
	defer bt.rwm.Unlock()
	bt.detail = "current target: " + strings.Join(bt.targets, ", ")
}

func (bt *bruteTask) Clean() {

}

// Task is the brute task.
type Task struct {
	ctx *Brute

	pauser *pauser.Pauser

	context context.Context
	cancel  context.CancelFunc
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

// Progress is used to get the current task progress and speed.
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
