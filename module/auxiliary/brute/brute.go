package brute

import (
	"context"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"

	"project/external/ranger"

	"project/internal/logger"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

// Login is the callback to the brute instance, if login successfully
// it will return true, if appear error, brute will log it.
type Login func(ctx context.Context, target interface{}, username, password string) (ok bool, err error)

// Config contains brute configuration.
// If service don't need username or password, brute instance can
// input a fake but not zero []string for pass Apply check, and
// in login callback don't use it.
type Config struct {
	Username []string `range:"nonzero"`
	Password []string `range:"nonzero"`
	Worker   int      `range:"min=1" default:"4"`
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

// Brute is the generic brute force attacker.
// It is used to build brute module easily.
type Brute struct {
	logger logger.Logger
	cfg    *Config

	// for control operations
	pause   *pauser.Pauser
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	rwm     sync.RWMutex
	wg      sync.WaitGroup
}

// New is used to create a new Brute.
func New(logger logger.Logger, cfg *Config) (*Brute, error) {
	cfg, err := cfg.Apply()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Start is used to start scanner, it will start to process jobs.
func (brute *Brute) Start() error {
	brute.rwm.Lock()
	defer brute.rwm.Unlock()
	return brute.start()
}

func (brute *Brute) start() error {
	if brute.started {
		return errors.New("brute is started")
	}
	brute.ctx, brute.cancel = context.WithCancel(context.Background())
	for i := 0; i < brute.cfg.Worker; i++ {
		brute.wg.Add(1)
		go brute.worker(i)
	}
	brute.started = true
	return nil
}

// Stop is used to stop scanner, it will kill all processing jobs.
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
	brute.cancel()
	// prevent panic before here
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

// Pause is used to pause brute.
func (brute *Brute) Pause() {
	brute.pause.Pause()
}

// Continue is used to continue brute.
func (brute *Brute) Continue() {
	brute.pause.Continue()
}

func (brute *Brute) logf(lv logger.Level, format string, log ...interface{}) {
	brute.logger.Printf(lv, "brute", format, log...)
}

func (brute *Brute) log(lv logger.Level, log ...interface{}) {
	brute.logger.Println(lv, "brute", log...)
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
