package env

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	// Packages is a where packages can be stored so VM import command can be
	// used to import them. reflect.Value must be valid or VM may crash.
	// For nil must use NilValue.
	Packages = make(map[string]map[string]reflect.Value)

	// PackageTypes is a where package types can be stored so VM import command
	// can be used to import them. reflect.Type must be valid or VM may crash.
	// For nil type must use NilType.
	PackageTypes = make(map[string]map[string]reflect.Type)

	// NilType is the reflect.type of nil.
	NilType = reflect.TypeOf(nil)

	// NilValue is the reflect.value of nil.
	NilValue = reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem()

	// ErrSymbolContainsDot symbol contains .
	ErrSymbolContainsDot = errors.New("symbol contains \".\"")
)

// ExternalLookup for Env external lookup of values and types.
type ExternalLookup interface {
	Get(string) (reflect.Value, error)
	Type(string) (reflect.Type, error)
}

// Env is the environment needed for a VM to run in.
type Env struct {
	parent    *Env
	extLookup ExternalLookup

	values map[string]reflect.Value
	types  map[string]reflect.Type
	rwm    *sync.RWMutex
}

// NewEnv creates new global scope.
func NewEnv() *Env {
	return &Env{
		values: make(map[string]reflect.Value),
		rwm:    new(sync.RWMutex),
	}
}

// SetExternalLookup sets an external lookup.
func (e *Env) SetExternalLookup(externalLookup ExternalLookup) {
	e.extLookup = externalLookup
}

// String returns string of values and types in current scope.
func (e *Env) String() string {
	var buffer bytes.Buffer
	e.rwm.RLock()
	defer e.rwm.RUnlock()
	if e.parent == nil {
		buffer.WriteString("No parent\n")
	} else {
		buffer.WriteString("Has parent\n")
	}
	for symbol, value := range e.values {
		buffer.WriteString(fmt.Sprintf("%v = %#v\n", symbol, value))
	}
	for symbol, aType := range e.types {
		buffer.WriteString(fmt.Sprintf("%v = %v\n", symbol, aType))
	}
	return buffer.String()
}

// NewEnv creates new child scope.
func (e *Env) NewEnv() *Env {
	return &Env{
		parent: e,
		values: make(map[string]reflect.Value),
		rwm:    new(sync.RWMutex),
	}
}

// NewModule creates new child scope and define it as a symbol.
// This is a shortcut for calling e.NewEnv then Define that new Env.
func (e *Env) NewModule(symbol string) (*Env, error) {
	module := &Env{
		parent: e,
		values: make(map[string]reflect.Value),
		rwm:    new(sync.RWMutex),
	}
	return module, e.Define(symbol, module)
}

// GetEnvFromPath returns Env from path.
func (e *Env) GetEnvFromPath(path []string) (*Env, error) {
	if len(path) < 1 {
		return e, nil
	}
	var (
		value reflect.Value
		ok    bool
	)
	// find starting env
	env := e
	for {
		func() {
			env.rwm.RLock()
			defer env.rwm.RUnlock()
			value, ok = env.values[path[0]]
		}()
		if ok {
			env, ok = value.Interface().(*Env)
			if ok {
				break
			}
		}
		if env.parent == nil {
			return nil, fmt.Errorf("no namespace called: %v", path[0])
		}
		env = env.parent
	}
	// find child env
	env.rwm.RLock()
	defer env.rwm.RUnlock()
	for i := 1; i < len(path); i++ {
		value, ok = env.values[path[i]]
		if ok {
			env, ok = value.Interface().(*Env)
			if ok {
				continue
			}
		}
		return nil, fmt.Errorf("no namespace called: %v", path[i])
	}
	return env, nil
}

// Copy the Env for current scope.
func (e *Env) Copy() *Env {
	e.rwm.RLock()
	defer e.rwm.RUnlock()
	env := Env{
		parent:    e.parent,
		extLookup: e.extLookup,
		values:    make(map[string]reflect.Value, len(e.values)),
		rwm:       new(sync.RWMutex),
	}
	for name, value := range e.values {
		env.values[name] = value
	}
	if e.types != nil {
		env.types = make(map[string]reflect.Type, len(e.types))
		for name, t := range e.types {
			env.types[name] = t
		}
	}
	return &env
}

// DeepCopy the Env for current scope and parent scopes.
// Note that each scope is a consistent snapshot but not the whole.
func (e *Env) DeepCopy() *Env {
	env := e.Copy()
	if env.parent != nil {
		env.parent = env.parent.DeepCopy()
	}
	return env
}
