package plugin

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"project/internal/anko"
)

// comFn is common function for Name, Description, Start, Stop, Info, Status and Methods functions.
type comFn = func(context.Context) (reflect.Value, reflect.Value)

// callFn is call function for Call function.
type callFn = func(context.Context, reflect.Value, ...interface{}) (reflect.Value, reflect.Value)

// Anko is a plugin from anko script.
//
// Don't cover "external" symbol, otherwise you can't use external functions.
// Use structure like Mutex in sync package, must use new(sync.Mutex), otherwise
// will appear data race warning.
type Anko struct {
	external interface{}
	output   io.Writer
	stmt     anko.Stmt

	// in anko script
	nameFn    comFn  // func() string
	descFn    comFn  // func() string
	startFn   comFn  // func() error
	stopFn    comFn  // func()
	infoFn    comFn  // func() string
	statusFn  comFn  // func() string
	methodsFn comFn  // func() []string
	callFn    callFn // func(method string, args ...interface{}) error

	env     *anko.Env
	started bool
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
	}
	err = ank.load()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to load plugin")
	}
	return &ank, nil
}

func (ank *Anko) load() error {
	// set output
	env := anko.NewEnv()
	env.SetOutput(ank.output)
	// load plugin
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
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
	err = ank.registerFunctions(env)
	if err != nil {
		return errors.WithMessage(err, "failed to register function")
	}
	ank.env = env
	return nil
}

func (ank *Anko) registerFunctions(env *anko.Env) error {
	// register common functions
	for _, item := range [...]*struct {
		method string
		field  *comFn
	}{
		{"Name", &ank.nameFn},
		{"Description", &ank.descFn},
		{"Start", &ank.startFn},
		{"Stop", &ank.stopFn},
		{"Info", &ank.infoFn},
		{"Status", &ank.statusFn},
		{"Methods", &ank.methodsFn},
	} {
		symbol, err := env.Get(item.method)
		if err != nil {
			return errors.Wrapf(err, "failed to get %s function", item.method)
		}
		fn, ok := symbol.(comFn)
		if !ok {
			return errors.Errorf("invalid %s function type", item.method)
		}
		*item.field = fn
	}
	// register Call function
	call, err := env.Get("Call")
	if err != nil {
		return errors.Wrap(err, "failed to get Call function")
	}
	callFn, ok := call.(callFn)
	if !ok {
		return errors.New("invalid Call function type")
	}
	ank.callFn = callFn
	return nil
}

// Name is used to get plugin name.
func (ank *Anko) Name() string {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	name, err := ank.name()
	if err != nil {
		return "[error]: " + err.Error()
	}
	return name
}

func (ank *Anko) name() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	name, ankoErr := ank.nameFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when get name: \"%s\" at line:%d column:%d"
		return "", errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return "", errors.Wrap(err, "failed to get name")
	default:
		return "", errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check return type
	switch name := name.Interface().(type) {
	case string:
		return name, nil
	default:
		return "", errors.Errorf("unexpected name type, value: %v", name)
	}
}

// Description is used to get plugin description.
func (ank *Anko) Description() string {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	desc, err := ank.description()
	if err != nil {
		return "[error]: " + err.Error()
	}
	return desc
}

func (ank *Anko) description() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	desc, ankoErr := ank.descFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when get description: \"%s\" at line:%d column:%d"
		return "", errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return "", errors.Wrap(err, "failed to get description")
	default:
		return "", errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check return type
	switch desc := desc.Interface().(type) {
	case string:
		return desc, nil
	default:
		return "", errors.Errorf("unexpected description type, value: %v", desc)
	}
}

// Start is used to start plugin like connect external program.
func (ank *Anko) Start() error {
	ank.rwm.Lock()
	defer ank.rwm.Unlock()
	return ank.start()
}

func (ank *Anko) start() error {
	if ank.started {
		return errors.Errorf("anko plugin %s is started", ank.Name())
	}
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	startErr, ankoErr := ank.startFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when start: \"%s\" at line:%d column:%d"
		return errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return errors.Wrap(err, "failed to start")
	default:
		return errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check start error
	switch err := startErr.Interface().(type) {
	case nil:
	case error:
		if err != nil {
			return errors.Wrap(err, "failed to start")
		}
	default:
		return errors.Errorf("unexpected start error type, value: %v", err)
	}
	// update module
	ank.ctx, ank.cancel = context.WithCancel(context.Background())
	ank.started = true
	return nil
}

