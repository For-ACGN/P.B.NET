package yaegi

import (
	"go/build"
	"reflect"

	"github.com/traefik/yaegi/interp"
)

// Interpreter is a wrapper about yaegi interpreter, but add compiler to merge
// source files to a single file, Process unsafe.Offsetof() and remote importer.
type Interpreter struct {
	yip *interp.Interpreter

	packages map[string]struct{}
}

// NewInterpreter is used to create a new interpreter.
func NewInterpreter() *Interpreter {
	i := Interpreter{yip: interp.New(interp.Options{})}
	return &i
}

// EvalPath is
func (ip *Interpreter) EvalPath(path string) (reflect.Value, error) {
	// CompileWithContext()

	return reflect.Value{}, nil
}

// EvalFiles is
func (ip *Interpreter) EvalFiles(files []string) (reflect.Value, error) {
	return reflect.Value{}, nil
}

// Eval is
func (ip *Interpreter) Eval(src string) (reflect.Value, error) {
	// ctx.ReadDir = func(string) ([]os.FileInfo, error) {
	// 	return []os.FileInfo{fakeDirInfo("foo")}, nil
	// }
	// ctx.OpenFile = func(string) (io.ReadCloser, error) {
	//
	// }
	// 	return ip.eval(ProcessUnsafeOffsetof(src))

	return reflect.Value{}, nil
}

func (ip *Interpreter) eval(pkg *build.Package, src string) (reflect.Value, error) {
	// pkg.Imports

	// process import

	// ip.yip.EvalWithContext()

	return reflect.Value{}, nil
}
