package yaegi

import (
	"bytes"
	"fmt"
	"go/build"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/script/internal/config"
	"project/script/internal/log"
)

func init() {
	log.SetSource("yaegi")
}

func TestExportGoRoot(t *testing.T) {
	var cfg config.Config
	if !config.Load("../../../config.json", &cfg) {
		t.Fatal("failed to load project config")
	}
	for _, item := range [...]*struct {
		version int
		path    string
	}{
		{version: 16, path: cfg.Common.Go116x},
		{version: 10, path: cfg.Common.Go1108},
		{version: 11, path: cfg.Specific.Go11113},
		{version: 12, path: cfg.Specific.Go11217},
		{version: 13, path: cfg.Specific.Go11315},
		{version: 14, path: cfg.Specific.Go11415},
		{version: 15, path: cfg.Specific.Go115x},
	} {
		if item.path == "" {
			continue
		}
		testExportGoRoot(t, item.version, item.path)
	}
}

func testExportGoRoot(t *testing.T, version int, path string) {
	const template = `
// Code generated by script/code/yaegi/goroot_test.go. DO NOT EDIT.

// +build go1.%d,!go1.%d

package goroot

import (
	"go/constant"
	"go/token"
%s)

// Symbols stores the map of unsafe package symbols.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
%s}

%s`
	// set build context
	var releaseTags []string
	for i := 1; i <= version; i++ {
		releaseTags = append(releaseTags, "go1."+strconv.Itoa(i))
	}
	build.Default.ReleaseTags = releaseTags
	build.Default.GOROOT = path

	importBuf := bytes.NewBuffer(make([]byte, 0, 2048))
	initBuf := bytes.NewBuffer(make([]byte, 0, 4096))
	codeBuf := bytes.NewBuffer(make([]byte, 0, 128*1024))

	for _, pkg := range []string{
		"archive/tar",
		"archive/zip",
		"bufio",
		"bytes",
		"compress/bzip2",
		"compress/flate",
		"compress/gzip",
		"compress/lzw",
		"compress/zlib",
		"container/heap",
		"container/list",
		"container/ring",
		"context",
		"crypto",
		"crypto/aes",
		"crypto/cipher",
		"crypto/des",
		"crypto/dsa",
		"crypto/ecdsa",
		"crypto/ed25519",
		"crypto/elliptic",
		"crypto/hmac",
		"crypto/md5",
		"crypto/rand",
		"crypto/rc4",
		"crypto/rsa",
		"crypto/sha1",
		"crypto/sha256",
		"crypto/sha512",
		"crypto/subtle",
		"crypto/tls",
		"crypto/x509",
		"database/sql",
		"debug/dwarf",
		"debug/elf",
		"debug/gosym",
		"debug/macho",
		"debug/pe",
		"debug/plan9obj",
		"encoding",
		"encoding/ascii85",
		"encoding/asn1",
		"encoding/base32",
		"encoding/base64",
		"encoding/binary",
		"encoding/csv",
		"encoding/gob",
		"encoding/hex",
		"encoding/json",
		"encoding/pem",
		"encoding/xml",
		"errors",
		"expvar",
		"flag",
		"fmt",
		"hash",
		"hash/adler32",
		"hash/crc32",
		"hash/crc64",
		"hash/fnv",
		"hash/maphash",
		"io",
		"math",
		"math/big",
		"math/bits",
		"math/cmplx",
		"math/rand",
		"reflect",
		"strings",
		"time",
	} {
		switch pkg {
		case "crypto/rand": // same package name with "math/rand"
			_, _ = fmt.Fprintln(importBuf, "\tcrypto_rand  \"crypto/rand\"")
		case "crypto/des", "crypto/dsa", "crypto/md5", "crypto/rc4", "crypto/sha1": // insecure package
			_, _ = fmt.Fprintf(importBuf, "\t\"%s\" // #nosec\n", pkg)
		default:
			_, _ = fmt.Fprintf(importBuf, "\t\"%s\"\n", pkg)
		}
		init := strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(pkg)
		_, _ = fmt.Fprintf(initBuf, "\tinit_%s()\n", init)
		code, err := generateCode(pkg, init)
		require.NoError(t, err)
		// process package "crypto/rand"
		if pkg == "crypto/rand" {
			code = strings.ReplaceAll(code, "rand.", "crypto_rand.")
		}
		codeBuf.WriteString(code)
	}

	code := fmt.Sprintf(template[1:], version, version+1, importBuf, initBuf, codeBuf)
	path = fmt.Sprintf("../../../internal/interpreter/yaegi/goroot/bundle_go1_%d.go", version)
	formatCodeAndSave(t, code, path)
}
