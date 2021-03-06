package anko

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/script/internal/config"
)

func TestExportThirdParty(t *testing.T) {
	const template = `
// Code generated by script/code/anko/thirdparty_test.go. DO NOT EDIT.

package thirdparty

import (
	"reflect"

%s
	"project/external/anko/env"
)

func init() {
%s}
%s
`
	// get module directory
	goMod, err := config.GoModCache()
	require.NoError(t, err)

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	srcBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, item := range [...]*struct {
		path string
		dir  string
		init string
	}{
		{
			path: "github.com/pelletier/go-toml",
			dir:  "github.com/pelletier/go-toml@v1.8.1",
			init: "GithubComPelletierGoTOML",
		},
		{
			path: "github.com/pkg/errors",
			dir:  "github.com/pkg/errors@v0.9.1",
			init: "GithubComPkgErrors",
		},
		{
			path: "github.com/vmihailenco/msgpack/v5",
			dir:  "github.com/vmihailenco/msgpack/v5@v5.0.0",
			init: "GithubComVmihailencoMsgpackV5",
		},
		{
			path: "github.com/vmihailenco/msgpack/v5/msgpcode",
			dir:  "github.com/vmihailenco/msgpack/v5@v5.0.0/msgpcode",
			init: "GithubComVmihailencoMsgpackV5Msgpcode",
		},
	} {
		_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", item.path)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(goMod, item.path, item.dir, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template[1:], importBuf, initBuf, srcBuf)

	// fix code
	// for _, item := range [...]*struct {
	// 	old string
	// 	new string
	// }{
	// 	{"interface service.Interface", "iface service.Interface"},
	// 	{"(&interface)", "(&iface)"},
	//
	// 	{"service service.Service", "svc service.Service"},
	// 	{"(&service)", "(&svc)"},
	// } {
	// 	src = strings.ReplaceAll(src, item.old, item.new)
	// }

	// delete code
	for _, item := range [...]string{
		`		"DecodeDatastoreKey": reflect.ValueOf(msgpack.DecodeDatastoreKey),` + "\n",
		`		"EncodeDatastoreKey": reflect.ValueOf(msgpack.EncodeDatastoreKey),` + "\n",
	} {
		src = strings.ReplaceAll(src, item, "")
	}

	const path = "../../../internal/interpreter/anko/thirdparty/bundle.go"
	formatCodeAndSave(t, src, path)
}

func TestExportThirdParty_Windows(t *testing.T) {
	const template = `
// Code generated by script/code/anko/thirdparty_test.go. DO NOT EDIT.

// +build windows

package thirdparty

import (
	"reflect"

%s
	"project/external/anko/env"
)

func init() {
%s}
%s
`
	// get module directory
	goMod, err := config.GoModCache()
	require.NoError(t, err)

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	srcBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, item := range [...]*struct {
		path string
		dir  string
		init string
	}{
		{
			path: "github.com/go-ole/go-ole",
			dir:  "github.com/go-ole/go-ole@v1.2.5-0.20201122170103-d467d8080fc3",
			init: "GithubComGoOLEGoOLE",
		},
		{
			path: "github.com/go-ole/go-ole/oleutil",
			dir:  "github.com/go-ole/go-ole@v1.2.5-0.20201122170103-d467d8080fc3/oleutil",
			init: "GithubComGoOLEGoOLEOLEUtil",
		},
	} {
		_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", item.path)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(goMod, item.path, item.dir, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template[1:], importBuf, initBuf, srcBuf)

	// fix code
	for _, item := range [...]*struct {
		old string
		new string
	}{
		// overflows int
		{"(ole.CO_E_CLASSSTRING)", "(uint32(ole.CO_E_CLASSSTRING))"},
		{"(ole.E_ABORT)", "(uint32(ole.E_ABORT))"},
		{"(ole.E_ACCESSDENIED)", "(uint32(ole.E_ACCESSDENIED))"},
		{"(ole.E_FAIL)", "(uint32(ole.E_FAIL))"},
		{"(ole.E_HANDLE)", "(uint32(ole.E_HANDLE))"},
		{"(ole.E_INVALIDARG)", "(uint32(ole.E_INVALIDARG))"},
		{"(ole.E_NOINTERFACE)", "(uint32(ole.E_NOINTERFACE))"},
		{"(ole.E_NOTIMPL)", "(uint32(ole.E_NOTIMPL))"},
		{"(ole.E_OUTOFMEMORY)", "(uint32(ole.E_OUTOFMEMORY))"},
		{"(ole.E_PENDING)", "(uint32(ole.E_PENDING))"},
		{"(ole.E_POINTER)", "(uint32(ole.E_POINTER))"},
		{"(ole.E_UNEXPECTED)", "(uint32(ole.E_UNEXPECTED))"},
	} {
		src = strings.ReplaceAll(src, item.old, item.new)
	}

	// delete code
	// for _, item := range [...]string{
	// 	`		"DecodeDatastoreKey": reflect.ValueOf(msgpack.DecodeDatastoreKey),` + "\n",
	// 	`		"EncodeDatastoreKey": reflect.ValueOf(msgpack.EncodeDatastoreKey),` + "\n",
	// } {
	// 	src = strings.ReplaceAll(src, item, "")
	// }

	const path = "../../../internal/interpreter/anko/thirdparty/windows.go"
	formatCodeAndSave(t, src, path)
}
