package plugin

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"project/internal/anko"
)

// comFn is common function for Start, Stop, Name, Info and Status function.
type comFn = func(context.Context) (reflect.Value, reflect.Value)

// callFn is call function for Call function.
type callFn = func(context.Context, reflect.Value, ...interface{}) (reflect.Value, reflect.Value)

var (
	ErrAnkoPluginStarted = fmt.Errorf("anko plugin is started")
	ErrAnkoPluginStopped = fmt.Errorf("anko plugin is stopped")
)

// Anko is a plugin from anko script.
type Anko struct {
	external interface{}
	output   io.Writer
	stmt     anko.Stmt

	// in script
	startFn  comFn  // func() error
	stopFn   comFn  // func()
	nameFn   comFn  // func() string
	infoFn   comFn  // func() string
	statusFn comFn  // func() string
	callFn   callFn // func(name string, args ...interface{}) error

	env     *anko.Env
	stopped bool
	ctx     context.Context
	cancel  context.CancelFunc
	rwm     sync.RWMutex
}

// NewAnko is used to create a custom plugin from anko script.
func NewAnko(external interface{}, output io.Writer, script string) (*Anko, error) {
	stmt, err := anko.ParseSrc(script)
	if err != nil {
		return nil, err
	}
	ank := Anko{
		external: external,
		output:   output,
		stmt:     stmt,
		stopped:  true,
	}
	err = ank.load(context.Background())
	if err != nil {
		return nil, err
	}
	return &ank, nil
}

func (ank *Anko) load(ctx context.Context) error {
	// set output
	env := anko.NewEnv()
	env.SetOutput(ank.output)
	// load plugin
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	ret, err := anko.RunContext(ctx, env, ank.stmt)
	if err != nil {
		return err
	}
	// check is load successfully
	switch ret := ret.(type) {
	case bool:
		if !ret {
			return errors.New("return value is false")
		}
	default:
		return errors.Errorf("unexcepted return value: %s", ret)
	}
	// define external
	ext, err := env.Get("external")
	if err == nil {
		return errors.Errorf("already define external: %v", ext)
	}
	err = env.Define("external", ank.external)
	if err != nil {
		return errors.Wrap(err, "failed to define external")
	}
	// register functions
	err = ank.getExportedFunctions(env)
	if err != nil {
		return errors.WithMessage(err, "failed to register function")
	}
	ank.env = env
	return nil
}

func (ank *Anko) getExportedFunctions(env *anko.Env) error {
	start, err := env.Get("Start")
	if err != nil {
		return errors.Wrap(err, "failed to get start function")
	}
	startFn, ok := start.(comFn)
	if !ok {
		return errors.New("invalid start function type")
	}

	stop, err := env.Get("Stop")
	if err != nil {
		return errors.Wrap(err, "failed to get stop function")
	}
	stopFn, ok := stop.(comFn)
	if !ok {
		return errors.New("invalid stop function type")
	}

	name, err := env.Get("Name")
	if err != nil {
		return errors.Wrap(err, "failed to get name function")
	}
	nameFn, ok := name.(comFn)
	if !ok {
		return errors.New("invalid name function type")
	}

	info, err := env.Get("Info")
	if err != nil {
		return errors.Wrap(err, "failed to get info function")
	}
	infoFn, ok := info.(comFn)
	if !ok {
		return errors.New("invalid info function type")
	}

	status, err := env.Get("Status")
	if err != nil {
		return errors.Wrap(err, "failed to get status function")
	}
	statusFn, ok := status.(comFn)
	if !ok {
		return errors.New("invalid status function type")
	}

	call, err := env.Get("Call")
	if err != nil {
		return errors.Wrap(err, "failed to get call function")
	}
	callFn, ok := call.(callFn)
	if !ok {
		return errors.New("invalid call function type")
	}

	ank.startFn = startFn
	ank.stopFn = stopFn
	ank.nameFn = nameFn
	ank.infoFn = infoFn
	ank.statusFn = statusFn
	ank.callFn = callFn
	return nil
}

// Start is used to start plugin like connect external program.
func (ank *Anko) Start() error {
	ank.rwm.Lock()
	defer ank.rwm.Unlock()
	return ank.start()
}

func (ank *Anko) start() error {
	if !ank.stopped {
		return errors.WithStack(ErrAnkoPluginStarted)
	}
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	startErr, ankoErr := ank.startFn(ctx)
	if err, ok := ankoErr.Interface().(error); ok && err != nil {
		return errors.Wrap(err, "appear error when execute script about start")
	}
	if err, ok := startErr.Interface().(error); ok && err != nil {
		return errors.Wrap(err, "appear error when execute script about start")
	}
	ank.stopped = false
	ank.ctx, ank.cancel = context.WithCancel(context.Background())
	return nil
}

// Stop is used to stop plugin and stop all tasks like port scan.
func (ank *Anko) Stop() {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	err := ank.stop()
	if err != nil {
		// log it
	}

}

func (ank *Anko) stop() error {
	if ank.stopped {
		return nil
	}

	// stop other call
	ank.cancel()
	// call stop
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	stopErr, ankoErr := ank.stopFn(ctx)
	err := ankoErr.Interface().(error)
	if err != nil {
		return errors.Wrap(err, "appear error when execute script about stop")
	}
	err = stopErr.Interface().(error)
	if err != nil {
		return err
	}
	return nil
}

// Restart will stop plugin and then start plugin.
func (ank *Anko) Restart() error {
	return nil
}

// Name is used to get plugin name.
func (ank *Anko) Name() string {
	return ""
}

// Info is used to get plugin information.
func (ank *Anko) Info() string {
	return ""
}

// Status is used to get plugin status.
func (ank *Anko) Status() string {
	return ""
}

// Call is used to call plugin inner function or other special function.
func (ank *Anko) Call(name string, args ...interface{}) error {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()

	if ank.stopped {
		return errors.WithStack(ErrAnkoPluginStopped)
	}
	callErr, ankoErr := ank.callFn(ank.ctx, reflect.ValueOf(name), args...)
	if err, ok := ankoErr.Interface().(error); ok && err != nil {
		return errors.Wrap(err, "appear error when execute script about call")
	}
	if err, ok := callErr.Interface().(error); ok && err != nil {
		return errors.WithStack(err)
	}
	return nil
}
