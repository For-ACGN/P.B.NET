package plugin

import (
	"context"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/anko"
)

// comFn is common function for Start, Stop, Name, Info and Status function.
type comFn func(context.Context) (reflect.Value, reflect.Value)

// callFn is call function for Call function.
type callFn func(context.Context, string, ...interface{}) (reflect.Value, reflect.Value)

// Anko is a plugin from anko script.
type Anko struct {
	external interface{}
	output   io.Writer
	stmt     anko.Stmt

	env *anko.Env

	// in script
	startFn comFn  // func() error
	stopFn  comFn  // func()
	name    comFn  // func() string
	info    comFn  // func() string
	status  comFn  // func() string
	call    callFn // func(name string, args ...interface{}) error

	mu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewAnko is used to create a custom module from anko script.
func NewAnko(external interface{}, output io.Writer, script string) (*Anko, error) {
	stmt, err := anko.ParseSrc(script)
	if err != nil {
		return nil, err
	}
	ank := Anko{
		external: external,
		output:   output,
		stmt:     stmt,
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
	// load module
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
	err = ank.getExportedFunctions(env)
	if err != nil {
		return err
	}
	return nil
}

func (ank *Anko) getExportedFunctions(env *anko.Env) error {
	start, err := env.Get("Start")
	if err != nil {
		return errors.Wrap(err, "failed to get start function")
	}
	startFn, ok := start.(comFn)
	if !ok {
		return errors.Wrap(err, "invalid start function type")
	}

	stop, err := env.Get("Stop")
	if err != nil {
		return errors.Wrap(err, "failed to get stop function")
	}
	stopFn, ok := stop.(comFn)
	if !ok {
		return errors.Wrap(err, "invalid stop function type")
	}

	name, err := env.Get("Name")
	if err != nil {
		return errors.Wrap(err, "failed to get name function")
	}
	nameFn, ok := name.(comFn)
	if !ok {
		return errors.Wrap(err, "invalid name function type")
	}

	info, err := env.Get("Info")
	if err != nil {
		return errors.Wrap(err, "failed to get info function")
	}
	infoFn, ok := info.(comFn)
	if !ok {
		return errors.Wrap(err, "invalid info function type")
	}

	status, err := env.Get("Status")
	if err != nil {
		return errors.Wrap(err, "failed to get status function")
	}
	statusFn, ok := status.(comFn)
	if !ok {
		return errors.Wrap(err, "invalid status function type")
	}

	call, err := env.Get("Call")
	if err != nil {
		return errors.Wrap(err, "failed to get call function")
	}
	callFn, ok := call.(callFn)
	if !ok {
		return errors.Wrap(err, "invalid call function type")
	}
	ank.startFn = startFn
	ank.stopFn = stopFn
	ank.name = nameFn
	ank.info = infoFn
	ank.status = statusFn
	ank.call = callFn
	return nil
}

// Start is used to start module like connect external program.
func (ank *Anko) Start() error {

	// ret1, e := startFn(ctx)
	// fmt.Println("ret2", ret2.Interface() == nil)
	// fmt.Println(ret1.Interface(), ret2.Type())
	// fmt.Println(&startFn)

	return ank.startFn()
}

// Stop is used to stop module and stop all tasks like port scan.
func (ank *Anko) Stop() {
	ank.stopFn()
}

// Restart will stop module and then start module.
func (ank *Anko) Restart() error {
	return ank.restart()
}

// Name is used to get module name.
func (ank *Anko) Name() string {
	return ank.name()
}

// Info is used to get module information.
func (ank *Anko) Info() string {
	return ank.info()
}

// Status is used to get module status.
func (ank *Anko) Status() string {
	return ank.status()
}

// Call is used to call module inner function or other special function.
func (ank *Anko) Call(name string, arg ...interface{}) error {

	return nil
}
