package yaegi

import (
	"bufio"
	"reflect"
	"strings"

	"github.com/traefik/yaegi/interp"
)

// Interpreter is a wrapper about yaegi interpreter, but add compiler to merge
// source files to a single file, Process unsafe.Offsetof() and remote importer.
type Interpreter struct {
	yip *interp.Interpreter
}

// NewInterpreter is used to create a new interpreter.
func NewInterpreter() *Interpreter {
	i := Interpreter{yip: interp.New(interp.Options{})}
	return &i
}

func (ip *Interpreter) EvalPath(path string) (reflect.Value, error) {
	return reflect.Value{}, nil
}

func (ip *Interpreter) EvalFiles(files []string) (reflect.Value, error) {
	return reflect.Value{}, nil
}

func (ip *Interpreter) Eval(src string) (reflect.Value, error) {
	// ctx.ReadDir = func(string) ([]os.FileInfo, error) {
	// 	return []os.FileInfo{fakeDirInfo("foo")}, nil
	// }
	// ctx.OpenFile = func(string) (io.ReadCloser, error) {
	//
	// }
	return ip.eval(ProcessUnsafeOffsetof(src))
}

func (ip *Interpreter) eval(src string) (reflect.Value, error) {
	// process import
	scanner := bufio.NewScanner(strings.NewReader(src))
	getPackageName(scanner)

	// ip.yip.EvalWithContext()

	return reflect.Value{}, nil
}

func getPackageName(scanner *bufio.Scanner) string {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, " package ") {

		}
	}
	return ""
}

func getImport(scanner *bufio.Scanner) []string {
	return nil
}
