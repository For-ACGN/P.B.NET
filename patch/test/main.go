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
	"strings"
	"time"
)

func main() {
	// bytes
	buf := bytes.NewBuffer(nil)
	bytes.ReplaceAll(nil, nil, nil)
	bytes.ToValidUTF8(nil, nil)

	// crypto/ed25519
	ed25519.GenerateKey(rand.Reader)

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
	io.ReadAll(buf)

	// log
	logger := log.New(os.Stderr, "", log.LstdFlags)
	fmt.Println(logger.Writer())

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
	http.NewRequestWithContext(context.Background(), http.MethodGet, "http://test.com/", nil)

	// net/textproto
	textproto.MIMEHeader(httpHeader).Values("test")

	// os
	new(os.ProcessState).ExitCode()
	os.UserCacheDir()
	os.UserHomeDir()
	os.UserConfigDir()
	os.ReadDir("test")
	os.CreateTemp("", "")
	os.MkdirTemp("", "")

	// strings
	new(strings.Builder).Len()
	strings.ReplaceAll("", "", "")
	strings.ToValidUTF8("", "")

	time.Duration(111).Microseconds()
	time.Duration(111).Milliseconds()

	fmt.Println("--------------------------------")
	fmt.Println("build pass!")
	fmt.Println("--------------------------------")
}