// Stop is used to stop plugin and stop all tasks like port scan.
func (ank *Anko) Stop() {
	ank.rwm.Lock()
	defer ank.rwm.Unlock()
	err := ank.stop()
	if err != nil {
		errStr := "appear error when stop: " + err.Error()
		_, _ = ank.output.Write([]byte(errStr))
	}
}

func (ank *Anko) stop() error {
	if !ank.started {
		return nil
	}
	// stop other call
	ank.cancel()
	// call stop
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	stopErr, ankoErr := ank.stopFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when stop: \"%s\" at line:%d column:%d"
		return errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return errors.Wrap(err, "failed to stop")
	default:
		return errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check stop error
	switch err := stopErr.Interface().(type) {
	case nil:
	case error:
		if err != nil {
			return errors.Wrap(err, "failed to stop")
		}
	default:
		return errors.Errorf("unexpected stop error type, value: %v", err)
	}
	// update module
	ank.env.Close()
	ank.started = false
	return nil
}

// Restart will stop plugin and then start plugin.
func (ank *Anko) Restart() error {
	ank.rwm.Lock()
	defer ank.rwm.Unlock()
	// stop
	err := ank.stop()
	if err != nil {
		errStr := "appear error when restart: " + err.Error()
		_, _ = ank.output.Write([]byte(errStr))
	}
	// reload
	err = ank.load()
	if err != nil {
		return errors.WithMessage(err, "failed to reload plugin")
	}
	// start
	return ank.start()
}

// IsStarted is used to check this plugin is started.
func (ank *Anko) IsStarted() bool {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	return ank.started
}

// Info is used to get plugin information.
func (ank *Anko) Info() string {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	info, err := ank.info()
	if err != nil {
		return "[error]: " + err.Error()
	}
	return info
}

func (ank *Anko) info() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	info, ankoErr := ank.infoFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when get info: \"%s\" at line:%d column:%d"
		return "", errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return "", errors.Wrap(err, "failed to get info")
	default:
		return "", errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check return type
	switch info := info.Interface().(type) {
	case string:
		return info, nil
	default:
		return "", errors.Errorf("unexpected info type, value: %v", info)
	}
}

// Status is used to get plugin status.
func (ank *Anko) Status() string {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	status, err := ank.status()
	if err != nil {
		return "[error]: " + err.Error()
	}
	return status
}

func (ank *Anko) status() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	status, ankoErr := ank.statusFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when get status: \"%s\" at line:%d column:%d"
		return "", errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return "", errors.Wrap(err, "failed to get status")
	default:
		return "", errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check return type
	switch status := status.Interface().(type) {
	case string:
		return status, nil
	default:
		return "", errors.Errorf("unexpected status type, value: %v", status)
	}
}

// Methods is used to get the information about extended methods.
func (ank *Anko) Methods() []string {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	methods, err := ank.methods()
	if err != nil {
		return []string{"[error]: " + err.Error()}
	}
	return methods
}

func (ank *Anko) methods() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	methods, ankoErr := ank.methodsFn(ctx)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when get methods: \"%s\" at line:%d column:%d"
		return nil, errors.Errorf(format, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return nil, errors.Wrap(err, "failed to get methods")
	default:
		return nil, errors.Errorf("unexpected anko error type, value: %v", err)
	}
	// check return type
	switch methods := methods.Interface().(type) {
	case []string:
		return methods, nil
	default:
		return nil, errors.Errorf("unexpected methods type, value: %v", methods)
	}
}

// Call is used to call plugin inner function or other special function.
func (ank *Anko) Call(method string, args ...interface{}) (interface{}, error) {
	ank.rwm.RLock()
	defer ank.rwm.RUnlock()
	if !ank.started {
		return nil, errors.Errorf("anko plugin %s is stopped", ank.Name())
	}
	ret, ankoErr := ank.callFn(ank.ctx, reflect.ValueOf(method), args...)
	// check anko error
	switch err := ankoErr.Interface().(type) {
	case nil:
	case *anko.VMError:
		const format = "appear error when call %s: \"%s\" at line:%d column:%d"
		return nil, errors.Errorf(format, method, err.Message, err.Pos.Line, err.Pos.Column)
	case error:
		return nil, errors.Wrapf(err, "failed to call %s", method)
	default:
		return nil, errors.Errorf("unexpected anko error type, value: %v", err)
	}
	return ret.Interface(), nil
}
