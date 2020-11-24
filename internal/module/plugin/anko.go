package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"project/internal/anko"
)

type Anko struct {
	env *anko.Env

	// in script
	start   func() error
	stop    func()
	restart func() error
	name    func() string
	info    func() string
	status  func() string
}

// NewAnko is used to create a custom module from anko script.
func NewAnko(ctx context.Context, external interface{}, script string) (*Anko, error) {
	ast, err := anko.ParseSrc(script)
	if err != nil {
		return nil, err
	}
	env := anko.NewEnv()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ret, err := anko.RunContext(ctx, env, ast)
	if err != nil {
		return nil, err
	}
	// check is load successfully
	switch ret := ret.(type) {
	case bool:
		if !ret {
			return nil, errors.New("return value is false")
		}
	default:
		return nil, errors.Errorf("unexcepted return value: %s", ret)
	}
	// define external
	ext, err := env.Get("external")
	if err == nil {
		return nil, errors.Errorf("already define external: %v", ext)
	}
	err = env.Define("external", external)
	if err != nil {
		return nil, errors.Wrap(err, "failed to define external")
	}

	// get exported function
	start, err := env.Get("Start")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get Start function")
	}
	startFn, ok := start.(func() error)
	if ok {
		return nil, errors.Wrap(err, "invalid Start function")
	}
	fmt.Println(startFn())

	return nil, nil
}

func (ank *Anko) Start() error {
	return ank.start()
}

func (ank *Anko) Stop() {
	ank.stop()
}

func (ank *Anko) Restart() error {
	return ank.restart()
}

func (ank *Anko) Name() string {
	return ank.name()
}

func (ank *Anko) Info() string {
	return ank.info()
}

func (ank *Anko) Status() string {
	return ank.status()
}
