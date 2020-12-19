package brute

import (
	"context"
	"fmt"
	"time"

	"github.com/creasty/defaults"
	"github.com/mohae/deepcopy"

	"project/external/ranger"
	"project/internal/logger"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

// Login is the callback to the brute instance, if login successfully
// it will return true, if appear error, brute will log it.
type Login func(ctx context.Context, target, username, password string) (bool, error)

// ErrInvalidCred is a error about login, if Login returned error
// is not this, brute will stop current task.
var ErrInvalidCred = fmt.Errorf("invalid username or password")

// Config contains brute configuration.
// If service don't need username or password, brute instance can
// input a fake but not zero []string for pass Apply check, and
// in login callback don't use it.
type Config struct {
	Username []string `range:"nonzero"`
	Password []string `range:"nonzero"`
	Worker   int      `range:"min=1" default:"4"`
	Timeout  time.Duration
	Proxy    func(ctx context.Context, network, address string)
}

// Apply is used to apply default value and check value range.
func (cfg *Config) Apply() (*Config, error) {
	cp := deepcopy.Copy(cfg).(*Config)
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

type task struct {
	pauser *pauser.Pauser
}

// Pause is used to pause this brute task.
func (task *task) Pause() {
	task.pauser.Pause()
}

// Continue is used to continue this brute task.
func (task *task) Continue() {
	task.pauser.Continue()
}

// Kill is used to kill this brute task.
func (task *task) Kill() {

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
