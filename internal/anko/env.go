package anko

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
)

// runtime is used to prevent loop reference and easy clean reference.
type runtime struct {
	// store values
	values    map[string]reflect.Value
	valuesRWM sync.RWMutex

	// store types
	types    map[string]reflect.Type
	typesRWM sync.RWMutex
}

func newRuntime(e *Env) *runtime {
	rt := runtime{
		values: make(map[string]reflect.Value),
		types:  make(map[string]reflect.Type),
	}
	// define built in function that use Env.
	for _, item := range []*struct {
		symbol string
		fn     interface{}
	}{
		{"printf", e.printf},
		{"print", e.print},
		{"println", e.println},
		{"eval", e.eval},
	} {
		err := rt.DefineValue(item.symbol, item.fn)
		if err != nil {
			panic(fmt.Sprintf("anko: internal error: %s", err))
		}
	}
	return &rt
}

func (rt *runtime) Get(symbol string) (reflect.Value, error) {
	rt.valuesRWM.RLock()
	defer rt.valuesRWM.RUnlock()
	if value, ok := rt.values[symbol]; ok {
		return value, nil
	}
	return reflect.Value{}, fmt.Errorf("value %q is not defined", symbol)
}

func (rt *runtime) DefineValue(symbol string, value interface{}) error {
	var reflectValue reflect.Value
	if value == nil {
		reflectValue = env.NilValue
	} else {
		var ok bool
		reflectValue, ok = value.(reflect.Value)
		if !ok {
			reflectValue = reflect.ValueOf(value)
		}
	}
	return rt.defineValue(symbol, reflectValue)
}

func (rt *runtime) defineValue(symbol string, value reflect.Value) error {
	if strings.Contains(symbol, ".") {
		return env.ErrSymbolContainsDot
	}
	rt.valuesRWM.Lock()
	defer rt.valuesRWM.Unlock()
	rt.values[symbol] = value
	return nil
}

func (rt *runtime) Type(symbol string) (reflect.Type, error) {
	rt.typesRWM.RLock()
	defer rt.typesRWM.RUnlock()
	if typ, ok := rt.types[symbol]; ok {
		return typ, nil
	}
	return nil, fmt.Errorf("type %q is not defined", symbol)
}

func (rt *runtime) DefineType(symbol string, typ interface{}) error {
	var reflectType reflect.Type
	if typ == nil {
		reflectType = env.NilType
	} else {
		var ok bool
		reflectType, ok = typ.(reflect.Type)
		if !ok {
			reflectType = reflect.TypeOf(typ)
		}
	}
	return rt.defineType(symbol, reflectType)
}

func (rt *runtime) defineType(symbol string, typ reflect.Type) error {
	if strings.Contains(symbol, ".") {
		return env.ErrSymbolContainsDot
	}
	rt.typesRWM.Lock()
	defer rt.typesRWM.Unlock()
	rt.types[symbol] = typ
	return nil
}

func (rt *runtime) Destroy() {
	rt.valuesRWM.Lock()
	defer rt.valuesRWM.Unlock()
	for symbol := range rt.values {
		delete(rt.values, symbol)
	}
	rt.typesRWM.Lock()
	defer rt.typesRWM.Unlock()
	for symbol := range rt.types {
		delete(rt.types, symbol)
	}
}

// Env is the environment needed for a VM to run in.
type Env struct {
	env     *env.Env
	runtime *runtime

	// for built in function about print
	output io.Writer

	// control eval
	ctx    context.Context
	cancel context.CancelFunc
}

func newEnv(e *env.Env, output io.Writer) *Env {
	core.ImportToX(e)
	defineConvert(e)
	defineCore(e)
	en := &Env{
		env:    e,
		output: output,
	}
	r := newRuntime(en)
	e.SetExternalLookup(r)
	en.runtime = r
	return en
}

func (e *Env) printf(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(e.output, format, v...)
}

func (e *Env) print(v ...interface{}) {
	_, _ = fmt.Fprint(e.output, v...)
}

func (e *Env) println(v ...interface{}) {
	_, _ = fmt.Fprintln(e.output, v...)
}

func (e *Env) eval(src string) (interface{}, error) {
	stmt, err := ParseSrc(src)
	if err != nil {
		return nil, err
	}
	ne := newEnv(e.env.NewEnv(), e.output)
	ne.ctx = e.ctx
	defer ne.Close()
	val, err := RunContext(e.ctx, ne, stmt)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Define is used to define out value to environment.
// It will redirect to runtime.
func (e *Env) Define(symbol string, value interface{}) error {
	return e.runtime.DefineValue(symbol, value)
}

// Get returns interface value from the scope where symbol is first found.
func (e *Env) Get(symbol string) (interface{}, error) {
	return e.env.Get(symbol)
}

// GetValue returns reflect value from the scope where symbol is first found.
func (e *Env) GetValue(symbol string) (reflect.Value, error) {
	return e.env.GetValue(symbol)
}

// DefineType will redirect to runtime.
func (e *Env) DefineType(symbol string, value interface{}) error {
	return e.runtime.DefineType(symbol, value)
}

// SetOutput is used to set output for printf, print and println.
func (e *Env) SetOutput(output io.Writer) {
	e.output = output
}

// Close is used to close env and delete functions that reference self.
func (e *Env) Close() {
	if e.cancel != nil {
		e.cancel()
	}
	destroyEnv(e.env)
	e.runtime.Destroy()
}

// destroyEnv is used to clean env, env.Env without functions about enumerate values and types.
func destroyEnv(e *env.Env) {
	scanner := bufio.NewScanner(strings.NewReader(e.String()))
	for scanner.Scan() {
		text := scanner.Text()
		index := strings.Index(text, " = ")
		if index == -1 {
			continue
		}
		symbol := text[:index]
		// clean inner env
		typ := text[index:]
		const flag = " = &env.Env"
		if len(typ) >= len(flag) && typ[:len(flag)] == flag {
			ie, _ := e.Get(symbol)
			if ie != nil {
				destroyEnv(ie.(*env.Env))
			}
		}
		e.Delete(symbol)
	}
}
