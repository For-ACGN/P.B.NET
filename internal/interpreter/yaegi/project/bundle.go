// Code generated by script/code/yaegi/project_test.go. DO NOT EDIT.

package project

import (
	"reflect"

	"project/internal/module"
)

// Symbols stores the map of unsafe package symbols.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
	initProjectInternalModule()
}

func initProjectInternalModule() {
	Symbols["project/internal/module"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewManager": reflect.ValueOf(module.NewManager),

		// type definitions
		"Manager": reflect.ValueOf((*module.Manager)(nil)),
		"Method":  reflect.ValueOf((*module.Method)(nil)),
		"Module":  reflect.ValueOf((*module.Module)(nil)),
		"Value":   reflect.ValueOf((*module.Value)(nil)),

		// interface wrapper definitions
		"_Module": reflect.ValueOf((*_project_internal_module_Module)(nil)),
	}
}

// _project_internal_module_Module is an interface wrapper for Module type
type _project_internal_module_Module struct {
	WCall        func(method string, args ...interface{}) (interface{}, error)
	WDescription func() string
	WInfo        func() string
	WIsStarted   func() bool
	WMethods     func() []string
	WName        func() string
	WRestart     func() error
	WStart       func() error
	WStatus      func() string
	WStop        func()
}

func (W _project_internal_module_Module) Call(method string, args ...interface{}) (interface{}, error) {
	return W.WCall(method, args...)
}
func (W _project_internal_module_Module) Description() string { return W.WDescription() }
func (W _project_internal_module_Module) Info() string        { return W.WInfo() }
func (W _project_internal_module_Module) IsStarted() bool     { return W.WIsStarted() }
func (W _project_internal_module_Module) Methods() []string   { return W.WMethods() }
func (W _project_internal_module_Module) Name() string        { return W.WName() }
func (W _project_internal_module_Module) Restart() error      { return W.WRestart() }
func (W _project_internal_module_Module) Start() error        { return W.WStart() }
func (W _project_internal_module_Module) Status() string      { return W.WStatus() }
func (W _project_internal_module_Module) Stop()               { W.WStop() }
