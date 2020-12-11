package plugin

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"

	"project/internal/interpreter/yaegi"
	"project/internal/security"
)

// Yaegi is a plugin from yaegi script.
type Yaegi struct {
	external interp.Exports
	output   io.Writer
	script   *security.String

	// in yaegi script
	nameFn    func() string
	descFn    func() string
	startFn   func() error
	stopFn    func()
	infoFn    func() string
	statusFn  func() string
	methodsFn func() []string
	callFn    func(method string, args ...interface{}) error

	ipt     *interp.Interpreter
	started bool
	rwm     sync.RWMutex
}

// NewYaegi is used to create a custom plugin from anko script.
func NewYaegi(external interface{}, output io.Writer, script string) (*Yaegi, error) {
	// make external to a package. import "external"
	pkg := make(interp.Exports)
	if external != nil {
		pkg["external"] = external.(map[string]reflect.Value)
	}
	yae := Yaegi{
		external: pkg,
		output:   output,
		script:   security.NewString(script),
	}
	err := yae.load()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to load plugin")
	}
	return &yae, nil
}

func (yae *Yaegi) load() error {
	ipt := interp.New(interp.Options{
		Stdout: yae.output,
		Stderr: yae.output,
	})
	ipt.Use(yaegi.Symbols)
	ipt.Use(yae.external)

	// load plugin
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	script := yae.script.Get()
	defer yae.script.Put(script)
	_, err := ipt.EvalWithContext(ctx, script)
	if err != nil {
		return err
	}
	return nil
}
