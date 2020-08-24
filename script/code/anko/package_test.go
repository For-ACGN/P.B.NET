package anko

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/system"

	"project/script/internal/config"
)

func TestExportGoRoot(t *testing.T) {
	const template = `// Package goroot generate by script/code/anko/package.go, don't edit it.
package goroot

import (
%s
	"github.com/mattn/anko/env"
)

func init() {
%s}
%s
`
	// get GOROOT
	goRoot, err := config.GoRoot()
	require.NoError(t, err)
	goRoot = filepath.Join(goRoot, "src")

	pkgBuf := new(bytes.Buffer)
	initBuf := new(bytes.Buffer)
	srcBuf := new(bytes.Buffer)

	for _, item := range [...]*struct {
		name string
		init string
	}{
		{"archive/zip", "ArchiveZip"},
		{"bufio", "BufIO"},
		{"bytes", "Bytes"},
		{"compress/bzip2", "CompressBZip2"},
		{"compress/flate", "CompressFlate"},
		{"compress/gzip", "CompressGZip"},
		{"compress/zlib", "CompressZlib"},
		{"container/heap", "ContainerHeap"},
		{"container/list", "ContainerList"},
		{"container/ring", "ContainerRing"},
		{"context", "Context"},
		{"crypto", "Crypto"},
		{"crypto/aes", "CryptoAES"},
		{"crypto/cipher", "CryptoCipher"},
		{"crypto/des", "CryptoDES"},
		{"crypto/dsa", "CryptoDSA"},
		{"crypto/ecdsa", "CryptoECDSA"},
		{"crypto/ed25519", "CryptoED25519"},
		{"crypto/elliptic", "CryptoElliptic"},
		{"crypto/hmac", "CryptoHMAC"},
		{"crypto/md5", "CryptoMD5"},
		{"crypto/rc4", "CryptoRC4"},
		{"crypto/rsa", "CryptoRSA"},
		{"crypto/sha1", "CryptoSHA1"},
		{"crypto/sha256", "CryptoSHA256"},
		{"crypto/sha512", "CryptoSHA512"},
		{"crypto/subtle", "CryptoSubtle"},
		{"crypto/tls", "CryptoTLS"},
		{"crypto/x509", "CryptoX509"},
		{"crypto/x509/pkix", "CryptoX509PKIX"},
		{"encoding", "Encoding"},
		{"encoding/ascii85", "EncodingASCII85"},
		{"encoding/base32", "EncodingBase32"},
		{"encoding/base64", "EncodingBase64"},
		{"encoding/binary", "EncodingBinary"},
		{"encoding/csv", "EncodingCSV"},
		{"encoding/hex", "EncodingHex"},
		{"encoding/json", "EncodingJSON"},
		{"encoding/pem", "EncodingPEM"},
		{"encoding/xml", "EncodingXML"},
		{"fmt", "FMT"},
		{"hash", "Hash"},
		{"hash/crc32", "HashCRC32"},
		{"hash/crc64", "HashCRC64"},
		{"image", "Image"},
		{"image/color", "ImageColor"},
		{"image/draw", "ImageDraw"},
		{"image/gif", "ImageGIF"},
		{"image/jpeg", "ImageJPEG"},
		{"image/png", "ImagePNG"},
		{"io", "IO"},
		{"io/ioutil", "IOioutil"},
		{"log", "Log"},
		{"math", "Math"},
		{"math/big", "MathBig"},
		{"math/bits", "MathBits"},
		{"math/cmplx", "MathCmplx"},
		{"math/rand", "MathRand"},
		{"mime", "MIME"},
		{"mime/multipart", "MIMEMultiPart"},
		{"mime/quotedprintable", "MIMEQuotedPrintable"},
		{"net", "Net"},
		{"net/http", "NetHTTP"},
		{"net/http/cookiejar", "NetHTTPCookieJar"},
		{"net/mail", "NetMail"},
		{"net/smtp", "NetSMTP"},
		{"net/textproto", "NetTextProto"},
		{"net/url", "NetURL"},
		{"os", "OS"},
		{"os/exec", "OSExec"},
		{"os/signal", "OSSignal"},
		{"os/user", "OSUser"},
		{"path", "Path"},
		{"path/filepath", "PathFilepath"},
		{"reflect", "Reflect"},
		{"regexp", "Regexp"},
		{"sort", "Sort"},
		{"strconv", "Strconv"},
		{"strings", "Strings"},
		{"sync", "Sync"},
		{"sync/atomic", "SyncAtomic"},
		{"time", "Time"},
		{"unicode", "Unicode"},
		{"unicode/utf16", "UnicodeUTF16"},
		{"unicode/utf8", "UnicodeUTF8"},
	} {
		_, _ = fmt.Fprintf(pkgBuf, `	"%s"`+"\n", item.name)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(goRoot, item.name, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template, pkgBuf, initBuf, srcBuf)

	// fix code
	for _, item := range [...]*struct {
		old string
		new string
	}{
		{"interface heap.Interface", "iface heap.Interface"},
		{"(&interface)", "(&iface)"},

		{"list list.List", "ll list.List"},
		{"(&list)", "(&ll)"},

		{"ring ring.Ring", "r ring.Ring"},
		{"(&ring)", "(&r)"},

		{"context context.Context", "ctx context.Context"},
		{"(&context)", "(&ctx)"},

		{"cipher rc4.Cipher", "cip rc4.Cipher"},
		{"(&cipher)", "(&cip)"},

		{"hash crypto.Hash", "h crypto.Hash"},
		{"(&hash)", "(&h)"},

		{"encoding base64.Encoding", "enc base64.Encoding"},
		{"(&encoding)", "(&enc)"},

		{"encoding base32.Encoding", "enc base32.Encoding"},
		{"(&encoding)", "(&enc)"},

		{"hash hash.Hash", "h hash.Hash"},
		{"(&hash)", "(&h)"},

		{"image image.Image", "img image.Image"},
		{"(&image)", "(&img)"},

		{"color color.Color", "c color.Color"},
		{"(&color)", "(&c)"},

		{"image draw.Image", "img draw.Image"},
		{"(&image)", "(&img)"},

		{"int big.Int", "i big.Int"},
		{"(&int)", "(&i)"},

		{"rand rand.Rand", "r rand.Rand"},
		{"(&rand)", "(&r)"},

		{"error net.Error", "err net.Error"},
		{"(&error)", "(&err)"},

		{"interface net.Interface", "iface net.Interface"},
		{"(&interface)", "(&iface)"},

		{"error textproto.Error", "err textproto.Error"},
		{"(&error)", "(&err)"},

		{"error url.Error", "err url.Error"},
		{"(&error)", "(&err)"},

		{"signal os.Signal", "sig os.Signal"},
		{"(&signal)", "(&sig)"},

		{"error exec.Error", "err exec.Error"},
		{"(&error)", "(&err)"},

		{"user user.User", "usr user.User"},
		{"(&user)", "(&usr)"},

		{"type reflect.Type", "typ reflect.Type"},
		{"(&type)", "(&typ)"},

		{"regexp regexp.Regexp", "reg regexp.Regexp"},
		{"(&regexp)", "(&reg)"},

		{"interface sort.Interface", "iface sort.Interface"},
		{"(&interface)", "(&iface)"},

		{"map sync.Map", "m sync.Map"},
		{"(&map)", "(&m)"},

		{"time time.Time", "t time.Time"},
		{"(&time)", "(&t)"},

		{"time time.Time", "t time.Time"},
		{"(&time)", "(&t)"},

		// overflows int
		{"(crc32.IEEE)", "(uint32(crc32.IEEE))"},
		{"(crc32.Castagnoli)", "(uint32(crc32.Castagnoli))"},
		{"(crc32.Koopman)", "(uint32(crc32.Koopman))"},
		{"(crc64.ECMA)", "(uint64(crc64.ECMA))"},
		{"(crc64.ISO)", "(uint64(crc64.ISO))"},
		{"(math.MinInt64)", "(int64(math.MinInt64))"},
		{"(math.MaxInt64)", "(int64(math.MaxInt64))"},
		{"(math.MaxUint32)", "(uint32(math.MaxUint32))"},
		{"(math.MaxUint64)", "(uint64(math.MaxUint64))"},
		{"(big.MaxPrec)", "(uint32(big.MaxPrec))"},

		// skip gosec
		{`	"crypto/des"`, `	"crypto/des" // #nosec`},
		{`	"crypto/md5"`, `	"crypto/md5" // #nosec`},
		{`	"crypto/rc4"`, `	"crypto/rc4" // #nosec`},
		{`	"crypto/sha1"`, `	"crypto/sha1" // #nosec`},

		// improve variable name
		{"unknownUserIdError", "unknownUserIDError"},
		{"unknownGroupIdError", "unknownGroupIDError"},
	} {
		src = strings.ReplaceAll(src, item.old, item.new)
	}

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/anko/goroot/bundle.go"
	err = system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}

func TestExportThirdParty(t *testing.T) {
	const template = `// Package thirdparty generate by script/code/anko/package.go, don't edit it.
package thirdparty

import (
	"reflect"

%s	"github.com/mattn/anko/env"
)

func init() {
%s}
%s
`
	// get module directory
	goMod, err := config.GoModCache()
	require.NoError(t, err)

	pkgBuf := new(bytes.Buffer)
	initBuf := new(bytes.Buffer)
	srcBuf := new(bytes.Buffer)

	for _, item := range [...]*struct {
		name string
		path string
		init string
	}{
		{
			name: "github.com/pelletier/go-toml",
			path: "github.com/pelletier/go-toml@v1.8.1",
			init: "GithubComPelletierGoTOML",
		},
		{
			name: "github.com/vmihailenco/msgpack/v5",
			path: "github.com/vmihailenco/msgpack/v5@v5.0.0",
			init: "GithubComVmihailencoMsgpackV5",
		},
		{
			name: "github.com/vmihailenco/msgpack/v5/msgpcode",
			path: "github.com/vmihailenco/msgpack/v5@v5.0.0/msgpcode",
			init: "GithubComVmihailencoMsgpackV5Msgpcode",
		},
	} {
		_, _ = fmt.Fprintf(pkgBuf, `	"%s"`+"\n", item.name)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(goMod, item.path, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template, pkgBuf, initBuf, srcBuf)

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
	for _, item := range []string{
		`		"DecodeDatastoreKey": reflect.ValueOf(msgpack.DecodeDatastoreKey),` + "\n",
		`		"EncodeDatastoreKey": reflect.ValueOf(msgpack.EncodeDatastoreKey),` + "\n",
	} {
		src = strings.ReplaceAll(src, item, "")
	}

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/anko/thirdparty/bundle.go"
	err = system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}

func TestExportThirdPartyWindows(t *testing.T) {
	const template = `// +build windows

// Package thirdparty generate by script/code/anko/package.go, don't edit it.
package thirdparty

import (
	"reflect"

%s	"github.com/mattn/anko/env"
)

func init() {
%s}
%s
`
	// get module directory
	goMod, err := config.GoModCache()
	require.NoError(t, err)

	pkgBuf := new(bytes.Buffer)
	initBuf := new(bytes.Buffer)
	srcBuf := new(bytes.Buffer)

	for _, item := range [...]*struct {
		name string
		path string
		init string
	}{
		{
			name: "github.com/go-ole/go-ole",
			path: "github.com/go-ole/go-ole@v1.2.5-0.20201122170103-d467d8080fc3",
			init: "GithubComGoOLEGoOLE",
		},
		{
			name: "github.com/go-ole/go-ole/oleutil",
			path: "github.com/go-ole/go-ole@v1.2.5-0.20201122170103-d467d8080fc3/oleutil",
			init: "GithubComGoOLEGoOLEOLEUtil",
		},
	} {
		_, _ = fmt.Fprintf(pkgBuf, `	"%s"`+"\n", item.name)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(goMod, item.path, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template, pkgBuf, initBuf, srcBuf)

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
	// for _, item := range []string{
	// 	`		"DecodeDatastoreKey": reflect.ValueOf(msgpack.DecodeDatastoreKey),` + "\n",
	// 	`		"EncodeDatastoreKey": reflect.ValueOf(msgpack.EncodeDatastoreKey),` + "\n",
	// } {
	// 	src = strings.ReplaceAll(src, item, "")
	// }

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/anko/thirdparty/windows.go"
	err = system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}

func TestExportProject(t *testing.T) {
	const template = `// Package project generate by script/code/anko/package.go, don't edit it.
package project

import (
	"reflect"

	"github.com/mattn/anko/env"

%s)

func init() {
%s}
%s
`
	// get project directory
	dir, err := os.Getwd()
	require.NoError(t, err)
	dir, err = filepath.Abs(dir + "/../../..")
	require.NoError(t, err)

	pkgBuf := new(bytes.Buffer)
	initBuf := new(bytes.Buffer)
	srcBuf := new(bytes.Buffer)

	for _, item := range [...]*struct {
		name string
		init string
	}{
		{"internal/cert", "InternalCert"},
		{"internal/convert", "InternalConvert"},
		{"internal/crypto/aes", "InternalCryptoAES"},
		{"internal/crypto/curve25519", "InternalCryptoCurve25519"},
		{"internal/crypto/ed25519", "InternalCryptoED25519"},
		{"internal/crypto/hmac", "InternalCryptoHMAC"},
		{"internal/crypto/lsb", "InternalCryptoLSB"},
		{"internal/crypto/rand", "InternalCryptoRand"},
		{"internal/dns", "InternalDNS"},
		{"internal/guid", "InternalGUID"},
		{"internal/httptool", "InternalHTTPTool"},
		{"internal/logger", "InternalLogger"},
		{"internal/namer", "InternalNamer"},
		{"internal/nettool", "InternalNetTool"},
		{"internal/option", "InternalOption"},
		{"internal/patch/json", "InternalPatchJSON"},
		{"internal/patch/msgpack", "InternalPatchMsgpack"},
		{"internal/patch/toml", "InternalPatchToml"},
		{"internal/proxy", "InternalProxy"},
		{"internal/proxy/direct", "InternalProxyDirect"},
		{"internal/proxy/http", "InternalProxyHTTP"},
		{"internal/proxy/socks", "InternalProxySocks"},
		{"internal/random", "InternalRandom"},
		{"internal/security", "InternalSecurity"},
		{"internal/system", "InternalSystem"},
		{"internal/timesync", "InternalTimeSync"},
		{"internal/xpanic", "InternalXPanic"},
		{"internal/xreflect", "InternalXReflect"},
		{"internal/xsync", "InternalXSync"},
	} {
		_, _ = fmt.Fprintf(pkgBuf, `	"project/%s"`+"\n", item.name)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(dir, "$"+item.name, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template, pkgBuf, initBuf, srcBuf)

	// fix code
	for _, item := range [...]*struct {
		old string
		new string
	}{
		{"logger logger.Logger", "lg logger.Logger"},
		{"(&logger)", "(&lg)"},

		{"namer namer.Namer", "n namer.Namer"},
		{"(&namer)", "(&n)"},

		{"direct direct.Direct", "d direct.Direct"},
		{"(&direct)", "(&d)"},

		{"rand random.Rand", "r random.Rand"},
		{"(&rand)", "(&r)"},
	} {
		src = strings.ReplaceAll(src, item.old, item.new)
	}

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/anko/project/bundle.go"
	err = system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}

func TestExportProjectWindows(t *testing.T) {
	const template = `// +build windows

// Package project generate by script/code/anko/package.go, don't edit it.
package project

import (
	"reflect"

	"github.com/mattn/anko/env"

%s)

func init() {
%s}
%s
`
	// get project directory
	dir, err := os.Getwd()
	require.NoError(t, err)
	dir, err = filepath.Abs(dir + "/../../..")
	require.NoError(t, err)

	pkgBuf := new(bytes.Buffer)
	initBuf := new(bytes.Buffer)
	srcBuf := new(bytes.Buffer)

	for _, item := range [...]*struct {
		name string
		init string
	}{
		{"internal/module/wmi", "InternalModuleWMI"},
	} {
		_, _ = fmt.Fprintf(pkgBuf, `	"project/%s"`+"\n", item.name)
		_, _ = fmt.Fprintf(initBuf, "\tinit%s()\n", item.init)
		src, err := exportDeclaration(dir, "$"+item.name, item.init)
		require.NoError(t, err)
		srcBuf.WriteString(src)
	}

	// generate code
	src := fmt.Sprintf(template, pkgBuf, initBuf, srcBuf)

	// fix code
	for _, item := range [...]*struct {
		old string
		new string
	}{
		// {"logger logger.Logger", "lg logger.Logger"},
		// {"(&logger)", "(&lg)"},
	} {
		src = strings.ReplaceAll(src, item.old, item.new)
	}

	// print and save code
	fmt.Println(src)
	const path = "../../../internal/anko/project/windows.go"
	err = system.WriteFile(path, []byte(src))
	require.NoError(t, err)
}
