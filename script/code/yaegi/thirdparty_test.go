package yaegi

import (
	"bytes"
	"fmt"
	"go/build"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportThirdParty(t *testing.T) {
	const template = `
// Code generated by script/code/yaegi/thirdparty_test.go. DO NOT EDIT.

package thirdparty

import (
	"crypto"
	"go/constant"
	"go/token"
	"io"
	"net"
	"reflect"

%s)

// Symbols stores the map of unsafe package symbols.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
%s}

%s`

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	codeBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, pkg := range []string{
		"github.com/pelletier/go-toml",
		"github.com/pkg/errors",
		"github.com/vmihailenco/msgpack/v5",
		"github.com/vmihailenco/msgpack/v5/msgpcode",
		"golang.org/x/crypto/ssh",
	} {
		_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", pkg)
		init := strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(pkg)
		_, _ = fmt.Fprintf(initBuf, "\tinit_%s()\n", init)
		code, err := generateCode(pkg, init)
		require.NoError(t, err)
		codeBuf.WriteString(code)
	}

	code := fmt.Sprintf(template[1:], importBuf, initBuf, codeBuf)
	const path = "../../../internal/interpreter/yaegi/thirdparty/bundle.go"
	formatCodeAndSave(t, code, path)
}

func TestExportThirdParty_Windows(t *testing.T) {
	const template = `
// Code generated by script/code/yaegi/thirdparty_test.go. DO NOT EDIT.

// +build windows

package thirdparty

import (
	"go/constant"
	"go/token"
	"reflect"

%s)

func init() {
%s}

%s`

	goos := build.Default.GOOS
	build.Default.GOOS = "windows"
	defer func() { build.Default.GOOS = goos }()

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	codeBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, pkg := range []string{
		"github.com/go-ole/go-ole",
		"github.com/go-ole/go-ole/oleutil",
	} {
		_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", pkg)
		init := strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(pkg)
		_, _ = fmt.Fprintf(initBuf, "\tinit_%s()\n", init)
		code, err := generateCode(pkg, init)
		require.NoError(t, err)
		codeBuf.WriteString(code)
	}

	code := fmt.Sprintf(template[1:], importBuf, initBuf, codeBuf)
	const path = "../../../internal/interpreter/yaegi/thirdparty/windows.go"
	formatCodeAndSave(t, code, path)
}
