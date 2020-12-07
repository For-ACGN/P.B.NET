// this package is used to test patch files can build pass.

package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

func main() {
	fmt.Println("--------------------------------")
	fmt.Println("go version:", runtime.Version())
	fmt.Println("--------------------------------")

	// bytes
	buf := bytes.NewBuffer(nil)
	bytes.ReplaceAll(nil, nil, nil)
	bytes.ToValidUTF8(nil, nil)

	// crypto/ed25519
	fmt.Println(ed25519.GenerateKey(rand.Reader))

	// crypto/x509
	x509.NewCertPool().Certs()

	// encoding/json
	json.NewDecoder(buf).InputOffset()

	// errors
	errors.Is(nil, nil)

	// io
	var sw io.StringWriter
	fmt.Println(sw)
	fmt.Println(io.Discard)
	fmt.Println(io.NopCloser(nil))
	fmt.Println(io.ReadAll(buf))

	// log
	logger := log.New(os.Stderr, "", log.LstdFlags)
	fmt.Println(logger.Writer())
	fmt.Println(log.Writer())
	fmt.Println(log.Default())

	// net
	fmt.Println(net.ErrClosed)

	// net/http
	httpClient := http.Client{
		Transport: new(http.Transport),
	}
	httpClient.CloseIdleConnections()
	httpHeader := make(http.Header)
	httpHeader.Clone()
	httpHeader.Values("test")
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://test.com/", nil)
	req.Clone(ctx)
	new(http.Transport).Clone()

	// net/textproto
	textproto.MIMEHeader(httpHeader).Values("test")

	// os
	new(os.ProcessState).ExitCode()
	fmt.Println(new(os.File).SyscallConn())
	fmt.Println(os.UserCacheDir())
	fmt.Println(os.UserHomeDir())
	fmt.Println(os.UserConfigDir())

	fmt.Println(os.ReadDir("test"))
	fmt.Println(os.CreateTemp("", ""))
	fmt.Println(os.MkdirTemp("", ""))

	// reflect
	fmt.Println(reflect.ValueOf("").IsZero()) // IsZero

	m := make(map[string]string) // MapIter
	m["test"] = "value"
	mapIter := reflect.ValueOf(m).MapRange()
	var ok bool
	for mapIter.Next() {
		if mapIter.Key().Interface().(string) != "test" {
			fmt.Println("invalid map key")
			os.Exit(1)
		}
		if mapIter.Value().Interface().(string) != "value" {
			fmt.Println("invalid map value")
			os.Exit(1)
		}
		ok = true
	}
	if !ok {
		fmt.Println("invalid map iter")
		os.Exit(1)
	}

	// runtime
	runtime.GC()

	// strings
	new(strings.Builder).Len()
	strings.ReplaceAll("", "", "")
	strings.ToValidUTF8("", "")

	// time
	time.Duration(111).Microseconds()
	time.Duration(111).Milliseconds()

	fmt.Println("--------------------------------")
	fmt.Println("pass!")
	fmt.Println("--------------------------------")
}
