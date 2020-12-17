package yaegi

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/system"
)

func TestExportGoRoot(t *testing.T) {
	const template = `
// Code generated by script/code/yaegi/goroot_test.go. DO NOT EDIT.

package goroot

import (
%s)

// Symbols stores the map of unsafe package symbols.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
%s}

%s`
	// TODO generate different go
	// build.Default.GOROOT = "D:\\Go1.10"

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	srcBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, pkg := range []string{
		"archive/zip",
		"reflect",
		"strings",
	} {
		init := strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(pkg)
		_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", pkg)
		_, _ = fmt.Fprintf(initBuf, "\tinit_%s()\n", init)
		code, err := generateCode(pkg, init)
		require.NoError(t, err)
		srcBuf.WriteString(code)
	}

	// generate code
	src := fmt.Sprintf(template[1:], importBuf, initBuf, srcBuf)

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/interpreter/yaegi/goroot/bundle.go"
	err := system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}