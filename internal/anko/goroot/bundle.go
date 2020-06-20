// Package goroot generate by script/code/anko/package.go, don't edit it.
package goroot

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"container/heap"
	"container/list"
	"container/ring"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des" // #nosec
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5" // #nosec
	"crypto/rc4" // #nosec
	"crypto/rsa"
	"crypto/sha1" // #nosec
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding"
	"encoding/ascii85"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/bits"
	"math/cmplx"
	"math/rand"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/mattn/anko/env"
)

func init() {
	initArchiveZip()
	initBufIO()
	initBytes()
	initCompressBZip2()
	initCompressFlate()
	initCompressGZip()
	initCompressZlib()
	initContainerHeap()
	initContainerList()
	initContainerRing()
	initContext()
	initCrypto()
	initCryptoAES()
	initCryptoCipher()
	initCryptoDES()
	initCryptoDSA()
	initCryptoECDSA()
	initCryptoEd25519()
	initCryptoElliptic()
	initCryptoHMAC()
	initCryptoMD5()
	initCryptoRC4()
	initCryptoRSA()
	initCryptoSHA1()
	initCryptoSHA256()
	initCryptoSHA512()
	initCryptoSubtle()
	initCryptoTLS()
	initCryptoX509()
	initCryptoX509PKIX()
	initEncoding()
	initEncodingASCII85()
	initEncodingBase32()
	initEncodingBase64()
	initEncodingBinary()
	initEncodingCSV()
	initEncodingHex()
	initEncodingJSON()
	initEncodingPEM()
	initEncodingXML()
	initFMT()
	initHash()
	initHashCRC32()
	initHashCRC64()
	initImage()
	initImageColor()
	initImageDraw()
	initImageGIF()
	initImageJPEG()
	initImagePNG()
	initIO()
	initIOioutil()
	initLog()
	initMath()
	initMathBig()
	initMathBits()
	initMathCmplx()
	initMathRand()
	initMIME()
	initMIMEMultiPart()
	initMIMEQuotedPrintable()
	initNet()
	initNetHTTP()
	initNetHTTPCookieJar()
	initNetMail()
	initNetSMTP()
	initNetTextProto()
	initNetURL()
	initOS()
	initOSExec()
	initOSSignal()
	initOSUser()
	initPath()
	initPathFilepath()
	initReflect()
	initRegexp()
	initSort()
	initStrconv()
	initStrings()
	initSync()
	initSyncAtomic()
	initTime()
	initUnicode()
	initUnicodeUTF16()
	initUnicodeUTF8()
}

func initArchiveZip() {
	env.Packages["archive/zip"] = map[string]reflect.Value{
		// define constants
		"Deflate": reflect.ValueOf(zip.Deflate),
		"Store":   reflect.ValueOf(zip.Store),

		// define variables
		"ErrAlgorithm": reflect.ValueOf(zip.ErrAlgorithm),
		"ErrChecksum":  reflect.ValueOf(zip.ErrChecksum),
		"ErrFormat":    reflect.ValueOf(zip.ErrFormat),

		// define functions
		"FileInfoHeader":       reflect.ValueOf(zip.FileInfoHeader),
		"NewReader":            reflect.ValueOf(zip.NewReader),
		"NewWriter":            reflect.ValueOf(zip.NewWriter),
		"OpenReader":           reflect.ValueOf(zip.OpenReader),
		"RegisterCompressor":   reflect.ValueOf(zip.RegisterCompressor),
		"RegisterDecompressor": reflect.ValueOf(zip.RegisterDecompressor),
	}
	var (
		compressor   zip.Compressor
		decompressor zip.Decompressor
		file         zip.File
		fileHeader   zip.FileHeader
		readCloser   zip.ReadCloser
		reader       zip.Reader
		writer       zip.Writer
	)
	env.PackageTypes["archive/zip"] = map[string]reflect.Type{
		"Compressor":   reflect.TypeOf(&compressor).Elem(),
		"Decompressor": reflect.TypeOf(&decompressor).Elem(),
		"File":         reflect.TypeOf(&file).Elem(),
		"FileHeader":   reflect.TypeOf(&fileHeader).Elem(),
		"ReadCloser":   reflect.TypeOf(&readCloser).Elem(),
		"Reader":       reflect.TypeOf(&reader).Elem(),
		"Writer":       reflect.TypeOf(&writer).Elem(),
	}
}

func initBufIO() {
	env.Packages["bufio"] = map[string]reflect.Value{
		// define constants
		"MaxScanTokenSize": reflect.ValueOf(bufio.MaxScanTokenSize),

		// define variables
		"ErrAdvanceTooFar":     reflect.ValueOf(bufio.ErrAdvanceTooFar),
		"ErrBadReadCount":      reflect.ValueOf(bufio.ErrBadReadCount),
		"ErrBufferFull":        reflect.ValueOf(bufio.ErrBufferFull),
		"ErrFinalToken":        reflect.ValueOf(bufio.ErrFinalToken),
		"ErrInvalidUnreadByte": reflect.ValueOf(bufio.ErrInvalidUnreadByte),
		"ErrInvalidUnreadRune": reflect.ValueOf(bufio.ErrInvalidUnreadRune),
		"ErrNegativeAdvance":   reflect.ValueOf(bufio.ErrNegativeAdvance),
		"ErrNegativeCount":     reflect.ValueOf(bufio.ErrNegativeCount),
		"ErrTooLong":           reflect.ValueOf(bufio.ErrTooLong),

		// define functions
		"NewReadWriter": reflect.ValueOf(bufio.NewReadWriter),
		"NewReader":     reflect.ValueOf(bufio.NewReader),
		"NewReaderSize": reflect.ValueOf(bufio.NewReaderSize),
		"NewScanner":    reflect.ValueOf(bufio.NewScanner),
		"NewWriter":     reflect.ValueOf(bufio.NewWriter),
		"NewWriterSize": reflect.ValueOf(bufio.NewWriterSize),
		"ScanBytes":     reflect.ValueOf(bufio.ScanBytes),
		"ScanLines":     reflect.ValueOf(bufio.ScanLines),
		"ScanRunes":     reflect.ValueOf(bufio.ScanRunes),
		"ScanWords":     reflect.ValueOf(bufio.ScanWords),
	}
	var (
		readWriter bufio.ReadWriter
		reader     bufio.Reader
		scanner    bufio.Scanner
		splitFunc  bufio.SplitFunc
		writer     bufio.Writer
	)
	env.PackageTypes["bufio"] = map[string]reflect.Type{
		"ReadWriter": reflect.TypeOf(&readWriter).Elem(),
		"Reader":     reflect.TypeOf(&reader).Elem(),
		"Scanner":    reflect.TypeOf(&scanner).Elem(),
		"SplitFunc":  reflect.TypeOf(&splitFunc).Elem(),
		"Writer":     reflect.TypeOf(&writer).Elem(),
	}
}

func initBytes() {
	env.Packages["bytes"] = map[string]reflect.Value{
		// define constants
		"MinRead": reflect.ValueOf(bytes.MinRead),

		// define variables
		"ErrTooLarge": reflect.ValueOf(bytes.ErrTooLarge),

		// define functions
		"Compare":         reflect.ValueOf(bytes.Compare),
		"Contains":        reflect.ValueOf(bytes.Contains),
		"ContainsAny":     reflect.ValueOf(bytes.ContainsAny),
		"ContainsRune":    reflect.ValueOf(bytes.ContainsRune),
		"Count":           reflect.ValueOf(bytes.Count),
		"Equal":           reflect.ValueOf(bytes.Equal),
		"EqualFold":       reflect.ValueOf(bytes.EqualFold),
		"Fields":          reflect.ValueOf(bytes.Fields),
		"FieldsFunc":      reflect.ValueOf(bytes.FieldsFunc),
		"HasPrefix":       reflect.ValueOf(bytes.HasPrefix),
		"HasSuffix":       reflect.ValueOf(bytes.HasSuffix),
		"Index":           reflect.ValueOf(bytes.Index),
		"IndexAny":        reflect.ValueOf(bytes.IndexAny),
		"IndexByte":       reflect.ValueOf(bytes.IndexByte),
		"IndexFunc":       reflect.ValueOf(bytes.IndexFunc),
		"IndexRune":       reflect.ValueOf(bytes.IndexRune),
		"Join":            reflect.ValueOf(bytes.Join),
		"LastIndex":       reflect.ValueOf(bytes.LastIndex),
		"LastIndexAny":    reflect.ValueOf(bytes.LastIndexAny),
		"LastIndexByte":   reflect.ValueOf(bytes.LastIndexByte),
		"LastIndexFunc":   reflect.ValueOf(bytes.LastIndexFunc),
		"Map":             reflect.ValueOf(bytes.Map),
		"NewBuffer":       reflect.ValueOf(bytes.NewBuffer),
		"NewBufferString": reflect.ValueOf(bytes.NewBufferString),
		"NewReader":       reflect.ValueOf(bytes.NewReader),
		"Repeat":          reflect.ValueOf(bytes.Repeat),
		"Replace":         reflect.ValueOf(bytes.Replace),
		"ReplaceAll":      reflect.ValueOf(bytes.ReplaceAll),
		"Runes":           reflect.ValueOf(bytes.Runes),
		"Split":           reflect.ValueOf(bytes.Split),
		"SplitAfter":      reflect.ValueOf(bytes.SplitAfter),
		"SplitAfterN":     reflect.ValueOf(bytes.SplitAfterN),
		"SplitN":          reflect.ValueOf(bytes.SplitN),
		"Title":           reflect.ValueOf(bytes.Title),
		"ToLower":         reflect.ValueOf(bytes.ToLower),
		"ToLowerSpecial":  reflect.ValueOf(bytes.ToLowerSpecial),
		"ToTitle":         reflect.ValueOf(bytes.ToTitle),
		"ToTitleSpecial":  reflect.ValueOf(bytes.ToTitleSpecial),
		"ToUpper":         reflect.ValueOf(bytes.ToUpper),
		"ToUpperSpecial":  reflect.ValueOf(bytes.ToUpperSpecial),
		"ToValidUTF8":     reflect.ValueOf(bytes.ToValidUTF8),
		"Trim":            reflect.ValueOf(bytes.Trim),
		"TrimFunc":        reflect.ValueOf(bytes.TrimFunc),
		"TrimLeft":        reflect.ValueOf(bytes.TrimLeft),
		"TrimLeftFunc":    reflect.ValueOf(bytes.TrimLeftFunc),
		"TrimPrefix":      reflect.ValueOf(bytes.TrimPrefix),
		"TrimRight":       reflect.ValueOf(bytes.TrimRight),
		"TrimRightFunc":   reflect.ValueOf(bytes.TrimRightFunc),
		"TrimSpace":       reflect.ValueOf(bytes.TrimSpace),
		"TrimSuffix":      reflect.ValueOf(bytes.TrimSuffix),
	}
	var (
		buffer bytes.Buffer
		reader bytes.Reader
	)
	env.PackageTypes["bytes"] = map[string]reflect.Type{
		"Buffer": reflect.TypeOf(&buffer).Elem(),
		"Reader": reflect.TypeOf(&reader).Elem(),
	}
}

func initCompressBZip2() {
	env.Packages["compress/bzip2"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewReader": reflect.ValueOf(bzip2.NewReader),
	}
	var (
		structuralError bzip2.StructuralError
	)
	env.PackageTypes["compress/bzip2"] = map[string]reflect.Type{
		"StructuralError": reflect.TypeOf(&structuralError).Elem(),
	}
}

func initCompressFlate() {
	env.Packages["compress/flate"] = map[string]reflect.Value{
		// define constants
		"BestCompression":    reflect.ValueOf(flate.BestCompression),
		"BestSpeed":          reflect.ValueOf(flate.BestSpeed),
		"DefaultCompression": reflect.ValueOf(flate.DefaultCompression),
		"HuffmanOnly":        reflect.ValueOf(flate.HuffmanOnly),
		"NoCompression":      reflect.ValueOf(flate.NoCompression),

		// define variables

		// define functions
		"NewReader":     reflect.ValueOf(flate.NewReader),
		"NewReaderDict": reflect.ValueOf(flate.NewReaderDict),
		"NewWriter":     reflect.ValueOf(flate.NewWriter),
		"NewWriterDict": reflect.ValueOf(flate.NewWriterDict),
	}
	var (
		corruptInputError flate.CorruptInputError
		internalError     flate.InternalError
		reader            flate.Reader
		resetter          flate.Resetter
		writer            flate.Writer
	)
	env.PackageTypes["compress/flate"] = map[string]reflect.Type{
		"CorruptInputError": reflect.TypeOf(&corruptInputError).Elem(),
		"InternalError":     reflect.TypeOf(&internalError).Elem(),
		"Reader":            reflect.TypeOf(&reader).Elem(),
		"Resetter":          reflect.TypeOf(&resetter).Elem(),
		"Writer":            reflect.TypeOf(&writer).Elem(),
	}
}

func initCompressGZip() {
	env.Packages["compress/gzip"] = map[string]reflect.Value{
		// define constants
		"BestCompression":    reflect.ValueOf(gzip.BestCompression),
		"BestSpeed":          reflect.ValueOf(gzip.BestSpeed),
		"DefaultCompression": reflect.ValueOf(gzip.DefaultCompression),
		"HuffmanOnly":        reflect.ValueOf(gzip.HuffmanOnly),
		"NoCompression":      reflect.ValueOf(gzip.NoCompression),

		// define variables
		"ErrChecksum": reflect.ValueOf(gzip.ErrChecksum),
		"ErrHeader":   reflect.ValueOf(gzip.ErrHeader),

		// define functions
		"NewReader":      reflect.ValueOf(gzip.NewReader),
		"NewWriter":      reflect.ValueOf(gzip.NewWriter),
		"NewWriterLevel": reflect.ValueOf(gzip.NewWriterLevel),
	}
	var (
		header gzip.Header
		reader gzip.Reader
		writer gzip.Writer
	)
	env.PackageTypes["compress/gzip"] = map[string]reflect.Type{
		"Header": reflect.TypeOf(&header).Elem(),
		"Reader": reflect.TypeOf(&reader).Elem(),
		"Writer": reflect.TypeOf(&writer).Elem(),
	}
}

func initCompressZlib() {
	env.Packages["compress/zlib"] = map[string]reflect.Value{
		// define constants
		"BestCompression":    reflect.ValueOf(zlib.BestCompression),
		"BestSpeed":          reflect.ValueOf(zlib.BestSpeed),
		"DefaultCompression": reflect.ValueOf(zlib.DefaultCompression),
		"HuffmanOnly":        reflect.ValueOf(zlib.HuffmanOnly),
		"NoCompression":      reflect.ValueOf(zlib.NoCompression),

		// define variables
		"ErrChecksum":   reflect.ValueOf(zlib.ErrChecksum),
		"ErrDictionary": reflect.ValueOf(zlib.ErrDictionary),
		"ErrHeader":     reflect.ValueOf(zlib.ErrHeader),

		// define functions
		"NewReader":          reflect.ValueOf(zlib.NewReader),
		"NewReaderDict":      reflect.ValueOf(zlib.NewReaderDict),
		"NewWriter":          reflect.ValueOf(zlib.NewWriter),
		"NewWriterLevel":     reflect.ValueOf(zlib.NewWriterLevel),
		"NewWriterLevelDict": reflect.ValueOf(zlib.NewWriterLevelDict),
	}
	var (
		resetter zlib.Resetter
		writer   zlib.Writer
	)
	env.PackageTypes["compress/zlib"] = map[string]reflect.Type{
		"Resetter": reflect.TypeOf(&resetter).Elem(),
		"Writer":   reflect.TypeOf(&writer).Elem(),
	}
}

func initContainerHeap() {
	env.Packages["container/heap"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Fix":    reflect.ValueOf(heap.Fix),
		"Init":   reflect.ValueOf(heap.Init),
		"Pop":    reflect.ValueOf(heap.Pop),
		"Push":   reflect.ValueOf(heap.Push),
		"Remove": reflect.ValueOf(heap.Remove),
	}
	var (
		iface heap.Interface
	)
	env.PackageTypes["container/heap"] = map[string]reflect.Type{
		"Interface": reflect.TypeOf(&iface).Elem(),
	}
}

func initContainerList() {
	env.Packages["container/list"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"New": reflect.ValueOf(list.New),
	}
	var (
		element list.Element
		ll      list.List
	)
	env.PackageTypes["container/list"] = map[string]reflect.Type{
		"Element": reflect.TypeOf(&element).Elem(),
		"List":    reflect.TypeOf(&ll).Elem(),
	}
}

func initContainerRing() {
	env.Packages["container/ring"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"New": reflect.ValueOf(ring.New),
	}
	var (
		r ring.Ring
	)
	env.PackageTypes["container/ring"] = map[string]reflect.Type{
		"Ring": reflect.TypeOf(&r).Elem(),
	}
}

func initContext() {
	env.Packages["context"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Canceled":         reflect.ValueOf(context.Canceled),
		"DeadlineExceeded": reflect.ValueOf(context.DeadlineExceeded),

		// define functions
		"Background":   reflect.ValueOf(context.Background),
		"TODO":         reflect.ValueOf(context.TODO),
		"WithCancel":   reflect.ValueOf(context.WithCancel),
		"WithDeadline": reflect.ValueOf(context.WithDeadline),
		"WithTimeout":  reflect.ValueOf(context.WithTimeout),
		"WithValue":    reflect.ValueOf(context.WithValue),
	}
	var (
		cancelFunc context.CancelFunc
		ctx        context.Context
	)
	env.PackageTypes["context"] = map[string]reflect.Type{
		"CancelFunc": reflect.TypeOf(&cancelFunc).Elem(),
		"Context":    reflect.TypeOf(&ctx).Elem(),
	}
}

func initCrypto() {
	env.Packages["crypto"] = map[string]reflect.Value{
		// define constants
		"BLAKE2b_256": reflect.ValueOf(crypto.BLAKE2b_256),
		"BLAKE2b_384": reflect.ValueOf(crypto.BLAKE2b_384),
		"BLAKE2b_512": reflect.ValueOf(crypto.BLAKE2b_512),
		"BLAKE2s_256": reflect.ValueOf(crypto.BLAKE2s_256),
		"MD4":         reflect.ValueOf(crypto.MD4),
		"MD5":         reflect.ValueOf(crypto.MD5),
		"MD5SHA1":     reflect.ValueOf(crypto.MD5SHA1),
		"RIPEMD160":   reflect.ValueOf(crypto.RIPEMD160),
		"SHA1":        reflect.ValueOf(crypto.SHA1),
		"SHA224":      reflect.ValueOf(crypto.SHA224),
		"SHA256":      reflect.ValueOf(crypto.SHA256),
		"SHA384":      reflect.ValueOf(crypto.SHA384),
		"SHA3_224":    reflect.ValueOf(crypto.SHA3_224),
		"SHA3_256":    reflect.ValueOf(crypto.SHA3_256),
		"SHA3_384":    reflect.ValueOf(crypto.SHA3_384),
		"SHA3_512":    reflect.ValueOf(crypto.SHA3_512),
		"SHA512":      reflect.ValueOf(crypto.SHA512),
		"SHA512_224":  reflect.ValueOf(crypto.SHA512_224),
		"SHA512_256":  reflect.ValueOf(crypto.SHA512_256),

		// define variables

		// define functions
		"RegisterHash": reflect.ValueOf(crypto.RegisterHash),
	}
	var (
		decrypter     crypto.Decrypter
		decrypterOpts crypto.DecrypterOpts
		h             crypto.Hash
		privateKey    crypto.PrivateKey
		publicKey     crypto.PublicKey
		signer        crypto.Signer
		signerOpts    crypto.SignerOpts
	)
	env.PackageTypes["crypto"] = map[string]reflect.Type{
		"Decrypter":     reflect.TypeOf(&decrypter).Elem(),
		"DecrypterOpts": reflect.TypeOf(&decrypterOpts).Elem(),
		"Hash":          reflect.TypeOf(&h).Elem(),
		"PrivateKey":    reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":     reflect.TypeOf(&publicKey).Elem(),
		"Signer":        reflect.TypeOf(&signer).Elem(),
		"SignerOpts":    reflect.TypeOf(&signerOpts).Elem(),
	}
}

func initCryptoAES() {
	env.Packages["crypto/aes"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(aes.BlockSize),

		// define variables

		// define functions
		"NewCipher": reflect.ValueOf(aes.NewCipher),
	}
	var (
		keySizeError aes.KeySizeError
	)
	env.PackageTypes["crypto/aes"] = map[string]reflect.Type{
		"KeySizeError": reflect.TypeOf(&keySizeError).Elem(),
	}
}

func initCryptoCipher() {
	env.Packages["crypto/cipher"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewCBCDecrypter":     reflect.ValueOf(cipher.NewCBCDecrypter),
		"NewCBCEncrypter":     reflect.ValueOf(cipher.NewCBCEncrypter),
		"NewCFBDecrypter":     reflect.ValueOf(cipher.NewCFBDecrypter),
		"NewCFBEncrypter":     reflect.ValueOf(cipher.NewCFBEncrypter),
		"NewCTR":              reflect.ValueOf(cipher.NewCTR),
		"NewGCM":              reflect.ValueOf(cipher.NewGCM),
		"NewGCMWithNonceSize": reflect.ValueOf(cipher.NewGCMWithNonceSize),
		"NewGCMWithTagSize":   reflect.ValueOf(cipher.NewGCMWithTagSize),
		"NewOFB":              reflect.ValueOf(cipher.NewOFB),
	}
	var (
		aEAD         cipher.AEAD
		block        cipher.Block
		blockMode    cipher.BlockMode
		stream       cipher.Stream
		streamReader cipher.StreamReader
		streamWriter cipher.StreamWriter
	)
	env.PackageTypes["crypto/cipher"] = map[string]reflect.Type{
		"AEAD":         reflect.TypeOf(&aEAD).Elem(),
		"Block":        reflect.TypeOf(&block).Elem(),
		"BlockMode":    reflect.TypeOf(&blockMode).Elem(),
		"Stream":       reflect.TypeOf(&stream).Elem(),
		"StreamReader": reflect.TypeOf(&streamReader).Elem(),
		"StreamWriter": reflect.TypeOf(&streamWriter).Elem(),
	}
}

func initCryptoDES() {
	env.Packages["crypto/des"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(des.BlockSize),

		// define variables

		// define functions
		"NewCipher":          reflect.ValueOf(des.NewCipher),
		"NewTripleDESCipher": reflect.ValueOf(des.NewTripleDESCipher),
	}
	var (
		keySizeError des.KeySizeError
	)
	env.PackageTypes["crypto/des"] = map[string]reflect.Type{
		"KeySizeError": reflect.TypeOf(&keySizeError).Elem(),
	}
}

func initCryptoDSA() {
	env.Packages["crypto/dsa"] = map[string]reflect.Value{
		// define constants
		"L1024N160": reflect.ValueOf(dsa.L1024N160),
		"L2048N224": reflect.ValueOf(dsa.L2048N224),
		"L2048N256": reflect.ValueOf(dsa.L2048N256),
		"L3072N256": reflect.ValueOf(dsa.L3072N256),

		// define variables
		"ErrInvalidPublicKey": reflect.ValueOf(dsa.ErrInvalidPublicKey),

		// define functions
		"GenerateKey":        reflect.ValueOf(dsa.GenerateKey),
		"GenerateParameters": reflect.ValueOf(dsa.GenerateParameters),
		"Sign":               reflect.ValueOf(dsa.Sign),
		"Verify":             reflect.ValueOf(dsa.Verify),
	}
	var (
		parameterSizes dsa.ParameterSizes
		parameters     dsa.Parameters
		privateKey     dsa.PrivateKey
		publicKey      dsa.PublicKey
	)
	env.PackageTypes["crypto/dsa"] = map[string]reflect.Type{
		"ParameterSizes": reflect.TypeOf(&parameterSizes).Elem(),
		"Parameters":     reflect.TypeOf(&parameters).Elem(),
		"PrivateKey":     reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":      reflect.TypeOf(&publicKey).Elem(),
	}
}

func initCryptoECDSA() {
	env.Packages["crypto/ecdsa"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"GenerateKey": reflect.ValueOf(ecdsa.GenerateKey),
		"Sign":        reflect.ValueOf(ecdsa.Sign),
		"SignASN1":    reflect.ValueOf(ecdsa.SignASN1),
		"Verify":      reflect.ValueOf(ecdsa.Verify),
		"VerifyASN1":  reflect.ValueOf(ecdsa.VerifyASN1),
	}
	var (
		privateKey ecdsa.PrivateKey
		publicKey  ecdsa.PublicKey
	)
	env.PackageTypes["crypto/ecdsa"] = map[string]reflect.Type{
		"PrivateKey": reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":  reflect.TypeOf(&publicKey).Elem(),
	}
}

func initCryptoEd25519() {
	env.Packages["crypto/ed25519"] = map[string]reflect.Value{
		// define constants
		"PrivateKeySize": reflect.ValueOf(ed25519.PrivateKeySize),
		"PublicKeySize":  reflect.ValueOf(ed25519.PublicKeySize),
		"SeedSize":       reflect.ValueOf(ed25519.SeedSize),
		"SignatureSize":  reflect.ValueOf(ed25519.SignatureSize),

		// define variables

		// define functions
		"GenerateKey":    reflect.ValueOf(ed25519.GenerateKey),
		"NewKeyFromSeed": reflect.ValueOf(ed25519.NewKeyFromSeed),
		"Sign":           reflect.ValueOf(ed25519.Sign),
		"Verify":         reflect.ValueOf(ed25519.Verify),
	}
	var (
		privateKey ed25519.PrivateKey
		publicKey  ed25519.PublicKey
	)
	env.PackageTypes["crypto/ed25519"] = map[string]reflect.Type{
		"PrivateKey": reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":  reflect.TypeOf(&publicKey).Elem(),
	}
}

func initCryptoElliptic() {
	env.Packages["crypto/elliptic"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"GenerateKey":         reflect.ValueOf(elliptic.GenerateKey),
		"Marshal":             reflect.ValueOf(elliptic.Marshal),
		"MarshalCompressed":   reflect.ValueOf(elliptic.MarshalCompressed),
		"P224":                reflect.ValueOf(elliptic.P224),
		"P256":                reflect.ValueOf(elliptic.P256),
		"P384":                reflect.ValueOf(elliptic.P384),
		"P521":                reflect.ValueOf(elliptic.P521),
		"Unmarshal":           reflect.ValueOf(elliptic.Unmarshal),
		"UnmarshalCompressed": reflect.ValueOf(elliptic.UnmarshalCompressed),
	}
	var (
		curve       elliptic.Curve
		curveParams elliptic.CurveParams
	)
	env.PackageTypes["crypto/elliptic"] = map[string]reflect.Type{
		"Curve":       reflect.TypeOf(&curve).Elem(),
		"CurveParams": reflect.TypeOf(&curveParams).Elem(),
	}
}

func initCryptoHMAC() {
	env.Packages["crypto/hmac"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Equal": reflect.ValueOf(hmac.Equal),
		"New":   reflect.ValueOf(hmac.New),
	}
	var ()
	env.PackageTypes["crypto/hmac"] = map[string]reflect.Type{}
}

func initCryptoMD5() {
	env.Packages["crypto/md5"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(md5.BlockSize),
		"Size":      reflect.ValueOf(md5.Size),

		// define variables

		// define functions
		"New": reflect.ValueOf(md5.New),
		"Sum": reflect.ValueOf(md5.Sum),
	}
	var ()
	env.PackageTypes["crypto/md5"] = map[string]reflect.Type{}
}

func initCryptoRC4() {
	env.Packages["crypto/rc4"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewCipher": reflect.ValueOf(rc4.NewCipher),
	}
	var (
		cip          rc4.Cipher
		keySizeError rc4.KeySizeError
	)
	env.PackageTypes["crypto/rc4"] = map[string]reflect.Type{
		"Cipher":       reflect.TypeOf(&cip).Elem(),
		"KeySizeError": reflect.TypeOf(&keySizeError).Elem(),
	}
}

func initCryptoRSA() {
	env.Packages["crypto/rsa"] = map[string]reflect.Value{
		// define constants
		"PSSSaltLengthAuto":       reflect.ValueOf(rsa.PSSSaltLengthAuto),
		"PSSSaltLengthEqualsHash": reflect.ValueOf(rsa.PSSSaltLengthEqualsHash),

		// define variables
		"ErrDecryption":     reflect.ValueOf(rsa.ErrDecryption),
		"ErrMessageTooLong": reflect.ValueOf(rsa.ErrMessageTooLong),
		"ErrVerification":   reflect.ValueOf(rsa.ErrVerification),

		// define functions
		"DecryptOAEP":               reflect.ValueOf(rsa.DecryptOAEP),
		"DecryptPKCS1v15":           reflect.ValueOf(rsa.DecryptPKCS1v15),
		"DecryptPKCS1v15SessionKey": reflect.ValueOf(rsa.DecryptPKCS1v15SessionKey),
		"EncryptOAEP":               reflect.ValueOf(rsa.EncryptOAEP),
		"EncryptPKCS1v15":           reflect.ValueOf(rsa.EncryptPKCS1v15),
		"GenerateKey":               reflect.ValueOf(rsa.GenerateKey),
		"GenerateMultiPrimeKey":     reflect.ValueOf(rsa.GenerateMultiPrimeKey),
		"SignPKCS1v15":              reflect.ValueOf(rsa.SignPKCS1v15),
		"SignPSS":                   reflect.ValueOf(rsa.SignPSS),
		"VerifyPKCS1v15":            reflect.ValueOf(rsa.VerifyPKCS1v15),
		"VerifyPSS":                 reflect.ValueOf(rsa.VerifyPSS),
	}
	var (
		cRTValue               rsa.CRTValue
		oAEPOptions            rsa.OAEPOptions
		pKCS1v15DecryptOptions rsa.PKCS1v15DecryptOptions
		pSSOptions             rsa.PSSOptions
		precomputedValues      rsa.PrecomputedValues
		privateKey             rsa.PrivateKey
		publicKey              rsa.PublicKey
	)
	env.PackageTypes["crypto/rsa"] = map[string]reflect.Type{
		"CRTValue":               reflect.TypeOf(&cRTValue).Elem(),
		"OAEPOptions":            reflect.TypeOf(&oAEPOptions).Elem(),
		"PKCS1v15DecryptOptions": reflect.TypeOf(&pKCS1v15DecryptOptions).Elem(),
		"PSSOptions":             reflect.TypeOf(&pSSOptions).Elem(),
		"PrecomputedValues":      reflect.TypeOf(&precomputedValues).Elem(),
		"PrivateKey":             reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":              reflect.TypeOf(&publicKey).Elem(),
	}
}

func initCryptoSHA1() {
	env.Packages["crypto/sha1"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(sha1.BlockSize),
		"Size":      reflect.ValueOf(sha1.Size),

		// define variables

		// define functions
		"New": reflect.ValueOf(sha1.New),
		"Sum": reflect.ValueOf(sha1.Sum),
	}
	var ()
	env.PackageTypes["crypto/sha1"] = map[string]reflect.Type{}
}

func initCryptoSHA256() {
	env.Packages["crypto/sha256"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(sha256.BlockSize),
		"Size":      reflect.ValueOf(sha256.Size),
		"Size224":   reflect.ValueOf(sha256.Size224),

		// define variables

		// define functions
		"New":    reflect.ValueOf(sha256.New),
		"New224": reflect.ValueOf(sha256.New224),
		"Sum224": reflect.ValueOf(sha256.Sum224),
		"Sum256": reflect.ValueOf(sha256.Sum256),
	}
	var ()
	env.PackageTypes["crypto/sha256"] = map[string]reflect.Type{}
}

func initCryptoSHA512() {
	env.Packages["crypto/sha512"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(sha512.BlockSize),
		"Size":      reflect.ValueOf(sha512.Size),
		"Size224":   reflect.ValueOf(sha512.Size224),
		"Size256":   reflect.ValueOf(sha512.Size256),
		"Size384":   reflect.ValueOf(sha512.Size384),

		// define variables

		// define functions
		"New":        reflect.ValueOf(sha512.New),
		"New384":     reflect.ValueOf(sha512.New384),
		"New512_224": reflect.ValueOf(sha512.New512_224),
		"New512_256": reflect.ValueOf(sha512.New512_256),
		"Sum384":     reflect.ValueOf(sha512.Sum384),
		"Sum512":     reflect.ValueOf(sha512.Sum512),
		"Sum512_224": reflect.ValueOf(sha512.Sum512_224),
		"Sum512_256": reflect.ValueOf(sha512.Sum512_256),
	}
	var ()
	env.PackageTypes["crypto/sha512"] = map[string]reflect.Type{}
}

func initCryptoSubtle() {
	env.Packages["crypto/subtle"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"ConstantTimeByteEq":   reflect.ValueOf(subtle.ConstantTimeByteEq),
		"ConstantTimeCompare":  reflect.ValueOf(subtle.ConstantTimeCompare),
		"ConstantTimeCopy":     reflect.ValueOf(subtle.ConstantTimeCopy),
		"ConstantTimeEq":       reflect.ValueOf(subtle.ConstantTimeEq),
		"ConstantTimeLessOrEq": reflect.ValueOf(subtle.ConstantTimeLessOrEq),
		"ConstantTimeSelect":   reflect.ValueOf(subtle.ConstantTimeSelect),
	}
	var ()
	env.PackageTypes["crypto/subtle"] = map[string]reflect.Type{}
}

func initCryptoTLS() {
	env.Packages["crypto/tls"] = map[string]reflect.Value{
		// define constants
		"CurveP256":                                     reflect.ValueOf(tls.CurveP256),
		"CurveP384":                                     reflect.ValueOf(tls.CurveP384),
		"CurveP521":                                     reflect.ValueOf(tls.CurveP521),
		"ECDSAWithP256AndSHA256":                        reflect.ValueOf(tls.ECDSAWithP256AndSHA256),
		"ECDSAWithP384AndSHA384":                        reflect.ValueOf(tls.ECDSAWithP384AndSHA384),
		"ECDSAWithP521AndSHA512":                        reflect.ValueOf(tls.ECDSAWithP521AndSHA512),
		"ECDSAWithSHA1":                                 reflect.ValueOf(tls.ECDSAWithSHA1),
		"Ed25519":                                       reflect.ValueOf(tls.Ed25519),
		"NoClientCert":                                  reflect.ValueOf(tls.NoClientCert),
		"PKCS1WithSHA1":                                 reflect.ValueOf(tls.PKCS1WithSHA1),
		"PKCS1WithSHA256":                               reflect.ValueOf(tls.PKCS1WithSHA256),
		"PKCS1WithSHA384":                               reflect.ValueOf(tls.PKCS1WithSHA384),
		"PKCS1WithSHA512":                               reflect.ValueOf(tls.PKCS1WithSHA512),
		"PSSWithSHA256":                                 reflect.ValueOf(tls.PSSWithSHA256),
		"PSSWithSHA384":                                 reflect.ValueOf(tls.PSSWithSHA384),
		"PSSWithSHA512":                                 reflect.ValueOf(tls.PSSWithSHA512),
		"RenegotiateFreelyAsClient":                     reflect.ValueOf(tls.RenegotiateFreelyAsClient),
		"RenegotiateNever":                              reflect.ValueOf(tls.RenegotiateNever),
		"RenegotiateOnceAsClient":                       reflect.ValueOf(tls.RenegotiateOnceAsClient),
		"RequestClientCert":                             reflect.ValueOf(tls.RequestClientCert),
		"RequireAndVerifyClientCert":                    reflect.ValueOf(tls.RequireAndVerifyClientCert),
		"RequireAnyClientCert":                          reflect.ValueOf(tls.RequireAnyClientCert),
		"TLS_AES_128_GCM_SHA256":                        reflect.ValueOf(tls.TLS_AES_128_GCM_SHA256),
		"TLS_AES_256_GCM_SHA384":                        reflect.ValueOf(tls.TLS_AES_256_GCM_SHA384),
		"TLS_CHACHA20_POLY1305_SHA256":                  reflect.ValueOf(tls.TLS_CHACHA20_POLY1305_SHA256),
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":          reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA),
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256":       reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256),
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256),
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":          reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA),
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384),
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":        reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305),
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256),
		"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":              reflect.ValueOf(tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA),
		"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":           reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA),
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":            reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA),
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":         reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256),
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256),
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":            reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA),
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384),
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":          reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305),
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256),
		"TLS_ECDHE_RSA_WITH_RC4_128_SHA":                reflect.ValueOf(tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA),
		"TLS_FALLBACK_SCSV":                             reflect.ValueOf(tls.TLS_FALLBACK_SCSV),
		"TLS_RSA_WITH_3DES_EDE_CBC_SHA":                 reflect.ValueOf(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA),
		"TLS_RSA_WITH_AES_128_CBC_SHA":                  reflect.ValueOf(tls.TLS_RSA_WITH_AES_128_CBC_SHA),
		"TLS_RSA_WITH_AES_128_CBC_SHA256":               reflect.ValueOf(tls.TLS_RSA_WITH_AES_128_CBC_SHA256),
		"TLS_RSA_WITH_AES_128_GCM_SHA256":               reflect.ValueOf(tls.TLS_RSA_WITH_AES_128_GCM_SHA256),
		"TLS_RSA_WITH_AES_256_CBC_SHA":                  reflect.ValueOf(tls.TLS_RSA_WITH_AES_256_CBC_SHA),
		"TLS_RSA_WITH_AES_256_GCM_SHA384":               reflect.ValueOf(tls.TLS_RSA_WITH_AES_256_GCM_SHA384),
		"TLS_RSA_WITH_RC4_128_SHA":                      reflect.ValueOf(tls.TLS_RSA_WITH_RC4_128_SHA),
		"VerifyClientCertIfGiven":                       reflect.ValueOf(tls.VerifyClientCertIfGiven),
		"VersionTLS10":                                  reflect.ValueOf(tls.VersionTLS10),
		"VersionTLS11":                                  reflect.ValueOf(tls.VersionTLS11),
		"VersionTLS12":                                  reflect.ValueOf(tls.VersionTLS12),
		"VersionTLS13":                                  reflect.ValueOf(tls.VersionTLS13),
		"X25519":                                        reflect.ValueOf(tls.X25519),

		// define variables

		// define functions
		"CipherSuiteName":          reflect.ValueOf(tls.CipherSuiteName),
		"CipherSuites":             reflect.ValueOf(tls.CipherSuites),
		"Client":                   reflect.ValueOf(tls.Client),
		"Dial":                     reflect.ValueOf(tls.Dial),
		"DialWithDialer":           reflect.ValueOf(tls.DialWithDialer),
		"InsecureCipherSuites":     reflect.ValueOf(tls.InsecureCipherSuites),
		"Listen":                   reflect.ValueOf(tls.Listen),
		"LoadX509KeyPair":          reflect.ValueOf(tls.LoadX509KeyPair),
		"NewLRUClientSessionCache": reflect.ValueOf(tls.NewLRUClientSessionCache),
		"NewListener":              reflect.ValueOf(tls.NewListener),
		"Server":                   reflect.ValueOf(tls.Server),
		"X509KeyPair":              reflect.ValueOf(tls.X509KeyPair),
	}
	var (
		certificate            tls.Certificate
		certificateRequestInfo tls.CertificateRequestInfo
		cipherSuite            tls.CipherSuite
		clientAuthType         tls.ClientAuthType
		clientHelloInfo        tls.ClientHelloInfo
		clientSessionCache     tls.ClientSessionCache
		clientSessionState     tls.ClientSessionState
		config                 tls.Config
		conn                   tls.Conn
		connectionState        tls.ConnectionState
		curveID                tls.CurveID
		dialer                 tls.Dialer
		recordHeaderError      tls.RecordHeaderError
		renegotiationSupport   tls.RenegotiationSupport
		signatureScheme        tls.SignatureScheme
	)
	env.PackageTypes["crypto/tls"] = map[string]reflect.Type{
		"Certificate":            reflect.TypeOf(&certificate).Elem(),
		"CertificateRequestInfo": reflect.TypeOf(&certificateRequestInfo).Elem(),
		"CipherSuite":            reflect.TypeOf(&cipherSuite).Elem(),
		"ClientAuthType":         reflect.TypeOf(&clientAuthType).Elem(),
		"ClientHelloInfo":        reflect.TypeOf(&clientHelloInfo).Elem(),
		"ClientSessionCache":     reflect.TypeOf(&clientSessionCache).Elem(),
		"ClientSessionState":     reflect.TypeOf(&clientSessionState).Elem(),
		"Config":                 reflect.TypeOf(&config).Elem(),
		"Conn":                   reflect.TypeOf(&conn).Elem(),
		"ConnectionState":        reflect.TypeOf(&connectionState).Elem(),
		"CurveID":                reflect.TypeOf(&curveID).Elem(),
		"Dialer":                 reflect.TypeOf(&dialer).Elem(),
		"RecordHeaderError":      reflect.TypeOf(&recordHeaderError).Elem(),
		"RenegotiationSupport":   reflect.TypeOf(&renegotiationSupport).Elem(),
		"SignatureScheme":        reflect.TypeOf(&signatureScheme).Elem(),
	}
}

func initCryptoX509() {
	env.Packages["crypto/x509"] = map[string]reflect.Value{
		// define constants
		"CANotAuthorizedForExtKeyUsage": reflect.ValueOf(x509.CANotAuthorizedForExtKeyUsage),
		"CANotAuthorizedForThisName":    reflect.ValueOf(x509.CANotAuthorizedForThisName),
		"DSA":                           reflect.ValueOf(x509.DSA),
		"DSAWithSHA1":                   reflect.ValueOf(x509.DSAWithSHA1),
		"DSAWithSHA256":                 reflect.ValueOf(x509.DSAWithSHA256),
		"ECDSA":                         reflect.ValueOf(x509.ECDSA),
		"ECDSAWithSHA1":                 reflect.ValueOf(x509.ECDSAWithSHA1),
		"ECDSAWithSHA256":               reflect.ValueOf(x509.ECDSAWithSHA256),
		"ECDSAWithSHA384":               reflect.ValueOf(x509.ECDSAWithSHA384),
		"ECDSAWithSHA512":               reflect.ValueOf(x509.ECDSAWithSHA512),
		"Ed25519":                       reflect.ValueOf(x509.Ed25519),
		"Expired":                       reflect.ValueOf(x509.Expired),
		"ExtKeyUsageAny":                reflect.ValueOf(x509.ExtKeyUsageAny),
		"ExtKeyUsageClientAuth":         reflect.ValueOf(x509.ExtKeyUsageClientAuth),
		"ExtKeyUsageCodeSigning":        reflect.ValueOf(x509.ExtKeyUsageCodeSigning),
		"ExtKeyUsageEmailProtection":    reflect.ValueOf(x509.ExtKeyUsageEmailProtection),
		"ExtKeyUsageIPSECEndSystem":     reflect.ValueOf(x509.ExtKeyUsageIPSECEndSystem),
		"ExtKeyUsageIPSECTunnel":        reflect.ValueOf(x509.ExtKeyUsageIPSECTunnel),
		"ExtKeyUsageIPSECUser":          reflect.ValueOf(x509.ExtKeyUsageIPSECUser),
		"ExtKeyUsageMicrosoftCommercialCodeSigning": reflect.ValueOf(x509.ExtKeyUsageMicrosoftCommercialCodeSigning),
		"ExtKeyUsageMicrosoftKernelCodeSigning":     reflect.ValueOf(x509.ExtKeyUsageMicrosoftKernelCodeSigning),
		"ExtKeyUsageMicrosoftServerGatedCrypto":     reflect.ValueOf(x509.ExtKeyUsageMicrosoftServerGatedCrypto),
		"ExtKeyUsageNetscapeServerGatedCrypto":      reflect.ValueOf(x509.ExtKeyUsageNetscapeServerGatedCrypto),
		"ExtKeyUsageOCSPSigning":                    reflect.ValueOf(x509.ExtKeyUsageOCSPSigning),
		"ExtKeyUsageServerAuth":                     reflect.ValueOf(x509.ExtKeyUsageServerAuth),
		"ExtKeyUsageTimeStamping":                   reflect.ValueOf(x509.ExtKeyUsageTimeStamping),
		"IncompatibleUsage":                         reflect.ValueOf(x509.IncompatibleUsage),
		"KeyUsageCRLSign":                           reflect.ValueOf(x509.KeyUsageCRLSign),
		"KeyUsageCertSign":                          reflect.ValueOf(x509.KeyUsageCertSign),
		"KeyUsageContentCommitment":                 reflect.ValueOf(x509.KeyUsageContentCommitment),
		"KeyUsageDataEncipherment":                  reflect.ValueOf(x509.KeyUsageDataEncipherment),
		"KeyUsageDecipherOnly":                      reflect.ValueOf(x509.KeyUsageDecipherOnly),
		"KeyUsageDigitalSignature":                  reflect.ValueOf(x509.KeyUsageDigitalSignature),
		"KeyUsageEncipherOnly":                      reflect.ValueOf(x509.KeyUsageEncipherOnly),
		"KeyUsageKeyAgreement":                      reflect.ValueOf(x509.KeyUsageKeyAgreement),
		"KeyUsageKeyEncipherment":                   reflect.ValueOf(x509.KeyUsageKeyEncipherment),
		"MD2WithRSA":                                reflect.ValueOf(x509.MD2WithRSA),
		"MD5WithRSA":                                reflect.ValueOf(x509.MD5WithRSA),
		"NameConstraintsWithoutSANs":                reflect.ValueOf(x509.NameConstraintsWithoutSANs),
		"NameMismatch":                              reflect.ValueOf(x509.NameMismatch),
		"NotAuthorizedToSign":                       reflect.ValueOf(x509.NotAuthorizedToSign),
		"PEMCipher3DES":                             reflect.ValueOf(x509.PEMCipher3DES),
		"PEMCipherAES128":                           reflect.ValueOf(x509.PEMCipherAES128),
		"PEMCipherAES192":                           reflect.ValueOf(x509.PEMCipherAES192),
		"PEMCipherAES256":                           reflect.ValueOf(x509.PEMCipherAES256),
		"PEMCipherDES":                              reflect.ValueOf(x509.PEMCipherDES),
		"PureEd25519":                               reflect.ValueOf(x509.PureEd25519),
		"RSA":                                       reflect.ValueOf(x509.RSA),
		"SHA1WithRSA":                               reflect.ValueOf(x509.SHA1WithRSA),
		"SHA256WithRSA":                             reflect.ValueOf(x509.SHA256WithRSA),
		"SHA256WithRSAPSS":                          reflect.ValueOf(x509.SHA256WithRSAPSS),
		"SHA384WithRSA":                             reflect.ValueOf(x509.SHA384WithRSA),
		"SHA384WithRSAPSS":                          reflect.ValueOf(x509.SHA384WithRSAPSS),
		"SHA512WithRSA":                             reflect.ValueOf(x509.SHA512WithRSA),
		"SHA512WithRSAPSS":                          reflect.ValueOf(x509.SHA512WithRSAPSS),
		"TooManyConstraints":                        reflect.ValueOf(x509.TooManyConstraints),
		"TooManyIntermediates":                      reflect.ValueOf(x509.TooManyIntermediates),
		"UnconstrainedName":                         reflect.ValueOf(x509.UnconstrainedName),
		"UnknownPublicKeyAlgorithm":                 reflect.ValueOf(x509.UnknownPublicKeyAlgorithm),
		"UnknownSignatureAlgorithm":                 reflect.ValueOf(x509.UnknownSignatureAlgorithm),

		// define variables
		"ErrUnsupportedAlgorithm": reflect.ValueOf(x509.ErrUnsupportedAlgorithm),
		"IncorrectPasswordError":  reflect.ValueOf(x509.IncorrectPasswordError),

		// define functions
		"CreateCertificate":        reflect.ValueOf(x509.CreateCertificate),
		"CreateCertificateRequest": reflect.ValueOf(x509.CreateCertificateRequest),
		"CreateRevocationList":     reflect.ValueOf(x509.CreateRevocationList),
		"DecryptPEMBlock":          reflect.ValueOf(x509.DecryptPEMBlock),
		"EncryptPEMBlock":          reflect.ValueOf(x509.EncryptPEMBlock),
		"IsEncryptedPEMBlock":      reflect.ValueOf(x509.IsEncryptedPEMBlock),
		"MarshalECPrivateKey":      reflect.ValueOf(x509.MarshalECPrivateKey),
		"MarshalPKCS1PrivateKey":   reflect.ValueOf(x509.MarshalPKCS1PrivateKey),
		"MarshalPKCS1PublicKey":    reflect.ValueOf(x509.MarshalPKCS1PublicKey),
		"MarshalPKCS8PrivateKey":   reflect.ValueOf(x509.MarshalPKCS8PrivateKey),
		"MarshalPKIXPublicKey":     reflect.ValueOf(x509.MarshalPKIXPublicKey),
		"NewCertPool":              reflect.ValueOf(x509.NewCertPool),
		"ParseCRL":                 reflect.ValueOf(x509.ParseCRL),
		"ParseCertificate":         reflect.ValueOf(x509.ParseCertificate),
		"ParseCertificateRequest":  reflect.ValueOf(x509.ParseCertificateRequest),
		"ParseCertificates":        reflect.ValueOf(x509.ParseCertificates),
		"ParseDERCRL":              reflect.ValueOf(x509.ParseDERCRL),
		"ParseECPrivateKey":        reflect.ValueOf(x509.ParseECPrivateKey),
		"ParsePKCS1PrivateKey":     reflect.ValueOf(x509.ParsePKCS1PrivateKey),
		"ParsePKCS1PublicKey":      reflect.ValueOf(x509.ParsePKCS1PublicKey),
		"ParsePKCS8PrivateKey":     reflect.ValueOf(x509.ParsePKCS8PrivateKey),
		"ParsePKIXPublicKey":       reflect.ValueOf(x509.ParsePKIXPublicKey),
		"SystemCertPool":           reflect.ValueOf(x509.SystemCertPool),
	}
	var (
		certPool                   x509.CertPool
		certificate                x509.Certificate
		certificateInvalidError    x509.CertificateInvalidError
		certificateRequest         x509.CertificateRequest
		constraintViolationError   x509.ConstraintViolationError
		extKeyUsage                x509.ExtKeyUsage
		hostnameError              x509.HostnameError
		insecureAlgorithmError     x509.InsecureAlgorithmError
		invalidReason              x509.InvalidReason
		keyUsage                   x509.KeyUsage
		pEMCipher                  x509.PEMCipher
		publicKeyAlgorithm         x509.PublicKeyAlgorithm
		revocationList             x509.RevocationList
		signatureAlgorithm         x509.SignatureAlgorithm
		systemRootsError           x509.SystemRootsError
		unhandledCriticalExtension x509.UnhandledCriticalExtension
		unknownAuthorityError      x509.UnknownAuthorityError
		verifyOptions              x509.VerifyOptions
	)
	env.PackageTypes["crypto/x509"] = map[string]reflect.Type{
		"CertPool":                   reflect.TypeOf(&certPool).Elem(),
		"Certificate":                reflect.TypeOf(&certificate).Elem(),
		"CertificateInvalidError":    reflect.TypeOf(&certificateInvalidError).Elem(),
		"CertificateRequest":         reflect.TypeOf(&certificateRequest).Elem(),
		"ConstraintViolationError":   reflect.TypeOf(&constraintViolationError).Elem(),
		"ExtKeyUsage":                reflect.TypeOf(&extKeyUsage).Elem(),
		"HostnameError":              reflect.TypeOf(&hostnameError).Elem(),
		"InsecureAlgorithmError":     reflect.TypeOf(&insecureAlgorithmError).Elem(),
		"InvalidReason":              reflect.TypeOf(&invalidReason).Elem(),
		"KeyUsage":                   reflect.TypeOf(&keyUsage).Elem(),
		"PEMCipher":                  reflect.TypeOf(&pEMCipher).Elem(),
		"PublicKeyAlgorithm":         reflect.TypeOf(&publicKeyAlgorithm).Elem(),
		"RevocationList":             reflect.TypeOf(&revocationList).Elem(),
		"SignatureAlgorithm":         reflect.TypeOf(&signatureAlgorithm).Elem(),
		"SystemRootsError":           reflect.TypeOf(&systemRootsError).Elem(),
		"UnhandledCriticalExtension": reflect.TypeOf(&unhandledCriticalExtension).Elem(),
		"UnknownAuthorityError":      reflect.TypeOf(&unknownAuthorityError).Elem(),
		"VerifyOptions":              reflect.TypeOf(&verifyOptions).Elem(),
	}
}

func initCryptoX509PKIX() {
	env.Packages["crypto/x509/pkix"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		algorithmIdentifier          pkix.AlgorithmIdentifier
		attributeTypeAndValue        pkix.AttributeTypeAndValue
		attributeTypeAndValueSET     pkix.AttributeTypeAndValueSET
		certificateList              pkix.CertificateList
		extension                    pkix.Extension
		name                         pkix.Name
		rDNSequence                  pkix.RDNSequence
		relativeDistinguishedNameSET pkix.RelativeDistinguishedNameSET
		revokedCertificate           pkix.RevokedCertificate
		tBSCertificateList           pkix.TBSCertificateList
	)
	env.PackageTypes["crypto/x509/pkix"] = map[string]reflect.Type{
		"AlgorithmIdentifier":          reflect.TypeOf(&algorithmIdentifier).Elem(),
		"AttributeTypeAndValue":        reflect.TypeOf(&attributeTypeAndValue).Elem(),
		"AttributeTypeAndValueSET":     reflect.TypeOf(&attributeTypeAndValueSET).Elem(),
		"CertificateList":              reflect.TypeOf(&certificateList).Elem(),
		"Extension":                    reflect.TypeOf(&extension).Elem(),
		"Name":                         reflect.TypeOf(&name).Elem(),
		"RDNSequence":                  reflect.TypeOf(&rDNSequence).Elem(),
		"RelativeDistinguishedNameSET": reflect.TypeOf(&relativeDistinguishedNameSET).Elem(),
		"RevokedCertificate":           reflect.TypeOf(&revokedCertificate).Elem(),
		"TBSCertificateList":           reflect.TypeOf(&tBSCertificateList).Elem(),
	}
}

func initEncoding() {
	env.Packages["encoding"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		binaryMarshaler   encoding.BinaryMarshaler
		binaryUnmarshaler encoding.BinaryUnmarshaler
		textMarshaler     encoding.TextMarshaler
		textUnmarshaler   encoding.TextUnmarshaler
	)
	env.PackageTypes["encoding"] = map[string]reflect.Type{
		"BinaryMarshaler":   reflect.TypeOf(&binaryMarshaler).Elem(),
		"BinaryUnmarshaler": reflect.TypeOf(&binaryUnmarshaler).Elem(),
		"TextMarshaler":     reflect.TypeOf(&textMarshaler).Elem(),
		"TextUnmarshaler":   reflect.TypeOf(&textUnmarshaler).Elem(),
	}
}

func initEncodingASCII85() {
	env.Packages["encoding/ascii85"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Decode":        reflect.ValueOf(ascii85.Decode),
		"Encode":        reflect.ValueOf(ascii85.Encode),
		"MaxEncodedLen": reflect.ValueOf(ascii85.MaxEncodedLen),
		"NewDecoder":    reflect.ValueOf(ascii85.NewDecoder),
		"NewEncoder":    reflect.ValueOf(ascii85.NewEncoder),
	}
	var (
		corruptInputError ascii85.CorruptInputError
	)
	env.PackageTypes["encoding/ascii85"] = map[string]reflect.Type{
		"CorruptInputError": reflect.TypeOf(&corruptInputError).Elem(),
	}
}

func initEncodingBase32() {
	env.Packages["encoding/base32"] = map[string]reflect.Value{
		// define constants
		"NoPadding":  reflect.ValueOf(base32.NoPadding),
		"StdPadding": reflect.ValueOf(base32.StdPadding),

		// define variables
		"HexEncoding": reflect.ValueOf(base32.HexEncoding),
		"StdEncoding": reflect.ValueOf(base32.StdEncoding),

		// define functions
		"NewDecoder":  reflect.ValueOf(base32.NewDecoder),
		"NewEncoder":  reflect.ValueOf(base32.NewEncoder),
		"NewEncoding": reflect.ValueOf(base32.NewEncoding),
	}
	var (
		corruptInputError base32.CorruptInputError
		enc               base32.Encoding
	)
	env.PackageTypes["encoding/base32"] = map[string]reflect.Type{
		"CorruptInputError": reflect.TypeOf(&corruptInputError).Elem(),
		"Encoding":          reflect.TypeOf(&enc).Elem(),
	}
}

func initEncodingBase64() {
	env.Packages["encoding/base64"] = map[string]reflect.Value{
		// define constants
		"NoPadding":  reflect.ValueOf(base64.NoPadding),
		"StdPadding": reflect.ValueOf(base64.StdPadding),

		// define variables
		"RawStdEncoding": reflect.ValueOf(base64.RawStdEncoding),
		"RawURLEncoding": reflect.ValueOf(base64.RawURLEncoding),
		"StdEncoding":    reflect.ValueOf(base64.StdEncoding),
		"URLEncoding":    reflect.ValueOf(base64.URLEncoding),

		// define functions
		"NewDecoder":  reflect.ValueOf(base64.NewDecoder),
		"NewEncoder":  reflect.ValueOf(base64.NewEncoder),
		"NewEncoding": reflect.ValueOf(base64.NewEncoding),
	}
	var (
		corruptInputError base64.CorruptInputError
		enc               base64.Encoding
	)
	env.PackageTypes["encoding/base64"] = map[string]reflect.Type{
		"CorruptInputError": reflect.TypeOf(&corruptInputError).Elem(),
		"Encoding":          reflect.TypeOf(&enc).Elem(),
	}
}

func initEncodingBinary() {
	env.Packages["encoding/binary"] = map[string]reflect.Value{
		// define constants
		"MaxVarintLen16": reflect.ValueOf(binary.MaxVarintLen16),
		"MaxVarintLen32": reflect.ValueOf(binary.MaxVarintLen32),
		"MaxVarintLen64": reflect.ValueOf(binary.MaxVarintLen64),

		// define variables
		"BigEndian":    reflect.ValueOf(binary.BigEndian),
		"LittleEndian": reflect.ValueOf(binary.LittleEndian),

		// define functions
		"PutUvarint":  reflect.ValueOf(binary.PutUvarint),
		"PutVarint":   reflect.ValueOf(binary.PutVarint),
		"Read":        reflect.ValueOf(binary.Read),
		"ReadUvarint": reflect.ValueOf(binary.ReadUvarint),
		"ReadVarint":  reflect.ValueOf(binary.ReadVarint),
		"Size":        reflect.ValueOf(binary.Size),
		"Uvarint":     reflect.ValueOf(binary.Uvarint),
		"Varint":      reflect.ValueOf(binary.Varint),
		"Write":       reflect.ValueOf(binary.Write),
	}
	var (
		byteOrder binary.ByteOrder
	)
	env.PackageTypes["encoding/binary"] = map[string]reflect.Type{
		"ByteOrder": reflect.TypeOf(&byteOrder).Elem(),
	}
}

func initEncodingCSV() {
	env.Packages["encoding/csv"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrBareQuote":  reflect.ValueOf(csv.ErrBareQuote),
		"ErrFieldCount": reflect.ValueOf(csv.ErrFieldCount),
		"ErrQuote":      reflect.ValueOf(csv.ErrQuote),

		// define functions
		"NewReader": reflect.ValueOf(csv.NewReader),
		"NewWriter": reflect.ValueOf(csv.NewWriter),
	}
	var (
		parseError csv.ParseError
		reader     csv.Reader
		writer     csv.Writer
	)
	env.PackageTypes["encoding/csv"] = map[string]reflect.Type{
		"ParseError": reflect.TypeOf(&parseError).Elem(),
		"Reader":     reflect.TypeOf(&reader).Elem(),
		"Writer":     reflect.TypeOf(&writer).Elem(),
	}
}

func initEncodingHex() {
	env.Packages["encoding/hex"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrLength": reflect.ValueOf(hex.ErrLength),

		// define functions
		"Decode":         reflect.ValueOf(hex.Decode),
		"DecodeString":   reflect.ValueOf(hex.DecodeString),
		"DecodedLen":     reflect.ValueOf(hex.DecodedLen),
		"Dump":           reflect.ValueOf(hex.Dump),
		"Dumper":         reflect.ValueOf(hex.Dumper),
		"Encode":         reflect.ValueOf(hex.Encode),
		"EncodeToString": reflect.ValueOf(hex.EncodeToString),
		"EncodedLen":     reflect.ValueOf(hex.EncodedLen),
		"NewDecoder":     reflect.ValueOf(hex.NewDecoder),
		"NewEncoder":     reflect.ValueOf(hex.NewEncoder),
	}
	var (
		invalidByteError hex.InvalidByteError
	)
	env.PackageTypes["encoding/hex"] = map[string]reflect.Type{
		"InvalidByteError": reflect.TypeOf(&invalidByteError).Elem(),
	}
}

func initEncodingJSON() {
	env.Packages["encoding/json"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Compact":       reflect.ValueOf(json.Compact),
		"HTMLEscape":    reflect.ValueOf(json.HTMLEscape),
		"Indent":        reflect.ValueOf(json.Indent),
		"Marshal":       reflect.ValueOf(json.Marshal),
		"MarshalIndent": reflect.ValueOf(json.MarshalIndent),
		"NewDecoder":    reflect.ValueOf(json.NewDecoder),
		"NewEncoder":    reflect.ValueOf(json.NewEncoder),
		"Unmarshal":     reflect.ValueOf(json.Unmarshal),
		"Valid":         reflect.ValueOf(json.Valid),
	}
	var (
		decoder               json.Decoder
		delim                 json.Delim
		encoder               json.Encoder
		invalidUnmarshalError json.InvalidUnmarshalError
		marshaler             json.Marshaler
		marshalerError        json.MarshalerError
		number                json.Number
		rawMessage            json.RawMessage
		syntaxError           json.SyntaxError
		token                 json.Token
		unmarshalTypeError    json.UnmarshalTypeError
		unmarshaler           json.Unmarshaler
		unsupportedTypeError  json.UnsupportedTypeError
		unsupportedValueError json.UnsupportedValueError
	)
	env.PackageTypes["encoding/json"] = map[string]reflect.Type{
		"Decoder":               reflect.TypeOf(&decoder).Elem(),
		"Delim":                 reflect.TypeOf(&delim).Elem(),
		"Encoder":               reflect.TypeOf(&encoder).Elem(),
		"InvalidUnmarshalError": reflect.TypeOf(&invalidUnmarshalError).Elem(),
		"Marshaler":             reflect.TypeOf(&marshaler).Elem(),
		"MarshalerError":        reflect.TypeOf(&marshalerError).Elem(),
		"Number":                reflect.TypeOf(&number).Elem(),
		"RawMessage":            reflect.TypeOf(&rawMessage).Elem(),
		"SyntaxError":           reflect.TypeOf(&syntaxError).Elem(),
		"Token":                 reflect.TypeOf(&token).Elem(),
		"UnmarshalTypeError":    reflect.TypeOf(&unmarshalTypeError).Elem(),
		"Unmarshaler":           reflect.TypeOf(&unmarshaler).Elem(),
		"UnsupportedTypeError":  reflect.TypeOf(&unsupportedTypeError).Elem(),
		"UnsupportedValueError": reflect.TypeOf(&unsupportedValueError).Elem(),
	}
}

func initEncodingPEM() {
	env.Packages["encoding/pem"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Decode":         reflect.ValueOf(pem.Decode),
		"Encode":         reflect.ValueOf(pem.Encode),
		"EncodeToMemory": reflect.ValueOf(pem.EncodeToMemory),
	}
	var (
		block pem.Block
	)
	env.PackageTypes["encoding/pem"] = map[string]reflect.Type{
		"Block": reflect.TypeOf(&block).Elem(),
	}
}

func initEncodingXML() {
	env.Packages["encoding/xml"] = map[string]reflect.Value{
		// define constants
		"Header": reflect.ValueOf(xml.Header),

		// define variables
		"HTMLAutoClose": reflect.ValueOf(xml.HTMLAutoClose),
		"HTMLEntity":    reflect.ValueOf(xml.HTMLEntity),

		// define functions
		"CopyToken":       reflect.ValueOf(xml.CopyToken),
		"Escape":          reflect.ValueOf(xml.Escape),
		"EscapeText":      reflect.ValueOf(xml.EscapeText),
		"Marshal":         reflect.ValueOf(xml.Marshal),
		"MarshalIndent":   reflect.ValueOf(xml.MarshalIndent),
		"NewDecoder":      reflect.ValueOf(xml.NewDecoder),
		"NewEncoder":      reflect.ValueOf(xml.NewEncoder),
		"NewTokenDecoder": reflect.ValueOf(xml.NewTokenDecoder),
		"Unmarshal":       reflect.ValueOf(xml.Unmarshal),
	}
	var (
		attr                 xml.Attr
		charData             xml.CharData
		comment              xml.Comment
		decoder              xml.Decoder
		directive            xml.Directive
		encoder              xml.Encoder
		endElement           xml.EndElement
		marshaler            xml.Marshaler
		marshalerAttr        xml.MarshalerAttr
		name                 xml.Name
		procInst             xml.ProcInst
		startElement         xml.StartElement
		syntaxError          xml.SyntaxError
		tagPathError         xml.TagPathError
		token                xml.Token
		tokenReader          xml.TokenReader
		unmarshalError       xml.UnmarshalError
		unmarshaler          xml.Unmarshaler
		unmarshalerAttr      xml.UnmarshalerAttr
		unsupportedTypeError xml.UnsupportedTypeError
	)
	env.PackageTypes["encoding/xml"] = map[string]reflect.Type{
		"Attr":                 reflect.TypeOf(&attr).Elem(),
		"CharData":             reflect.TypeOf(&charData).Elem(),
		"Comment":              reflect.TypeOf(&comment).Elem(),
		"Decoder":              reflect.TypeOf(&decoder).Elem(),
		"Directive":            reflect.TypeOf(&directive).Elem(),
		"Encoder":              reflect.TypeOf(&encoder).Elem(),
		"EndElement":           reflect.TypeOf(&endElement).Elem(),
		"Marshaler":            reflect.TypeOf(&marshaler).Elem(),
		"MarshalerAttr":        reflect.TypeOf(&marshalerAttr).Elem(),
		"Name":                 reflect.TypeOf(&name).Elem(),
		"ProcInst":             reflect.TypeOf(&procInst).Elem(),
		"StartElement":         reflect.TypeOf(&startElement).Elem(),
		"SyntaxError":          reflect.TypeOf(&syntaxError).Elem(),
		"TagPathError":         reflect.TypeOf(&tagPathError).Elem(),
		"Token":                reflect.TypeOf(&token).Elem(),
		"TokenReader":          reflect.TypeOf(&tokenReader).Elem(),
		"UnmarshalError":       reflect.TypeOf(&unmarshalError).Elem(),
		"Unmarshaler":          reflect.TypeOf(&unmarshaler).Elem(),
		"UnmarshalerAttr":      reflect.TypeOf(&unmarshalerAttr).Elem(),
		"UnsupportedTypeError": reflect.TypeOf(&unsupportedTypeError).Elem(),
	}
}

func initFMT() {
	env.Packages["fmt"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Errorf":   reflect.ValueOf(fmt.Errorf),
		"Fprint":   reflect.ValueOf(fmt.Fprint),
		"Fprintf":  reflect.ValueOf(fmt.Fprintf),
		"Fprintln": reflect.ValueOf(fmt.Fprintln),
		"Fscan":    reflect.ValueOf(fmt.Fscan),
		"Fscanf":   reflect.ValueOf(fmt.Fscanf),
		"Fscanln":  reflect.ValueOf(fmt.Fscanln),
		"Print":    reflect.ValueOf(fmt.Print),
		"Printf":   reflect.ValueOf(fmt.Printf),
		"Println":  reflect.ValueOf(fmt.Println),
		"Scan":     reflect.ValueOf(fmt.Scan),
		"Scanf":    reflect.ValueOf(fmt.Scanf),
		"Scanln":   reflect.ValueOf(fmt.Scanln),
		"Sprint":   reflect.ValueOf(fmt.Sprint),
		"Sprintf":  reflect.ValueOf(fmt.Sprintf),
		"Sprintln": reflect.ValueOf(fmt.Sprintln),
		"Sscan":    reflect.ValueOf(fmt.Sscan),
		"Sscanf":   reflect.ValueOf(fmt.Sscanf),
		"Sscanln":  reflect.ValueOf(fmt.Sscanln),
	}
	var (
		formatter  fmt.Formatter
		goStringer fmt.GoStringer
		scanState  fmt.ScanState
		scanner    fmt.Scanner
		state      fmt.State
		stringer   fmt.Stringer
	)
	env.PackageTypes["fmt"] = map[string]reflect.Type{
		"Formatter":  reflect.TypeOf(&formatter).Elem(),
		"GoStringer": reflect.TypeOf(&goStringer).Elem(),
		"ScanState":  reflect.TypeOf(&scanState).Elem(),
		"Scanner":    reflect.TypeOf(&scanner).Elem(),
		"State":      reflect.TypeOf(&state).Elem(),
		"Stringer":   reflect.TypeOf(&stringer).Elem(),
	}
}

func initHash() {
	env.Packages["hash"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		h      hash.Hash
		hash32 hash.Hash32
		hash64 hash.Hash64
	)
	env.PackageTypes["hash"] = map[string]reflect.Type{
		"Hash":   reflect.TypeOf(&h).Elem(),
		"Hash32": reflect.TypeOf(&hash32).Elem(),
		"Hash64": reflect.TypeOf(&hash64).Elem(),
	}
}

func initHashCRC32() {
	env.Packages["hash/crc32"] = map[string]reflect.Value{
		// define constants
		"Castagnoli": reflect.ValueOf(uint32(crc32.Castagnoli)),
		"IEEE":       reflect.ValueOf(uint32(crc32.IEEE)),
		"Koopman":    reflect.ValueOf(uint32(crc32.Koopman)),
		"Size":       reflect.ValueOf(crc32.Size),

		// define variables
		"IEEETable": reflect.ValueOf(crc32.IEEETable),

		// define functions
		"Checksum":     reflect.ValueOf(crc32.Checksum),
		"ChecksumIEEE": reflect.ValueOf(crc32.ChecksumIEEE),
		"MakeTable":    reflect.ValueOf(crc32.MakeTable),
		"New":          reflect.ValueOf(crc32.New),
		"NewIEEE":      reflect.ValueOf(crc32.NewIEEE),
		"Update":       reflect.ValueOf(crc32.Update),
	}
	var (
		table crc32.Table
	)
	env.PackageTypes["hash/crc32"] = map[string]reflect.Type{
		"Table": reflect.TypeOf(&table).Elem(),
	}
}

func initHashCRC64() {
	env.Packages["hash/crc64"] = map[string]reflect.Value{
		// define constants
		"ECMA": reflect.ValueOf(uint64(crc64.ECMA)),
		"ISO":  reflect.ValueOf(uint64(crc64.ISO)),
		"Size": reflect.ValueOf(crc64.Size),

		// define variables

		// define functions
		"Checksum":  reflect.ValueOf(crc64.Checksum),
		"MakeTable": reflect.ValueOf(crc64.MakeTable),
		"New":       reflect.ValueOf(crc64.New),
		"Update":    reflect.ValueOf(crc64.Update),
	}
	var (
		table crc64.Table
	)
	env.PackageTypes["hash/crc64"] = map[string]reflect.Type{
		"Table": reflect.TypeOf(&table).Elem(),
	}
}

func initImage() {
	env.Packages["image"] = map[string]reflect.Value{
		// define constants
		"YCbCrSubsampleRatio410": reflect.ValueOf(image.YCbCrSubsampleRatio410),
		"YCbCrSubsampleRatio411": reflect.ValueOf(image.YCbCrSubsampleRatio411),
		"YCbCrSubsampleRatio420": reflect.ValueOf(image.YCbCrSubsampleRatio420),
		"YCbCrSubsampleRatio422": reflect.ValueOf(image.YCbCrSubsampleRatio422),
		"YCbCrSubsampleRatio440": reflect.ValueOf(image.YCbCrSubsampleRatio440),
		"YCbCrSubsampleRatio444": reflect.ValueOf(image.YCbCrSubsampleRatio444),

		// define variables
		"Black":       reflect.ValueOf(image.Black),
		"ErrFormat":   reflect.ValueOf(image.ErrFormat),
		"Opaque":      reflect.ValueOf(image.Opaque),
		"Transparent": reflect.ValueOf(image.Transparent),
		"White":       reflect.ValueOf(image.White),

		// define functions
		"Decode":         reflect.ValueOf(image.Decode),
		"DecodeConfig":   reflect.ValueOf(image.DecodeConfig),
		"NewAlpha":       reflect.ValueOf(image.NewAlpha),
		"NewAlpha16":     reflect.ValueOf(image.NewAlpha16),
		"NewCMYK":        reflect.ValueOf(image.NewCMYK),
		"NewGray":        reflect.ValueOf(image.NewGray),
		"NewGray16":      reflect.ValueOf(image.NewGray16),
		"NewNRGBA":       reflect.ValueOf(image.NewNRGBA),
		"NewNRGBA64":     reflect.ValueOf(image.NewNRGBA64),
		"NewNYCbCrA":     reflect.ValueOf(image.NewNYCbCrA),
		"NewPaletted":    reflect.ValueOf(image.NewPaletted),
		"NewRGBA":        reflect.ValueOf(image.NewRGBA),
		"NewRGBA64":      reflect.ValueOf(image.NewRGBA64),
		"NewUniform":     reflect.ValueOf(image.NewUniform),
		"NewYCbCr":       reflect.ValueOf(image.NewYCbCr),
		"Pt":             reflect.ValueOf(image.Pt),
		"Rect":           reflect.ValueOf(image.Rect),
		"RegisterFormat": reflect.ValueOf(image.RegisterFormat),
	}
	var (
		alpha               image.Alpha
		alpha16             image.Alpha16
		cMYK                image.CMYK
		config              image.Config
		gray                image.Gray
		gray16              image.Gray16
		img                 image.Image
		nRGBA               image.NRGBA
		nRGBA64             image.NRGBA64
		nYCbCrA             image.NYCbCrA
		paletted            image.Paletted
		palettedImage       image.PalettedImage
		point               image.Point
		rGBA                image.RGBA
		rGBA64              image.RGBA64
		rectangle           image.Rectangle
		uniform             image.Uniform
		yCbCr               image.YCbCr
		yCbCrSubsampleRatio image.YCbCrSubsampleRatio
	)
	env.PackageTypes["image"] = map[string]reflect.Type{
		"Alpha":               reflect.TypeOf(&alpha).Elem(),
		"Alpha16":             reflect.TypeOf(&alpha16).Elem(),
		"CMYK":                reflect.TypeOf(&cMYK).Elem(),
		"Config":              reflect.TypeOf(&config).Elem(),
		"Gray":                reflect.TypeOf(&gray).Elem(),
		"Gray16":              reflect.TypeOf(&gray16).Elem(),
		"Image":               reflect.TypeOf(&img).Elem(),
		"NRGBA":               reflect.TypeOf(&nRGBA).Elem(),
		"NRGBA64":             reflect.TypeOf(&nRGBA64).Elem(),
		"NYCbCrA":             reflect.TypeOf(&nYCbCrA).Elem(),
		"Paletted":            reflect.TypeOf(&paletted).Elem(),
		"PalettedImage":       reflect.TypeOf(&palettedImage).Elem(),
		"Point":               reflect.TypeOf(&point).Elem(),
		"RGBA":                reflect.TypeOf(&rGBA).Elem(),
		"RGBA64":              reflect.TypeOf(&rGBA64).Elem(),
		"Rectangle":           reflect.TypeOf(&rectangle).Elem(),
		"Uniform":             reflect.TypeOf(&uniform).Elem(),
		"YCbCr":               reflect.TypeOf(&yCbCr).Elem(),
		"YCbCrSubsampleRatio": reflect.TypeOf(&yCbCrSubsampleRatio).Elem(),
	}
}

func initImageColor() {
	env.Packages["image/color"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Alpha16Model": reflect.ValueOf(color.Alpha16Model),
		"AlphaModel":   reflect.ValueOf(color.AlphaModel),
		"Black":        reflect.ValueOf(color.Black),
		"CMYKModel":    reflect.ValueOf(color.CMYKModel),
		"Gray16Model":  reflect.ValueOf(color.Gray16Model),
		"GrayModel":    reflect.ValueOf(color.GrayModel),
		"NRGBA64Model": reflect.ValueOf(color.NRGBA64Model),
		"NRGBAModel":   reflect.ValueOf(color.NRGBAModel),
		"NYCbCrAModel": reflect.ValueOf(color.NYCbCrAModel),
		"Opaque":       reflect.ValueOf(color.Opaque),
		"RGBA64Model":  reflect.ValueOf(color.RGBA64Model),
		"RGBAModel":    reflect.ValueOf(color.RGBAModel),
		"Transparent":  reflect.ValueOf(color.Transparent),
		"White":        reflect.ValueOf(color.White),
		"YCbCrModel":   reflect.ValueOf(color.YCbCrModel),

		// define functions
		"CMYKToRGB":  reflect.ValueOf(color.CMYKToRGB),
		"ModelFunc":  reflect.ValueOf(color.ModelFunc),
		"RGBToCMYK":  reflect.ValueOf(color.RGBToCMYK),
		"RGBToYCbCr": reflect.ValueOf(color.RGBToYCbCr),
		"YCbCrToRGB": reflect.ValueOf(color.YCbCrToRGB),
	}
	var (
		alpha   color.Alpha
		alpha16 color.Alpha16
		cMYK    color.CMYK
		c       color.Color
		gray    color.Gray
		gray16  color.Gray16
		model   color.Model
		nRGBA   color.NRGBA
		nRGBA64 color.NRGBA64
		nYCbCrA color.NYCbCrA
		palette color.Palette
		rGBA    color.RGBA
		rGBA64  color.RGBA64
		yCbCr   color.YCbCr
	)
	env.PackageTypes["image/color"] = map[string]reflect.Type{
		"Alpha":   reflect.TypeOf(&alpha).Elem(),
		"Alpha16": reflect.TypeOf(&alpha16).Elem(),
		"CMYK":    reflect.TypeOf(&cMYK).Elem(),
		"Color":   reflect.TypeOf(&c).Elem(),
		"Gray":    reflect.TypeOf(&gray).Elem(),
		"Gray16":  reflect.TypeOf(&gray16).Elem(),
		"Model":   reflect.TypeOf(&model).Elem(),
		"NRGBA":   reflect.TypeOf(&nRGBA).Elem(),
		"NRGBA64": reflect.TypeOf(&nRGBA64).Elem(),
		"NYCbCrA": reflect.TypeOf(&nYCbCrA).Elem(),
		"Palette": reflect.TypeOf(&palette).Elem(),
		"RGBA":    reflect.TypeOf(&rGBA).Elem(),
		"RGBA64":  reflect.TypeOf(&rGBA64).Elem(),
		"YCbCr":   reflect.TypeOf(&yCbCr).Elem(),
	}
}

func initImageDraw() {
	env.Packages["image/draw"] = map[string]reflect.Value{
		// define constants
		"Over": reflect.ValueOf(draw.Over),
		"Src":  reflect.ValueOf(draw.Src),

		// define variables
		"FloydSteinberg": reflect.ValueOf(draw.FloydSteinberg),

		// define functions
		"Draw":     reflect.ValueOf(draw.Draw),
		"DrawMask": reflect.ValueOf(draw.DrawMask),
	}
	var (
		drawer    draw.Drawer
		img       draw.Image
		op        draw.Op
		quantizer draw.Quantizer
	)
	env.PackageTypes["image/draw"] = map[string]reflect.Type{
		"Drawer":    reflect.TypeOf(&drawer).Elem(),
		"Image":     reflect.TypeOf(&img).Elem(),
		"Op":        reflect.TypeOf(&op).Elem(),
		"Quantizer": reflect.TypeOf(&quantizer).Elem(),
	}
}

func initImageGIF() {
	env.Packages["image/gif"] = map[string]reflect.Value{
		// define constants
		"DisposalBackground": reflect.ValueOf(gif.DisposalBackground),
		"DisposalNone":       reflect.ValueOf(gif.DisposalNone),
		"DisposalPrevious":   reflect.ValueOf(gif.DisposalPrevious),

		// define variables

		// define functions
		"Decode":       reflect.ValueOf(gif.Decode),
		"DecodeAll":    reflect.ValueOf(gif.DecodeAll),
		"DecodeConfig": reflect.ValueOf(gif.DecodeConfig),
		"Encode":       reflect.ValueOf(gif.Encode),
		"EncodeAll":    reflect.ValueOf(gif.EncodeAll),
	}
	var (
		gIF     gif.GIF
		options gif.Options
	)
	env.PackageTypes["image/gif"] = map[string]reflect.Type{
		"GIF":     reflect.TypeOf(&gIF).Elem(),
		"Options": reflect.TypeOf(&options).Elem(),
	}
}

func initImageJPEG() {
	env.Packages["image/jpeg"] = map[string]reflect.Value{
		// define constants
		"DefaultQuality": reflect.ValueOf(jpeg.DefaultQuality),

		// define variables

		// define functions
		"Decode":       reflect.ValueOf(jpeg.Decode),
		"DecodeConfig": reflect.ValueOf(jpeg.DecodeConfig),
		"Encode":       reflect.ValueOf(jpeg.Encode),
	}
	var (
		formatError      jpeg.FormatError
		options          jpeg.Options
		unsupportedError jpeg.UnsupportedError
	)
	env.PackageTypes["image/jpeg"] = map[string]reflect.Type{
		"FormatError":      reflect.TypeOf(&formatError).Elem(),
		"Options":          reflect.TypeOf(&options).Elem(),
		"UnsupportedError": reflect.TypeOf(&unsupportedError).Elem(),
	}
}

func initImagePNG() {
	env.Packages["image/png"] = map[string]reflect.Value{
		// define constants
		"BestCompression":    reflect.ValueOf(png.BestCompression),
		"BestSpeed":          reflect.ValueOf(png.BestSpeed),
		"DefaultCompression": reflect.ValueOf(png.DefaultCompression),
		"NoCompression":      reflect.ValueOf(png.NoCompression),

		// define variables

		// define functions
		"Decode":       reflect.ValueOf(png.Decode),
		"DecodeConfig": reflect.ValueOf(png.DecodeConfig),
		"Encode":       reflect.ValueOf(png.Encode),
	}
	var (
		compressionLevel  png.CompressionLevel
		encoder           png.Encoder
		encoderBuffer     png.EncoderBuffer
		encoderBufferPool png.EncoderBufferPool
		formatError       png.FormatError
		unsupportedError  png.UnsupportedError
	)
	env.PackageTypes["image/png"] = map[string]reflect.Type{
		"CompressionLevel":  reflect.TypeOf(&compressionLevel).Elem(),
		"Encoder":           reflect.TypeOf(&encoder).Elem(),
		"EncoderBuffer":     reflect.TypeOf(&encoderBuffer).Elem(),
		"EncoderBufferPool": reflect.TypeOf(&encoderBufferPool).Elem(),
		"FormatError":       reflect.TypeOf(&formatError).Elem(),
		"UnsupportedError":  reflect.TypeOf(&unsupportedError).Elem(),
	}
}

func initIO() {
	env.Packages["io"] = map[string]reflect.Value{
		// define constants
		"SeekCurrent": reflect.ValueOf(io.SeekCurrent),
		"SeekEnd":     reflect.ValueOf(io.SeekEnd),
		"SeekStart":   reflect.ValueOf(io.SeekStart),

		// define variables
		"EOF":              reflect.ValueOf(io.EOF),
		"ErrClosedPipe":    reflect.ValueOf(io.ErrClosedPipe),
		"ErrNoProgress":    reflect.ValueOf(io.ErrNoProgress),
		"ErrShortBuffer":   reflect.ValueOf(io.ErrShortBuffer),
		"ErrShortWrite":    reflect.ValueOf(io.ErrShortWrite),
		"ErrUnexpectedEOF": reflect.ValueOf(io.ErrUnexpectedEOF),

		// define functions
		"Copy":             reflect.ValueOf(io.Copy),
		"CopyBuffer":       reflect.ValueOf(io.CopyBuffer),
		"CopyN":            reflect.ValueOf(io.CopyN),
		"LimitReader":      reflect.ValueOf(io.LimitReader),
		"MultiReader":      reflect.ValueOf(io.MultiReader),
		"MultiWriter":      reflect.ValueOf(io.MultiWriter),
		"NewSectionReader": reflect.ValueOf(io.NewSectionReader),
		"Pipe":             reflect.ValueOf(io.Pipe),
		"ReadAtLeast":      reflect.ValueOf(io.ReadAtLeast),
		"ReadFull":         reflect.ValueOf(io.ReadFull),
		"TeeReader":        reflect.ValueOf(io.TeeReader),
		"WriteString":      reflect.ValueOf(io.WriteString),
	}
	var (
		byteReader      io.ByteReader
		byteScanner     io.ByteScanner
		byteWriter      io.ByteWriter
		closer          io.Closer
		limitedReader   io.LimitedReader
		pipeReader      io.PipeReader
		pipeWriter      io.PipeWriter
		readCloser      io.ReadCloser
		readSeeker      io.ReadSeeker
		readWriteCloser io.ReadWriteCloser
		readWriteSeeker io.ReadWriteSeeker
		readWriter      io.ReadWriter
		reader          io.Reader
		readerAt        io.ReaderAt
		readerFrom      io.ReaderFrom
		runeReader      io.RuneReader
		runeScanner     io.RuneScanner
		sectionReader   io.SectionReader
		seeker          io.Seeker
		stringWriter    io.StringWriter
		writeCloser     io.WriteCloser
		writeSeeker     io.WriteSeeker
		writer          io.Writer
		writerAt        io.WriterAt
		writerTo        io.WriterTo
	)
	env.PackageTypes["io"] = map[string]reflect.Type{
		"ByteReader":      reflect.TypeOf(&byteReader).Elem(),
		"ByteScanner":     reflect.TypeOf(&byteScanner).Elem(),
		"ByteWriter":      reflect.TypeOf(&byteWriter).Elem(),
		"Closer":          reflect.TypeOf(&closer).Elem(),
		"LimitedReader":   reflect.TypeOf(&limitedReader).Elem(),
		"PipeReader":      reflect.TypeOf(&pipeReader).Elem(),
		"PipeWriter":      reflect.TypeOf(&pipeWriter).Elem(),
		"ReadCloser":      reflect.TypeOf(&readCloser).Elem(),
		"ReadSeeker":      reflect.TypeOf(&readSeeker).Elem(),
		"ReadWriteCloser": reflect.TypeOf(&readWriteCloser).Elem(),
		"ReadWriteSeeker": reflect.TypeOf(&readWriteSeeker).Elem(),
		"ReadWriter":      reflect.TypeOf(&readWriter).Elem(),
		"Reader":          reflect.TypeOf(&reader).Elem(),
		"ReaderAt":        reflect.TypeOf(&readerAt).Elem(),
		"ReaderFrom":      reflect.TypeOf(&readerFrom).Elem(),
		"RuneReader":      reflect.TypeOf(&runeReader).Elem(),
		"RuneScanner":     reflect.TypeOf(&runeScanner).Elem(),
		"SectionReader":   reflect.TypeOf(&sectionReader).Elem(),
		"Seeker":          reflect.TypeOf(&seeker).Elem(),
		"StringWriter":    reflect.TypeOf(&stringWriter).Elem(),
		"WriteCloser":     reflect.TypeOf(&writeCloser).Elem(),
		"WriteSeeker":     reflect.TypeOf(&writeSeeker).Elem(),
		"Writer":          reflect.TypeOf(&writer).Elem(),
		"WriterAt":        reflect.TypeOf(&writerAt).Elem(),
		"WriterTo":        reflect.TypeOf(&writerTo).Elem(),
	}
}

func initIOioutil() {
	env.Packages["io/ioutil"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Discard": reflect.ValueOf(ioutil.Discard),

		// define functions
		"NopCloser": reflect.ValueOf(ioutil.NopCloser),
		"ReadAll":   reflect.ValueOf(ioutil.ReadAll),
		"ReadDir":   reflect.ValueOf(ioutil.ReadDir),
		"ReadFile":  reflect.ValueOf(ioutil.ReadFile),
		"TempDir":   reflect.ValueOf(ioutil.TempDir),
		"TempFile":  reflect.ValueOf(ioutil.TempFile),
		"WriteFile": reflect.ValueOf(ioutil.WriteFile),
	}
	var ()
	env.PackageTypes["io/ioutil"] = map[string]reflect.Type{}
}

func initLog() {
	env.Packages["log"] = map[string]reflect.Value{
		// define constants
		"LUTC":          reflect.ValueOf(log.LUTC),
		"Ldate":         reflect.ValueOf(log.Ldate),
		"Llongfile":     reflect.ValueOf(log.Llongfile),
		"Lmicroseconds": reflect.ValueOf(log.Lmicroseconds),
		"Lmsgprefix":    reflect.ValueOf(log.Lmsgprefix),
		"Lshortfile":    reflect.ValueOf(log.Lshortfile),
		"LstdFlags":     reflect.ValueOf(log.LstdFlags),
		"Ltime":         reflect.ValueOf(log.Ltime),

		// define variables

		// define functions
		"Fatal":     reflect.ValueOf(log.Fatal),
		"Fatalf":    reflect.ValueOf(log.Fatalf),
		"Fatalln":   reflect.ValueOf(log.Fatalln),
		"Flags":     reflect.ValueOf(log.Flags),
		"New":       reflect.ValueOf(log.New),
		"Output":    reflect.ValueOf(log.Output),
		"Panic":     reflect.ValueOf(log.Panic),
		"Panicf":    reflect.ValueOf(log.Panicf),
		"Panicln":   reflect.ValueOf(log.Panicln),
		"Prefix":    reflect.ValueOf(log.Prefix),
		"Print":     reflect.ValueOf(log.Print),
		"Printf":    reflect.ValueOf(log.Printf),
		"Println":   reflect.ValueOf(log.Println),
		"SetFlags":  reflect.ValueOf(log.SetFlags),
		"SetOutput": reflect.ValueOf(log.SetOutput),
		"SetPrefix": reflect.ValueOf(log.SetPrefix),
		"Writer":    reflect.ValueOf(log.Writer),
	}
	var (
		logger log.Logger
	)
	env.PackageTypes["log"] = map[string]reflect.Type{
		"Logger": reflect.TypeOf(&logger).Elem(),
	}
}

func initMath() {
	env.Packages["math"] = map[string]reflect.Value{
		// define constants
		"E":                      reflect.ValueOf(math.E),
		"Ln10":                   reflect.ValueOf(math.Ln10),
		"Ln2":                    reflect.ValueOf(math.Ln2),
		"Log10E":                 reflect.ValueOf(math.Log10E),
		"Log2E":                  reflect.ValueOf(math.Log2E),
		"MaxFloat32":             reflect.ValueOf(math.MaxFloat32),
		"MaxFloat64":             reflect.ValueOf(math.MaxFloat64),
		"MaxInt16":               reflect.ValueOf(math.MaxInt16),
		"MaxInt32":               reflect.ValueOf(math.MaxInt32),
		"MaxInt64":               reflect.ValueOf(int64(math.MaxInt64)),
		"MaxInt8":                reflect.ValueOf(math.MaxInt8),
		"MaxUint16":              reflect.ValueOf(math.MaxUint16),
		"MaxUint32":              reflect.ValueOf(uint32(math.MaxUint32)),
		"MaxUint64":              reflect.ValueOf(uint64(math.MaxUint64)),
		"MaxUint8":               reflect.ValueOf(math.MaxUint8),
		"MinInt16":               reflect.ValueOf(math.MinInt16),
		"MinInt32":               reflect.ValueOf(math.MinInt32),
		"MinInt64":               reflect.ValueOf(int64(math.MinInt64)),
		"MinInt8":                reflect.ValueOf(math.MinInt8),
		"Phi":                    reflect.ValueOf(math.Phi),
		"Pi":                     reflect.ValueOf(math.Pi),
		"SmallestNonzeroFloat32": reflect.ValueOf(math.SmallestNonzeroFloat32),
		"SmallestNonzeroFloat64": reflect.ValueOf(math.SmallestNonzeroFloat64),
		"Sqrt2":                  reflect.ValueOf(math.Sqrt2),
		"SqrtE":                  reflect.ValueOf(math.SqrtE),
		"SqrtPhi":                reflect.ValueOf(math.SqrtPhi),
		"SqrtPi":                 reflect.ValueOf(math.SqrtPi),

		// define variables

		// define functions
		"Abs":             reflect.ValueOf(math.Abs),
		"Acos":            reflect.ValueOf(math.Acos),
		"Acosh":           reflect.ValueOf(math.Acosh),
		"Asin":            reflect.ValueOf(math.Asin),
		"Asinh":           reflect.ValueOf(math.Asinh),
		"Atan":            reflect.ValueOf(math.Atan),
		"Atan2":           reflect.ValueOf(math.Atan2),
		"Atanh":           reflect.ValueOf(math.Atanh),
		"Cbrt":            reflect.ValueOf(math.Cbrt),
		"Ceil":            reflect.ValueOf(math.Ceil),
		"Copysign":        reflect.ValueOf(math.Copysign),
		"Cos":             reflect.ValueOf(math.Cos),
		"Cosh":            reflect.ValueOf(math.Cosh),
		"Dim":             reflect.ValueOf(math.Dim),
		"Erf":             reflect.ValueOf(math.Erf),
		"Erfc":            reflect.ValueOf(math.Erfc),
		"Erfcinv":         reflect.ValueOf(math.Erfcinv),
		"Erfinv":          reflect.ValueOf(math.Erfinv),
		"Exp":             reflect.ValueOf(math.Exp),
		"Exp2":            reflect.ValueOf(math.Exp2),
		"Expm1":           reflect.ValueOf(math.Expm1),
		"FMA":             reflect.ValueOf(math.FMA),
		"Float32bits":     reflect.ValueOf(math.Float32bits),
		"Float32frombits": reflect.ValueOf(math.Float32frombits),
		"Float64bits":     reflect.ValueOf(math.Float64bits),
		"Float64frombits": reflect.ValueOf(math.Float64frombits),
		"Floor":           reflect.ValueOf(math.Floor),
		"Frexp":           reflect.ValueOf(math.Frexp),
		"Gamma":           reflect.ValueOf(math.Gamma),
		"Hypot":           reflect.ValueOf(math.Hypot),
		"Ilogb":           reflect.ValueOf(math.Ilogb),
		"Inf":             reflect.ValueOf(math.Inf),
		"IsInf":           reflect.ValueOf(math.IsInf),
		"IsNaN":           reflect.ValueOf(math.IsNaN),
		"J0":              reflect.ValueOf(math.J0),
		"J1":              reflect.ValueOf(math.J1),
		"Jn":              reflect.ValueOf(math.Jn),
		"Ldexp":           reflect.ValueOf(math.Ldexp),
		"Lgamma":          reflect.ValueOf(math.Lgamma),
		"Log":             reflect.ValueOf(math.Log),
		"Log10":           reflect.ValueOf(math.Log10),
		"Log1p":           reflect.ValueOf(math.Log1p),
		"Log2":            reflect.ValueOf(math.Log2),
		"Logb":            reflect.ValueOf(math.Logb),
		"Max":             reflect.ValueOf(math.Max),
		"Min":             reflect.ValueOf(math.Min),
		"Mod":             reflect.ValueOf(math.Mod),
		"Modf":            reflect.ValueOf(math.Modf),
		"NaN":             reflect.ValueOf(math.NaN),
		"Nextafter":       reflect.ValueOf(math.Nextafter),
		"Nextafter32":     reflect.ValueOf(math.Nextafter32),
		"Pow":             reflect.ValueOf(math.Pow),
		"Pow10":           reflect.ValueOf(math.Pow10),
		"Remainder":       reflect.ValueOf(math.Remainder),
		"Round":           reflect.ValueOf(math.Round),
		"RoundToEven":     reflect.ValueOf(math.RoundToEven),
		"Signbit":         reflect.ValueOf(math.Signbit),
		"Sin":             reflect.ValueOf(math.Sin),
		"Sincos":          reflect.ValueOf(math.Sincos),
		"Sinh":            reflect.ValueOf(math.Sinh),
		"Sqrt":            reflect.ValueOf(math.Sqrt),
		"Tan":             reflect.ValueOf(math.Tan),
		"Tanh":            reflect.ValueOf(math.Tanh),
		"Trunc":           reflect.ValueOf(math.Trunc),
		"Y0":              reflect.ValueOf(math.Y0),
		"Y1":              reflect.ValueOf(math.Y1),
		"Yn":              reflect.ValueOf(math.Yn),
	}
	var ()
	env.PackageTypes["math"] = map[string]reflect.Type{}
}

func initMathBig() {
	env.Packages["math/big"] = map[string]reflect.Value{
		// define constants
		"Above":         reflect.ValueOf(big.Above),
		"AwayFromZero":  reflect.ValueOf(big.AwayFromZero),
		"Below":         reflect.ValueOf(big.Below),
		"Exact":         reflect.ValueOf(big.Exact),
		"MaxBase":       reflect.ValueOf(big.MaxBase),
		"MaxExp":        reflect.ValueOf(big.MaxExp),
		"MaxPrec":       reflect.ValueOf(uint32(big.MaxPrec)),
		"MinExp":        reflect.ValueOf(big.MinExp),
		"ToNearestAway": reflect.ValueOf(big.ToNearestAway),
		"ToNearestEven": reflect.ValueOf(big.ToNearestEven),
		"ToNegativeInf": reflect.ValueOf(big.ToNegativeInf),
		"ToPositiveInf": reflect.ValueOf(big.ToPositiveInf),
		"ToZero":        reflect.ValueOf(big.ToZero),

		// define variables

		// define functions
		"Jacobi":     reflect.ValueOf(big.Jacobi),
		"NewFloat":   reflect.ValueOf(big.NewFloat),
		"NewInt":     reflect.ValueOf(big.NewInt),
		"NewRat":     reflect.ValueOf(big.NewRat),
		"ParseFloat": reflect.ValueOf(big.ParseFloat),
	}
	var (
		accuracy     big.Accuracy
		errNaN       big.ErrNaN
		float        big.Float
		i            big.Int
		rat          big.Rat
		roundingMode big.RoundingMode
		word         big.Word
	)
	env.PackageTypes["math/big"] = map[string]reflect.Type{
		"Accuracy":     reflect.TypeOf(&accuracy).Elem(),
		"ErrNaN":       reflect.TypeOf(&errNaN).Elem(),
		"Float":        reflect.TypeOf(&float).Elem(),
		"Int":          reflect.TypeOf(&i).Elem(),
		"Rat":          reflect.TypeOf(&rat).Elem(),
		"RoundingMode": reflect.TypeOf(&roundingMode).Elem(),
		"Word":         reflect.TypeOf(&word).Elem(),
	}
}

func initMathBits() {
	env.Packages["math/bits"] = map[string]reflect.Value{
		// define constants
		"UintSize": reflect.ValueOf(bits.UintSize),

		// define variables

		// define functions
		"Add":             reflect.ValueOf(bits.Add),
		"Add32":           reflect.ValueOf(bits.Add32),
		"Add64":           reflect.ValueOf(bits.Add64),
		"Div":             reflect.ValueOf(bits.Div),
		"Div32":           reflect.ValueOf(bits.Div32),
		"Div64":           reflect.ValueOf(bits.Div64),
		"LeadingZeros":    reflect.ValueOf(bits.LeadingZeros),
		"LeadingZeros16":  reflect.ValueOf(bits.LeadingZeros16),
		"LeadingZeros32":  reflect.ValueOf(bits.LeadingZeros32),
		"LeadingZeros64":  reflect.ValueOf(bits.LeadingZeros64),
		"LeadingZeros8":   reflect.ValueOf(bits.LeadingZeros8),
		"Len":             reflect.ValueOf(bits.Len),
		"Len16":           reflect.ValueOf(bits.Len16),
		"Len32":           reflect.ValueOf(bits.Len32),
		"Len64":           reflect.ValueOf(bits.Len64),
		"Len8":            reflect.ValueOf(bits.Len8),
		"Mul":             reflect.ValueOf(bits.Mul),
		"Mul32":           reflect.ValueOf(bits.Mul32),
		"Mul64":           reflect.ValueOf(bits.Mul64),
		"OnesCount":       reflect.ValueOf(bits.OnesCount),
		"OnesCount16":     reflect.ValueOf(bits.OnesCount16),
		"OnesCount32":     reflect.ValueOf(bits.OnesCount32),
		"OnesCount64":     reflect.ValueOf(bits.OnesCount64),
		"OnesCount8":      reflect.ValueOf(bits.OnesCount8),
		"Rem":             reflect.ValueOf(bits.Rem),
		"Rem32":           reflect.ValueOf(bits.Rem32),
		"Rem64":           reflect.ValueOf(bits.Rem64),
		"Reverse":         reflect.ValueOf(bits.Reverse),
		"Reverse16":       reflect.ValueOf(bits.Reverse16),
		"Reverse32":       reflect.ValueOf(bits.Reverse32),
		"Reverse64":       reflect.ValueOf(bits.Reverse64),
		"Reverse8":        reflect.ValueOf(bits.Reverse8),
		"ReverseBytes":    reflect.ValueOf(bits.ReverseBytes),
		"ReverseBytes16":  reflect.ValueOf(bits.ReverseBytes16),
		"ReverseBytes32":  reflect.ValueOf(bits.ReverseBytes32),
		"ReverseBytes64":  reflect.ValueOf(bits.ReverseBytes64),
		"RotateLeft":      reflect.ValueOf(bits.RotateLeft),
		"RotateLeft16":    reflect.ValueOf(bits.RotateLeft16),
		"RotateLeft32":    reflect.ValueOf(bits.RotateLeft32),
		"RotateLeft64":    reflect.ValueOf(bits.RotateLeft64),
		"RotateLeft8":     reflect.ValueOf(bits.RotateLeft8),
		"Sub":             reflect.ValueOf(bits.Sub),
		"Sub32":           reflect.ValueOf(bits.Sub32),
		"Sub64":           reflect.ValueOf(bits.Sub64),
		"TrailingZeros":   reflect.ValueOf(bits.TrailingZeros),
		"TrailingZeros16": reflect.ValueOf(bits.TrailingZeros16),
		"TrailingZeros32": reflect.ValueOf(bits.TrailingZeros32),
		"TrailingZeros64": reflect.ValueOf(bits.TrailingZeros64),
		"TrailingZeros8":  reflect.ValueOf(bits.TrailingZeros8),
	}
	var ()
	env.PackageTypes["math/bits"] = map[string]reflect.Type{}
}

func initMathCmplx() {
	env.Packages["math/cmplx"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Abs":   reflect.ValueOf(cmplx.Abs),
		"Acos":  reflect.ValueOf(cmplx.Acos),
		"Acosh": reflect.ValueOf(cmplx.Acosh),
		"Asin":  reflect.ValueOf(cmplx.Asin),
		"Asinh": reflect.ValueOf(cmplx.Asinh),
		"Atan":  reflect.ValueOf(cmplx.Atan),
		"Atanh": reflect.ValueOf(cmplx.Atanh),
		"Conj":  reflect.ValueOf(cmplx.Conj),
		"Cos":   reflect.ValueOf(cmplx.Cos),
		"Cosh":  reflect.ValueOf(cmplx.Cosh),
		"Cot":   reflect.ValueOf(cmplx.Cot),
		"Exp":   reflect.ValueOf(cmplx.Exp),
		"Inf":   reflect.ValueOf(cmplx.Inf),
		"IsInf": reflect.ValueOf(cmplx.IsInf),
		"IsNaN": reflect.ValueOf(cmplx.IsNaN),
		"Log":   reflect.ValueOf(cmplx.Log),
		"Log10": reflect.ValueOf(cmplx.Log10),
		"NaN":   reflect.ValueOf(cmplx.NaN),
		"Phase": reflect.ValueOf(cmplx.Phase),
		"Polar": reflect.ValueOf(cmplx.Polar),
		"Pow":   reflect.ValueOf(cmplx.Pow),
		"Rect":  reflect.ValueOf(cmplx.Rect),
		"Sin":   reflect.ValueOf(cmplx.Sin),
		"Sinh":  reflect.ValueOf(cmplx.Sinh),
		"Sqrt":  reflect.ValueOf(cmplx.Sqrt),
		"Tan":   reflect.ValueOf(cmplx.Tan),
		"Tanh":  reflect.ValueOf(cmplx.Tanh),
	}
	var ()
	env.PackageTypes["math/cmplx"] = map[string]reflect.Type{}
}

func initMathRand() {
	env.Packages["math/rand"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"ExpFloat64":  reflect.ValueOf(rand.ExpFloat64),
		"Float32":     reflect.ValueOf(rand.Float32),
		"Float64":     reflect.ValueOf(rand.Float64),
		"Int":         reflect.ValueOf(rand.Int),
		"Int31":       reflect.ValueOf(rand.Int31),
		"Int31n":      reflect.ValueOf(rand.Int31n),
		"Int63":       reflect.ValueOf(rand.Int63),
		"Int63n":      reflect.ValueOf(rand.Int63n),
		"Intn":        reflect.ValueOf(rand.Intn),
		"New":         reflect.ValueOf(rand.New),
		"NewSource":   reflect.ValueOf(rand.NewSource),
		"NewZipf":     reflect.ValueOf(rand.NewZipf),
		"NormFloat64": reflect.ValueOf(rand.NormFloat64),
		"Perm":        reflect.ValueOf(rand.Perm),
		"Read":        reflect.ValueOf(rand.Read),
		"Seed":        reflect.ValueOf(rand.Seed),
		"Shuffle":     reflect.ValueOf(rand.Shuffle),
		"Uint32":      reflect.ValueOf(rand.Uint32),
		"Uint64":      reflect.ValueOf(rand.Uint64),
	}
	var (
		r        rand.Rand
		source   rand.Source
		source64 rand.Source64
		zipf     rand.Zipf
	)
	env.PackageTypes["math/rand"] = map[string]reflect.Type{
		"Rand":     reflect.TypeOf(&r).Elem(),
		"Source":   reflect.TypeOf(&source).Elem(),
		"Source64": reflect.TypeOf(&source64).Elem(),
		"Zipf":     reflect.TypeOf(&zipf).Elem(),
	}
}

func initMIME() {
	env.Packages["mime"] = map[string]reflect.Value{
		// define constants
		"BEncoding": reflect.ValueOf(mime.BEncoding),
		"QEncoding": reflect.ValueOf(mime.QEncoding),

		// define variables
		"ErrInvalidMediaParameter": reflect.ValueOf(mime.ErrInvalidMediaParameter),

		// define functions
		"AddExtensionType": reflect.ValueOf(mime.AddExtensionType),
		"ExtensionsByType": reflect.ValueOf(mime.ExtensionsByType),
		"FormatMediaType":  reflect.ValueOf(mime.FormatMediaType),
		"ParseMediaType":   reflect.ValueOf(mime.ParseMediaType),
		"TypeByExtension":  reflect.ValueOf(mime.TypeByExtension),
	}
	var (
		wordDecoder mime.WordDecoder
		wordEncoder mime.WordEncoder
	)
	env.PackageTypes["mime"] = map[string]reflect.Type{
		"WordDecoder": reflect.TypeOf(&wordDecoder).Elem(),
		"WordEncoder": reflect.TypeOf(&wordEncoder).Elem(),
	}
}

func initMIMEMultiPart() {
	env.Packages["mime/multipart"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrMessageTooLarge": reflect.ValueOf(multipart.ErrMessageTooLarge),

		// define functions
		"NewReader": reflect.ValueOf(multipart.NewReader),
		"NewWriter": reflect.ValueOf(multipart.NewWriter),
	}
	var (
		file       multipart.File
		fileHeader multipart.FileHeader
		form       multipart.Form
		part       multipart.Part
		reader     multipart.Reader
		writer     multipart.Writer
	)
	env.PackageTypes["mime/multipart"] = map[string]reflect.Type{
		"File":       reflect.TypeOf(&file).Elem(),
		"FileHeader": reflect.TypeOf(&fileHeader).Elem(),
		"Form":       reflect.TypeOf(&form).Elem(),
		"Part":       reflect.TypeOf(&part).Elem(),
		"Reader":     reflect.TypeOf(&reader).Elem(),
		"Writer":     reflect.TypeOf(&writer).Elem(),
	}
}

func initMIMEQuotedPrintable() {
	env.Packages["mime/quotedprintable"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewReader": reflect.ValueOf(quotedprintable.NewReader),
		"NewWriter": reflect.ValueOf(quotedprintable.NewWriter),
	}
	var (
		reader quotedprintable.Reader
		writer quotedprintable.Writer
	)
	env.PackageTypes["mime/quotedprintable"] = map[string]reflect.Type{
		"Reader": reflect.TypeOf(&reader).Elem(),
		"Writer": reflect.TypeOf(&writer).Elem(),
	}
}

func initNet() {
	env.Packages["net"] = map[string]reflect.Value{
		// define constants
		"FlagBroadcast":    reflect.ValueOf(net.FlagBroadcast),
		"FlagLoopback":     reflect.ValueOf(net.FlagLoopback),
		"FlagMulticast":    reflect.ValueOf(net.FlagMulticast),
		"FlagPointToPoint": reflect.ValueOf(net.FlagPointToPoint),
		"FlagUp":           reflect.ValueOf(net.FlagUp),
		"IPv4len":          reflect.ValueOf(net.IPv4len),
		"IPv6len":          reflect.ValueOf(net.IPv6len),

		// define variables
		"DefaultResolver":            reflect.ValueOf(net.DefaultResolver),
		"ErrWriteToConnected":        reflect.ValueOf(net.ErrWriteToConnected),
		"IPv4allrouter":              reflect.ValueOf(net.IPv4allrouter),
		"IPv4allsys":                 reflect.ValueOf(net.IPv4allsys),
		"IPv4bcast":                  reflect.ValueOf(net.IPv4bcast),
		"IPv4zero":                   reflect.ValueOf(net.IPv4zero),
		"IPv6interfacelocalallnodes": reflect.ValueOf(net.IPv6interfacelocalallnodes),
		"IPv6linklocalallnodes":      reflect.ValueOf(net.IPv6linklocalallnodes),
		"IPv6linklocalallrouters":    reflect.ValueOf(net.IPv6linklocalallrouters),
		"IPv6loopback":               reflect.ValueOf(net.IPv6loopback),
		"IPv6unspecified":            reflect.ValueOf(net.IPv6unspecified),
		"IPv6zero":                   reflect.ValueOf(net.IPv6zero),

		// define functions
		"CIDRMask":           reflect.ValueOf(net.CIDRMask),
		"Dial":               reflect.ValueOf(net.Dial),
		"DialIP":             reflect.ValueOf(net.DialIP),
		"DialTCP":            reflect.ValueOf(net.DialTCP),
		"DialTimeout":        reflect.ValueOf(net.DialTimeout),
		"DialUDP":            reflect.ValueOf(net.DialUDP),
		"DialUnix":           reflect.ValueOf(net.DialUnix),
		"FileConn":           reflect.ValueOf(net.FileConn),
		"FileListener":       reflect.ValueOf(net.FileListener),
		"FilePacketConn":     reflect.ValueOf(net.FilePacketConn),
		"IPv4":               reflect.ValueOf(net.IPv4),
		"IPv4Mask":           reflect.ValueOf(net.IPv4Mask),
		"InterfaceAddrs":     reflect.ValueOf(net.InterfaceAddrs),
		"InterfaceByIndex":   reflect.ValueOf(net.InterfaceByIndex),
		"InterfaceByName":    reflect.ValueOf(net.InterfaceByName),
		"Interfaces":         reflect.ValueOf(net.Interfaces),
		"JoinHostPort":       reflect.ValueOf(net.JoinHostPort),
		"Listen":             reflect.ValueOf(net.Listen),
		"ListenIP":           reflect.ValueOf(net.ListenIP),
		"ListenMulticastUDP": reflect.ValueOf(net.ListenMulticastUDP),
		"ListenPacket":       reflect.ValueOf(net.ListenPacket),
		"ListenTCP":          reflect.ValueOf(net.ListenTCP),
		"ListenUDP":          reflect.ValueOf(net.ListenUDP),
		"ListenUnix":         reflect.ValueOf(net.ListenUnix),
		"ListenUnixgram":     reflect.ValueOf(net.ListenUnixgram),
		"LookupAddr":         reflect.ValueOf(net.LookupAddr),
		"LookupCNAME":        reflect.ValueOf(net.LookupCNAME),
		"LookupHost":         reflect.ValueOf(net.LookupHost),
		"LookupIP":           reflect.ValueOf(net.LookupIP),
		"LookupMX":           reflect.ValueOf(net.LookupMX),
		"LookupNS":           reflect.ValueOf(net.LookupNS),
		"LookupPort":         reflect.ValueOf(net.LookupPort),
		"LookupSRV":          reflect.ValueOf(net.LookupSRV),
		"LookupTXT":          reflect.ValueOf(net.LookupTXT),
		"ParseCIDR":          reflect.ValueOf(net.ParseCIDR),
		"ParseIP":            reflect.ValueOf(net.ParseIP),
		"ParseMAC":           reflect.ValueOf(net.ParseMAC),
		"Pipe":               reflect.ValueOf(net.Pipe),
		"ResolveIPAddr":      reflect.ValueOf(net.ResolveIPAddr),
		"ResolveTCPAddr":     reflect.ValueOf(net.ResolveTCPAddr),
		"ResolveUDPAddr":     reflect.ValueOf(net.ResolveUDPAddr),
		"ResolveUnixAddr":    reflect.ValueOf(net.ResolveUnixAddr),
		"SplitHostPort":      reflect.ValueOf(net.SplitHostPort),
	}
	var (
		addr                net.Addr
		addrError           net.AddrError
		buffers             net.Buffers
		conn                net.Conn
		dNSConfigError      net.DNSConfigError
		dNSError            net.DNSError
		dialer              net.Dialer
		err                 net.Error
		flags               net.Flags
		hardwareAddr        net.HardwareAddr
		iP                  net.IP
		iPAddr              net.IPAddr
		iPConn              net.IPConn
		iPMask              net.IPMask
		iPNet               net.IPNet
		iface               net.Interface
		invalidAddrError    net.InvalidAddrError
		listenConfig        net.ListenConfig
		listener            net.Listener
		mX                  net.MX
		nS                  net.NS
		opError             net.OpError
		packetConn          net.PacketConn
		parseError          net.ParseError
		resolver            net.Resolver
		sRV                 net.SRV
		tCPAddr             net.TCPAddr
		tCPConn             net.TCPConn
		tCPListener         net.TCPListener
		uDPAddr             net.UDPAddr
		uDPConn             net.UDPConn
		unixAddr            net.UnixAddr
		unixConn            net.UnixConn
		unixListener        net.UnixListener
		unknownNetworkError net.UnknownNetworkError
	)
	env.PackageTypes["net"] = map[string]reflect.Type{
		"Addr":                reflect.TypeOf(&addr).Elem(),
		"AddrError":           reflect.TypeOf(&addrError).Elem(),
		"Buffers":             reflect.TypeOf(&buffers).Elem(),
		"Conn":                reflect.TypeOf(&conn).Elem(),
		"DNSConfigError":      reflect.TypeOf(&dNSConfigError).Elem(),
		"DNSError":            reflect.TypeOf(&dNSError).Elem(),
		"Dialer":              reflect.TypeOf(&dialer).Elem(),
		"Error":               reflect.TypeOf(&err).Elem(),
		"Flags":               reflect.TypeOf(&flags).Elem(),
		"HardwareAddr":        reflect.TypeOf(&hardwareAddr).Elem(),
		"IP":                  reflect.TypeOf(&iP).Elem(),
		"IPAddr":              reflect.TypeOf(&iPAddr).Elem(),
		"IPConn":              reflect.TypeOf(&iPConn).Elem(),
		"IPMask":              reflect.TypeOf(&iPMask).Elem(),
		"IPNet":               reflect.TypeOf(&iPNet).Elem(),
		"Interface":           reflect.TypeOf(&iface).Elem(),
		"InvalidAddrError":    reflect.TypeOf(&invalidAddrError).Elem(),
		"ListenConfig":        reflect.TypeOf(&listenConfig).Elem(),
		"Listener":            reflect.TypeOf(&listener).Elem(),
		"MX":                  reflect.TypeOf(&mX).Elem(),
		"NS":                  reflect.TypeOf(&nS).Elem(),
		"OpError":             reflect.TypeOf(&opError).Elem(),
		"PacketConn":          reflect.TypeOf(&packetConn).Elem(),
		"ParseError":          reflect.TypeOf(&parseError).Elem(),
		"Resolver":            reflect.TypeOf(&resolver).Elem(),
		"SRV":                 reflect.TypeOf(&sRV).Elem(),
		"TCPAddr":             reflect.TypeOf(&tCPAddr).Elem(),
		"TCPConn":             reflect.TypeOf(&tCPConn).Elem(),
		"TCPListener":         reflect.TypeOf(&tCPListener).Elem(),
		"UDPAddr":             reflect.TypeOf(&uDPAddr).Elem(),
		"UDPConn":             reflect.TypeOf(&uDPConn).Elem(),
		"UnixAddr":            reflect.TypeOf(&unixAddr).Elem(),
		"UnixConn":            reflect.TypeOf(&unixConn).Elem(),
		"UnixListener":        reflect.TypeOf(&unixListener).Elem(),
		"UnknownNetworkError": reflect.TypeOf(&unknownNetworkError).Elem(),
	}
}

func initNetHTTP() {
	env.Packages["net/http"] = map[string]reflect.Value{
		// define constants
		"DefaultMaxHeaderBytes":               reflect.ValueOf(http.DefaultMaxHeaderBytes),
		"DefaultMaxIdleConnsPerHost":          reflect.ValueOf(http.DefaultMaxIdleConnsPerHost),
		"MethodConnect":                       reflect.ValueOf(http.MethodConnect),
		"MethodDelete":                        reflect.ValueOf(http.MethodDelete),
		"MethodGet":                           reflect.ValueOf(http.MethodGet),
		"MethodHead":                          reflect.ValueOf(http.MethodHead),
		"MethodOptions":                       reflect.ValueOf(http.MethodOptions),
		"MethodPatch":                         reflect.ValueOf(http.MethodPatch),
		"MethodPost":                          reflect.ValueOf(http.MethodPost),
		"MethodPut":                           reflect.ValueOf(http.MethodPut),
		"MethodTrace":                         reflect.ValueOf(http.MethodTrace),
		"SameSiteDefaultMode":                 reflect.ValueOf(http.SameSiteDefaultMode),
		"SameSiteLaxMode":                     reflect.ValueOf(http.SameSiteLaxMode),
		"SameSiteNoneMode":                    reflect.ValueOf(http.SameSiteNoneMode),
		"SameSiteStrictMode":                  reflect.ValueOf(http.SameSiteStrictMode),
		"StateActive":                         reflect.ValueOf(http.StateActive),
		"StateClosed":                         reflect.ValueOf(http.StateClosed),
		"StateHijacked":                       reflect.ValueOf(http.StateHijacked),
		"StateIdle":                           reflect.ValueOf(http.StateIdle),
		"StateNew":                            reflect.ValueOf(http.StateNew),
		"StatusAccepted":                      reflect.ValueOf(http.StatusAccepted),
		"StatusAlreadyReported":               reflect.ValueOf(http.StatusAlreadyReported),
		"StatusBadGateway":                    reflect.ValueOf(http.StatusBadGateway),
		"StatusBadRequest":                    reflect.ValueOf(http.StatusBadRequest),
		"StatusConflict":                      reflect.ValueOf(http.StatusConflict),
		"StatusContinue":                      reflect.ValueOf(http.StatusContinue),
		"StatusCreated":                       reflect.ValueOf(http.StatusCreated),
		"StatusEarlyHints":                    reflect.ValueOf(http.StatusEarlyHints),
		"StatusExpectationFailed":             reflect.ValueOf(http.StatusExpectationFailed),
		"StatusFailedDependency":              reflect.ValueOf(http.StatusFailedDependency),
		"StatusForbidden":                     reflect.ValueOf(http.StatusForbidden),
		"StatusFound":                         reflect.ValueOf(http.StatusFound),
		"StatusGatewayTimeout":                reflect.ValueOf(http.StatusGatewayTimeout),
		"StatusGone":                          reflect.ValueOf(http.StatusGone),
		"StatusHTTPVersionNotSupported":       reflect.ValueOf(http.StatusHTTPVersionNotSupported),
		"StatusIMUsed":                        reflect.ValueOf(http.StatusIMUsed),
		"StatusInsufficientStorage":           reflect.ValueOf(http.StatusInsufficientStorage),
		"StatusInternalServerError":           reflect.ValueOf(http.StatusInternalServerError),
		"StatusLengthRequired":                reflect.ValueOf(http.StatusLengthRequired),
		"StatusLocked":                        reflect.ValueOf(http.StatusLocked),
		"StatusLoopDetected":                  reflect.ValueOf(http.StatusLoopDetected),
		"StatusMethodNotAllowed":              reflect.ValueOf(http.StatusMethodNotAllowed),
		"StatusMisdirectedRequest":            reflect.ValueOf(http.StatusMisdirectedRequest),
		"StatusMovedPermanently":              reflect.ValueOf(http.StatusMovedPermanently),
		"StatusMultiStatus":                   reflect.ValueOf(http.StatusMultiStatus),
		"StatusMultipleChoices":               reflect.ValueOf(http.StatusMultipleChoices),
		"StatusNetworkAuthenticationRequired": reflect.ValueOf(http.StatusNetworkAuthenticationRequired),
		"StatusNoContent":                     reflect.ValueOf(http.StatusNoContent),
		"StatusNonAuthoritativeInfo":          reflect.ValueOf(http.StatusNonAuthoritativeInfo),
		"StatusNotAcceptable":                 reflect.ValueOf(http.StatusNotAcceptable),
		"StatusNotExtended":                   reflect.ValueOf(http.StatusNotExtended),
		"StatusNotFound":                      reflect.ValueOf(http.StatusNotFound),
		"StatusNotImplemented":                reflect.ValueOf(http.StatusNotImplemented),
		"StatusNotModified":                   reflect.ValueOf(http.StatusNotModified),
		"StatusOK":                            reflect.ValueOf(http.StatusOK),
		"StatusPartialContent":                reflect.ValueOf(http.StatusPartialContent),
		"StatusPaymentRequired":               reflect.ValueOf(http.StatusPaymentRequired),
		"StatusPermanentRedirect":             reflect.ValueOf(http.StatusPermanentRedirect),
		"StatusPreconditionFailed":            reflect.ValueOf(http.StatusPreconditionFailed),
		"StatusPreconditionRequired":          reflect.ValueOf(http.StatusPreconditionRequired),
		"StatusProcessing":                    reflect.ValueOf(http.StatusProcessing),
		"StatusProxyAuthRequired":             reflect.ValueOf(http.StatusProxyAuthRequired),
		"StatusRequestEntityTooLarge":         reflect.ValueOf(http.StatusRequestEntityTooLarge),
		"StatusRequestHeaderFieldsTooLarge":   reflect.ValueOf(http.StatusRequestHeaderFieldsTooLarge),
		"StatusRequestTimeout":                reflect.ValueOf(http.StatusRequestTimeout),
		"StatusRequestURITooLong":             reflect.ValueOf(http.StatusRequestURITooLong),
		"StatusRequestedRangeNotSatisfiable":  reflect.ValueOf(http.StatusRequestedRangeNotSatisfiable),
		"StatusResetContent":                  reflect.ValueOf(http.StatusResetContent),
		"StatusSeeOther":                      reflect.ValueOf(http.StatusSeeOther),
		"StatusServiceUnavailable":            reflect.ValueOf(http.StatusServiceUnavailable),
		"StatusSwitchingProtocols":            reflect.ValueOf(http.StatusSwitchingProtocols),
		"StatusTeapot":                        reflect.ValueOf(http.StatusTeapot),
		"StatusTemporaryRedirect":             reflect.ValueOf(http.StatusTemporaryRedirect),
		"StatusTooEarly":                      reflect.ValueOf(http.StatusTooEarly),
		"StatusTooManyRequests":               reflect.ValueOf(http.StatusTooManyRequests),
		"StatusUnauthorized":                  reflect.ValueOf(http.StatusUnauthorized),
		"StatusUnavailableForLegalReasons":    reflect.ValueOf(http.StatusUnavailableForLegalReasons),
		"StatusUnprocessableEntity":           reflect.ValueOf(http.StatusUnprocessableEntity),
		"StatusUnsupportedMediaType":          reflect.ValueOf(http.StatusUnsupportedMediaType),
		"StatusUpgradeRequired":               reflect.ValueOf(http.StatusUpgradeRequired),
		"StatusUseProxy":                      reflect.ValueOf(http.StatusUseProxy),
		"StatusVariantAlsoNegotiates":         reflect.ValueOf(http.StatusVariantAlsoNegotiates),
		"TimeFormat":                          reflect.ValueOf(http.TimeFormat),
		"TrailerPrefix":                       reflect.ValueOf(http.TrailerPrefix),

		// define variables
		"DefaultClient":         reflect.ValueOf(http.DefaultClient),
		"DefaultServeMux":       reflect.ValueOf(http.DefaultServeMux),
		"DefaultTransport":      reflect.ValueOf(http.DefaultTransport),
		"ErrAbortHandler":       reflect.ValueOf(http.ErrAbortHandler),
		"ErrBodyNotAllowed":     reflect.ValueOf(http.ErrBodyNotAllowed),
		"ErrBodyReadAfterClose": reflect.ValueOf(http.ErrBodyReadAfterClose),
		"ErrContentLength":      reflect.ValueOf(http.ErrContentLength),
		"ErrHandlerTimeout":     reflect.ValueOf(http.ErrHandlerTimeout),
		"ErrHijacked":           reflect.ValueOf(http.ErrHijacked),
		"ErrLineTooLong":        reflect.ValueOf(http.ErrLineTooLong),
		"ErrMissingBoundary":    reflect.ValueOf(http.ErrMissingBoundary),
		"ErrMissingFile":        reflect.ValueOf(http.ErrMissingFile),
		"ErrNoCookie":           reflect.ValueOf(http.ErrNoCookie),
		"ErrNoLocation":         reflect.ValueOf(http.ErrNoLocation),
		"ErrNotMultipart":       reflect.ValueOf(http.ErrNotMultipart),
		"ErrNotSupported":       reflect.ValueOf(http.ErrNotSupported),
		"ErrServerClosed":       reflect.ValueOf(http.ErrServerClosed),
		"ErrSkipAltProtocol":    reflect.ValueOf(http.ErrSkipAltProtocol),
		"ErrUseLastResponse":    reflect.ValueOf(http.ErrUseLastResponse),
		"LocalAddrContextKey":   reflect.ValueOf(http.LocalAddrContextKey),
		"NoBody":                reflect.ValueOf(http.NoBody),
		"ServerContextKey":      reflect.ValueOf(http.ServerContextKey),

		// define functions
		"CanonicalHeaderKey":    reflect.ValueOf(http.CanonicalHeaderKey),
		"DetectContentType":     reflect.ValueOf(http.DetectContentType),
		"Error":                 reflect.ValueOf(http.Error),
		"FileServer":            reflect.ValueOf(http.FileServer),
		"Get":                   reflect.ValueOf(http.Get),
		"Handle":                reflect.ValueOf(http.Handle),
		"HandleFunc":            reflect.ValueOf(http.HandleFunc),
		"Head":                  reflect.ValueOf(http.Head),
		"ListenAndServe":        reflect.ValueOf(http.ListenAndServe),
		"ListenAndServeTLS":     reflect.ValueOf(http.ListenAndServeTLS),
		"MaxBytesReader":        reflect.ValueOf(http.MaxBytesReader),
		"NewFileTransport":      reflect.ValueOf(http.NewFileTransport),
		"NewRequest":            reflect.ValueOf(http.NewRequest),
		"NewRequestWithContext": reflect.ValueOf(http.NewRequestWithContext),
		"NewServeMux":           reflect.ValueOf(http.NewServeMux),
		"NotFound":              reflect.ValueOf(http.NotFound),
		"NotFoundHandler":       reflect.ValueOf(http.NotFoundHandler),
		"ParseHTTPVersion":      reflect.ValueOf(http.ParseHTTPVersion),
		"ParseTime":             reflect.ValueOf(http.ParseTime),
		"Post":                  reflect.ValueOf(http.Post),
		"PostForm":              reflect.ValueOf(http.PostForm),
		"ProxyFromEnvironment":  reflect.ValueOf(http.ProxyFromEnvironment),
		"ProxyURL":              reflect.ValueOf(http.ProxyURL),
		"ReadRequest":           reflect.ValueOf(http.ReadRequest),
		"ReadResponse":          reflect.ValueOf(http.ReadResponse),
		"Redirect":              reflect.ValueOf(http.Redirect),
		"RedirectHandler":       reflect.ValueOf(http.RedirectHandler),
		"Serve":                 reflect.ValueOf(http.Serve),
		"ServeContent":          reflect.ValueOf(http.ServeContent),
		"ServeFile":             reflect.ValueOf(http.ServeFile),
		"ServeTLS":              reflect.ValueOf(http.ServeTLS),
		"SetCookie":             reflect.ValueOf(http.SetCookie),
		"StatusText":            reflect.ValueOf(http.StatusText),
		"StripPrefix":           reflect.ValueOf(http.StripPrefix),
		"TimeoutHandler":        reflect.ValueOf(http.TimeoutHandler),
	}
	var (
		client         http.Client
		connState      http.ConnState
		cookie         http.Cookie
		cookieJar      http.CookieJar
		dir            http.Dir
		file           http.File
		fileSystem     http.FileSystem
		flusher        http.Flusher
		handler        http.Handler
		handlerFunc    http.HandlerFunc
		header         http.Header
		hijacker       http.Hijacker
		pushOptions    http.PushOptions
		pusher         http.Pusher
		request        http.Request
		response       http.Response
		responseWriter http.ResponseWriter
		roundTripper   http.RoundTripper
		sameSite       http.SameSite
		serveMux       http.ServeMux
		server         http.Server
		transport      http.Transport
	)
	env.PackageTypes["net/http"] = map[string]reflect.Type{
		"Client":         reflect.TypeOf(&client).Elem(),
		"ConnState":      reflect.TypeOf(&connState).Elem(),
		"Cookie":         reflect.TypeOf(&cookie).Elem(),
		"CookieJar":      reflect.TypeOf(&cookieJar).Elem(),
		"Dir":            reflect.TypeOf(&dir).Elem(),
		"File":           reflect.TypeOf(&file).Elem(),
		"FileSystem":     reflect.TypeOf(&fileSystem).Elem(),
		"Flusher":        reflect.TypeOf(&flusher).Elem(),
		"Handler":        reflect.TypeOf(&handler).Elem(),
		"HandlerFunc":    reflect.TypeOf(&handlerFunc).Elem(),
		"Header":         reflect.TypeOf(&header).Elem(),
		"Hijacker":       reflect.TypeOf(&hijacker).Elem(),
		"PushOptions":    reflect.TypeOf(&pushOptions).Elem(),
		"Pusher":         reflect.TypeOf(&pusher).Elem(),
		"Request":        reflect.TypeOf(&request).Elem(),
		"Response":       reflect.TypeOf(&response).Elem(),
		"ResponseWriter": reflect.TypeOf(&responseWriter).Elem(),
		"RoundTripper":   reflect.TypeOf(&roundTripper).Elem(),
		"SameSite":       reflect.TypeOf(&sameSite).Elem(),
		"ServeMux":       reflect.TypeOf(&serveMux).Elem(),
		"Server":         reflect.TypeOf(&server).Elem(),
		"Transport":      reflect.TypeOf(&transport).Elem(),
	}
}

func initNetHTTPCookieJar() {
	env.Packages["net/http/cookiejar"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"New": reflect.ValueOf(cookiejar.New),
	}
	var (
		jar              cookiejar.Jar
		options          cookiejar.Options
		publicSuffixList cookiejar.PublicSuffixList
	)
	env.PackageTypes["net/http/cookiejar"] = map[string]reflect.Type{
		"Jar":              reflect.TypeOf(&jar).Elem(),
		"Options":          reflect.TypeOf(&options).Elem(),
		"PublicSuffixList": reflect.TypeOf(&publicSuffixList).Elem(),
	}
}

func initNetMail() {
	env.Packages["net/mail"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrHeaderNotPresent": reflect.ValueOf(mail.ErrHeaderNotPresent),

		// define functions
		"ParseAddress":     reflect.ValueOf(mail.ParseAddress),
		"ParseAddressList": reflect.ValueOf(mail.ParseAddressList),
		"ParseDate":        reflect.ValueOf(mail.ParseDate),
		"ReadMessage":      reflect.ValueOf(mail.ReadMessage),
	}
	var (
		address       mail.Address
		addressParser mail.AddressParser
		header        mail.Header
		message       mail.Message
	)
	env.PackageTypes["net/mail"] = map[string]reflect.Type{
		"Address":       reflect.TypeOf(&address).Elem(),
		"AddressParser": reflect.TypeOf(&addressParser).Elem(),
		"Header":        reflect.TypeOf(&header).Elem(),
		"Message":       reflect.TypeOf(&message).Elem(),
	}
}

func initNetSMTP() {
	env.Packages["net/smtp"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CRAMMD5Auth": reflect.ValueOf(smtp.CRAMMD5Auth),
		"Dial":        reflect.ValueOf(smtp.Dial),
		"NewClient":   reflect.ValueOf(smtp.NewClient),
		"PlainAuth":   reflect.ValueOf(smtp.PlainAuth),
		"SendMail":    reflect.ValueOf(smtp.SendMail),
	}
	var (
		auth       smtp.Auth
		client     smtp.Client
		serverInfo smtp.ServerInfo
	)
	env.PackageTypes["net/smtp"] = map[string]reflect.Type{
		"Auth":       reflect.TypeOf(&auth).Elem(),
		"Client":     reflect.TypeOf(&client).Elem(),
		"ServerInfo": reflect.TypeOf(&serverInfo).Elem(),
	}
}

func initNetTextProto() {
	env.Packages["net/textproto"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CanonicalMIMEHeaderKey": reflect.ValueOf(textproto.CanonicalMIMEHeaderKey),
		"Dial":                   reflect.ValueOf(textproto.Dial),
		"NewConn":                reflect.ValueOf(textproto.NewConn),
		"NewReader":              reflect.ValueOf(textproto.NewReader),
		"NewWriter":              reflect.ValueOf(textproto.NewWriter),
		"TrimBytes":              reflect.ValueOf(textproto.TrimBytes),
		"TrimString":             reflect.ValueOf(textproto.TrimString),
	}
	var (
		conn          textproto.Conn
		err           textproto.Error
		mIMEHeader    textproto.MIMEHeader
		pipeline      textproto.Pipeline
		protocolError textproto.ProtocolError
		reader        textproto.Reader
		writer        textproto.Writer
	)
	env.PackageTypes["net/textproto"] = map[string]reflect.Type{
		"Conn":          reflect.TypeOf(&conn).Elem(),
		"Error":         reflect.TypeOf(&err).Elem(),
		"MIMEHeader":    reflect.TypeOf(&mIMEHeader).Elem(),
		"Pipeline":      reflect.TypeOf(&pipeline).Elem(),
		"ProtocolError": reflect.TypeOf(&protocolError).Elem(),
		"Reader":        reflect.TypeOf(&reader).Elem(),
		"Writer":        reflect.TypeOf(&writer).Elem(),
	}
}

func initNetURL() {
	env.Packages["net/url"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Parse":           reflect.ValueOf(url.Parse),
		"ParseQuery":      reflect.ValueOf(url.ParseQuery),
		"ParseRequestURI": reflect.ValueOf(url.ParseRequestURI),
		"PathEscape":      reflect.ValueOf(url.PathEscape),
		"PathUnescape":    reflect.ValueOf(url.PathUnescape),
		"QueryEscape":     reflect.ValueOf(url.QueryEscape),
		"QueryUnescape":   reflect.ValueOf(url.QueryUnescape),
		"User":            reflect.ValueOf(url.User),
		"UserPassword":    reflect.ValueOf(url.UserPassword),
	}
	var (
		err              url.Error
		escapeError      url.EscapeError
		invalidHostError url.InvalidHostError
		uRL              url.URL
		userinfo         url.Userinfo
		values           url.Values
	)
	env.PackageTypes["net/url"] = map[string]reflect.Type{
		"Error":            reflect.TypeOf(&err).Elem(),
		"EscapeError":      reflect.TypeOf(&escapeError).Elem(),
		"InvalidHostError": reflect.TypeOf(&invalidHostError).Elem(),
		"URL":              reflect.TypeOf(&uRL).Elem(),
		"Userinfo":         reflect.TypeOf(&userinfo).Elem(),
		"Values":           reflect.TypeOf(&values).Elem(),
	}
}

func initOS() {
	env.Packages["os"] = map[string]reflect.Value{
		// define constants
		"DevNull":           reflect.ValueOf(os.DevNull),
		"ModeAppend":        reflect.ValueOf(os.ModeAppend),
		"ModeCharDevice":    reflect.ValueOf(os.ModeCharDevice),
		"ModeDevice":        reflect.ValueOf(os.ModeDevice),
		"ModeDir":           reflect.ValueOf(os.ModeDir),
		"ModeExclusive":     reflect.ValueOf(os.ModeExclusive),
		"ModeIrregular":     reflect.ValueOf(os.ModeIrregular),
		"ModeNamedPipe":     reflect.ValueOf(os.ModeNamedPipe),
		"ModePerm":          reflect.ValueOf(os.ModePerm),
		"ModeSetgid":        reflect.ValueOf(os.ModeSetgid),
		"ModeSetuid":        reflect.ValueOf(os.ModeSetuid),
		"ModeSocket":        reflect.ValueOf(os.ModeSocket),
		"ModeSticky":        reflect.ValueOf(os.ModeSticky),
		"ModeSymlink":       reflect.ValueOf(os.ModeSymlink),
		"ModeTemporary":     reflect.ValueOf(os.ModeTemporary),
		"ModeType":          reflect.ValueOf(os.ModeType),
		"O_APPEND":          reflect.ValueOf(os.O_APPEND),
		"O_CREATE":          reflect.ValueOf(os.O_CREATE),
		"O_EXCL":            reflect.ValueOf(os.O_EXCL),
		"O_RDONLY":          reflect.ValueOf(os.O_RDONLY),
		"O_RDWR":            reflect.ValueOf(os.O_RDWR),
		"O_SYNC":            reflect.ValueOf(os.O_SYNC),
		"O_TRUNC":           reflect.ValueOf(os.O_TRUNC),
		"O_WRONLY":          reflect.ValueOf(os.O_WRONLY),
		"PathListSeparator": reflect.ValueOf(os.PathListSeparator),
		"PathSeparator":     reflect.ValueOf(os.PathSeparator),

		// define variables
		"Args":                reflect.ValueOf(os.Args),
		"ErrClosed":           reflect.ValueOf(os.ErrClosed),
		"ErrDeadlineExceeded": reflect.ValueOf(os.ErrDeadlineExceeded),
		"ErrExist":            reflect.ValueOf(os.ErrExist),
		"ErrInvalid":          reflect.ValueOf(os.ErrInvalid),
		"ErrNoDeadline":       reflect.ValueOf(os.ErrNoDeadline),
		"ErrNotExist":         reflect.ValueOf(os.ErrNotExist),
		"ErrPermission":       reflect.ValueOf(os.ErrPermission),
		"Interrupt":           reflect.ValueOf(os.Interrupt),
		"Kill":                reflect.ValueOf(os.Kill),
		"Stderr":              reflect.ValueOf(os.Stderr),
		"Stdin":               reflect.ValueOf(os.Stdin),
		"Stdout":              reflect.ValueOf(os.Stdout),

		// define functions
		"Chdir":           reflect.ValueOf(os.Chdir),
		"Chmod":           reflect.ValueOf(os.Chmod),
		"Chown":           reflect.ValueOf(os.Chown),
		"Chtimes":         reflect.ValueOf(os.Chtimes),
		"Clearenv":        reflect.ValueOf(os.Clearenv),
		"Create":          reflect.ValueOf(os.Create),
		"Environ":         reflect.ValueOf(os.Environ),
		"Executable":      reflect.ValueOf(os.Executable),
		"Exit":            reflect.ValueOf(os.Exit),
		"Expand":          reflect.ValueOf(os.Expand),
		"ExpandEnv":       reflect.ValueOf(os.ExpandEnv),
		"FindProcess":     reflect.ValueOf(os.FindProcess),
		"Getegid":         reflect.ValueOf(os.Getegid),
		"Getenv":          reflect.ValueOf(os.Getenv),
		"Geteuid":         reflect.ValueOf(os.Geteuid),
		"Getgid":          reflect.ValueOf(os.Getgid),
		"Getgroups":       reflect.ValueOf(os.Getgroups),
		"Getpagesize":     reflect.ValueOf(os.Getpagesize),
		"Getpid":          reflect.ValueOf(os.Getpid),
		"Getppid":         reflect.ValueOf(os.Getppid),
		"Getuid":          reflect.ValueOf(os.Getuid),
		"Getwd":           reflect.ValueOf(os.Getwd),
		"Hostname":        reflect.ValueOf(os.Hostname),
		"IsExist":         reflect.ValueOf(os.IsExist),
		"IsNotExist":      reflect.ValueOf(os.IsNotExist),
		"IsPathSeparator": reflect.ValueOf(os.IsPathSeparator),
		"IsPermission":    reflect.ValueOf(os.IsPermission),
		"IsTimeout":       reflect.ValueOf(os.IsTimeout),
		"Lchown":          reflect.ValueOf(os.Lchown),
		"Link":            reflect.ValueOf(os.Link),
		"LookupEnv":       reflect.ValueOf(os.LookupEnv),
		"Lstat":           reflect.ValueOf(os.Lstat),
		"Mkdir":           reflect.ValueOf(os.Mkdir),
		"MkdirAll":        reflect.ValueOf(os.MkdirAll),
		"NewFile":         reflect.ValueOf(os.NewFile),
		"NewSyscallError": reflect.ValueOf(os.NewSyscallError),
		"Open":            reflect.ValueOf(os.Open),
		"OpenFile":        reflect.ValueOf(os.OpenFile),
		"Pipe":            reflect.ValueOf(os.Pipe),
		"Readlink":        reflect.ValueOf(os.Readlink),
		"Remove":          reflect.ValueOf(os.Remove),
		"RemoveAll":       reflect.ValueOf(os.RemoveAll),
		"Rename":          reflect.ValueOf(os.Rename),
		"SameFile":        reflect.ValueOf(os.SameFile),
		"Setenv":          reflect.ValueOf(os.Setenv),
		"StartProcess":    reflect.ValueOf(os.StartProcess),
		"Stat":            reflect.ValueOf(os.Stat),
		"Symlink":         reflect.ValueOf(os.Symlink),
		"TempDir":         reflect.ValueOf(os.TempDir),
		"Truncate":        reflect.ValueOf(os.Truncate),
		"Unsetenv":        reflect.ValueOf(os.Unsetenv),
		"UserCacheDir":    reflect.ValueOf(os.UserCacheDir),
		"UserConfigDir":   reflect.ValueOf(os.UserConfigDir),
		"UserHomeDir":     reflect.ValueOf(os.UserHomeDir),
	}
	var (
		file         os.File
		fileInfo     os.FileInfo
		fileMode     os.FileMode
		linkError    os.LinkError
		pathError    os.PathError
		procAttr     os.ProcAttr
		process      os.Process
		processState os.ProcessState
		sig          os.Signal
		syscallError os.SyscallError
	)
	env.PackageTypes["os"] = map[string]reflect.Type{
		"File":         reflect.TypeOf(&file).Elem(),
		"FileInfo":     reflect.TypeOf(&fileInfo).Elem(),
		"FileMode":     reflect.TypeOf(&fileMode).Elem(),
		"LinkError":    reflect.TypeOf(&linkError).Elem(),
		"PathError":    reflect.TypeOf(&pathError).Elem(),
		"ProcAttr":     reflect.TypeOf(&procAttr).Elem(),
		"Process":      reflect.TypeOf(&process).Elem(),
		"ProcessState": reflect.TypeOf(&processState).Elem(),
		"Signal":       reflect.TypeOf(&sig).Elem(),
		"SyscallError": reflect.TypeOf(&syscallError).Elem(),
	}
}

func initOSExec() {
	env.Packages["os/exec"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrNotFound": reflect.ValueOf(exec.ErrNotFound),

		// define functions
		"Command":        reflect.ValueOf(exec.Command),
		"CommandContext": reflect.ValueOf(exec.CommandContext),
		"LookPath":       reflect.ValueOf(exec.LookPath),
	}
	var (
		cmd       exec.Cmd
		err       exec.Error
		exitError exec.ExitError
	)
	env.PackageTypes["os/exec"] = map[string]reflect.Type{
		"Cmd":       reflect.TypeOf(&cmd).Elem(),
		"Error":     reflect.TypeOf(&err).Elem(),
		"ExitError": reflect.TypeOf(&exitError).Elem(),
	}
}

func initOSSignal() {
	env.Packages["os/signal"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Ignore":  reflect.ValueOf(signal.Ignore),
		"Ignored": reflect.ValueOf(signal.Ignored),
		"Notify":  reflect.ValueOf(signal.Notify),
		"Reset":   reflect.ValueOf(signal.Reset),
		"Stop":    reflect.ValueOf(signal.Stop),
	}
	var ()
	env.PackageTypes["os/signal"] = map[string]reflect.Type{}
}

func initOSUser() {
	env.Packages["os/user"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Current":       reflect.ValueOf(user.Current),
		"Lookup":        reflect.ValueOf(user.Lookup),
		"LookupGroup":   reflect.ValueOf(user.LookupGroup),
		"LookupGroupId": reflect.ValueOf(user.LookupGroupId),
		"LookupId":      reflect.ValueOf(user.LookupId),
	}
	var (
		group               user.Group
		unknownGroupError   user.UnknownGroupError
		unknownGroupIDError user.UnknownGroupIdError
		unknownUserError    user.UnknownUserError
		unknownUserIDError  user.UnknownUserIdError
		usr                 user.User
	)
	env.PackageTypes["os/user"] = map[string]reflect.Type{
		"Group":               reflect.TypeOf(&group).Elem(),
		"UnknownGroupError":   reflect.TypeOf(&unknownGroupError).Elem(),
		"UnknownGroupIdError": reflect.TypeOf(&unknownGroupIDError).Elem(),
		"UnknownUserError":    reflect.TypeOf(&unknownUserError).Elem(),
		"UnknownUserIdError":  reflect.TypeOf(&unknownUserIDError).Elem(),
		"User":                reflect.TypeOf(&usr).Elem(),
	}
}

func initPath() {
	env.Packages["path"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrBadPattern": reflect.ValueOf(path.ErrBadPattern),

		// define functions
		"Base":  reflect.ValueOf(path.Base),
		"Clean": reflect.ValueOf(path.Clean),
		"Dir":   reflect.ValueOf(path.Dir),
		"Ext":   reflect.ValueOf(path.Ext),
		"IsAbs": reflect.ValueOf(path.IsAbs),
		"Join":  reflect.ValueOf(path.Join),
		"Match": reflect.ValueOf(path.Match),
		"Split": reflect.ValueOf(path.Split),
	}
	var ()
	env.PackageTypes["path"] = map[string]reflect.Type{}
}

func initPathFilepath() {
	env.Packages["path/filepath"] = map[string]reflect.Value{
		// define constants
		"ListSeparator": reflect.ValueOf(filepath.ListSeparator),
		"Separator":     reflect.ValueOf(filepath.Separator),

		// define variables
		"ErrBadPattern": reflect.ValueOf(filepath.ErrBadPattern),
		"SkipDir":       reflect.ValueOf(filepath.SkipDir),

		// define functions
		"Abs":          reflect.ValueOf(filepath.Abs),
		"Base":         reflect.ValueOf(filepath.Base),
		"Clean":        reflect.ValueOf(filepath.Clean),
		"Dir":          reflect.ValueOf(filepath.Dir),
		"EvalSymlinks": reflect.ValueOf(filepath.EvalSymlinks),
		"Ext":          reflect.ValueOf(filepath.Ext),
		"FromSlash":    reflect.ValueOf(filepath.FromSlash),
		"Glob":         reflect.ValueOf(filepath.Glob),
		"IsAbs":        reflect.ValueOf(filepath.IsAbs),
		"Join":         reflect.ValueOf(filepath.Join),
		"Match":        reflect.ValueOf(filepath.Match),
		"Rel":          reflect.ValueOf(filepath.Rel),
		"Split":        reflect.ValueOf(filepath.Split),
		"SplitList":    reflect.ValueOf(filepath.SplitList),
		"ToSlash":      reflect.ValueOf(filepath.ToSlash),
		"VolumeName":   reflect.ValueOf(filepath.VolumeName),
		"Walk":         reflect.ValueOf(filepath.Walk),
	}
	var (
		walkFunc filepath.WalkFunc
	)
	env.PackageTypes["path/filepath"] = map[string]reflect.Type{
		"WalkFunc": reflect.TypeOf(&walkFunc).Elem(),
	}
}

func initReflect() {
	env.Packages["reflect"] = map[string]reflect.Value{
		// define constants
		"Array":         reflect.ValueOf(reflect.Array),
		"Bool":          reflect.ValueOf(reflect.Bool),
		"BothDir":       reflect.ValueOf(reflect.BothDir),
		"Chan":          reflect.ValueOf(reflect.Chan),
		"Complex128":    reflect.ValueOf(reflect.Complex128),
		"Complex64":     reflect.ValueOf(reflect.Complex64),
		"Float32":       reflect.ValueOf(reflect.Float32),
		"Float64":       reflect.ValueOf(reflect.Float64),
		"Func":          reflect.ValueOf(reflect.Func),
		"Int":           reflect.ValueOf(reflect.Int),
		"Int16":         reflect.ValueOf(reflect.Int16),
		"Int32":         reflect.ValueOf(reflect.Int32),
		"Int64":         reflect.ValueOf(reflect.Int64),
		"Int8":          reflect.ValueOf(reflect.Int8),
		"Interface":     reflect.ValueOf(reflect.Interface),
		"Invalid":       reflect.ValueOf(reflect.Invalid),
		"Map":           reflect.ValueOf(reflect.Map),
		"Ptr":           reflect.ValueOf(reflect.Ptr),
		"RecvDir":       reflect.ValueOf(reflect.RecvDir),
		"SelectDefault": reflect.ValueOf(reflect.SelectDefault),
		"SelectRecv":    reflect.ValueOf(reflect.SelectRecv),
		"SelectSend":    reflect.ValueOf(reflect.SelectSend),
		"SendDir":       reflect.ValueOf(reflect.SendDir),
		"Slice":         reflect.ValueOf(reflect.Slice),
		"String":        reflect.ValueOf(reflect.String),
		"Struct":        reflect.ValueOf(reflect.Struct),
		"Uint":          reflect.ValueOf(reflect.Uint),
		"Uint16":        reflect.ValueOf(reflect.Uint16),
		"Uint32":        reflect.ValueOf(reflect.Uint32),
		"Uint64":        reflect.ValueOf(reflect.Uint64),
		"Uint8":         reflect.ValueOf(reflect.Uint8),
		"Uintptr":       reflect.ValueOf(reflect.Uintptr),
		"UnsafePointer": reflect.ValueOf(reflect.UnsafePointer),

		// define variables

		// define functions
		"Append":          reflect.ValueOf(reflect.Append),
		"AppendSlice":     reflect.ValueOf(reflect.AppendSlice),
		"ArrayOf":         reflect.ValueOf(reflect.ArrayOf),
		"ChanOf":          reflect.ValueOf(reflect.ChanOf),
		"Copy":            reflect.ValueOf(reflect.Copy),
		"DeepEqual":       reflect.ValueOf(reflect.DeepEqual),
		"FuncOf":          reflect.ValueOf(reflect.FuncOf),
		"Indirect":        reflect.ValueOf(reflect.Indirect),
		"MakeChan":        reflect.ValueOf(reflect.MakeChan),
		"MakeFunc":        reflect.ValueOf(reflect.MakeFunc),
		"MakeMap":         reflect.ValueOf(reflect.MakeMap),
		"MakeMapWithSize": reflect.ValueOf(reflect.MakeMapWithSize),
		"MakeSlice":       reflect.ValueOf(reflect.MakeSlice),
		"MapOf":           reflect.ValueOf(reflect.MapOf),
		"New":             reflect.ValueOf(reflect.New),
		"NewAt":           reflect.ValueOf(reflect.NewAt),
		"PtrTo":           reflect.ValueOf(reflect.PtrTo),
		"Select":          reflect.ValueOf(reflect.Select),
		"SliceOf":         reflect.ValueOf(reflect.SliceOf),
		"StructOf":        reflect.ValueOf(reflect.StructOf),
		"Swapper":         reflect.ValueOf(reflect.Swapper),
		"TypeOf":          reflect.ValueOf(reflect.TypeOf),
		"ValueOf":         reflect.ValueOf(reflect.ValueOf),
		"Zero":            reflect.ValueOf(reflect.Zero),
	}
	var (
		chanDir      reflect.ChanDir
		kind         reflect.Kind
		mapIter      reflect.MapIter
		method       reflect.Method
		selectCase   reflect.SelectCase
		selectDir    reflect.SelectDir
		sliceHeader  reflect.SliceHeader
		stringHeader reflect.StringHeader
		structField  reflect.StructField
		structTag    reflect.StructTag
		typ          reflect.Type
		value        reflect.Value
		valueError   reflect.ValueError
	)
	env.PackageTypes["reflect"] = map[string]reflect.Type{
		"ChanDir":      reflect.TypeOf(&chanDir).Elem(),
		"Kind":         reflect.TypeOf(&kind).Elem(),
		"MapIter":      reflect.TypeOf(&mapIter).Elem(),
		"Method":       reflect.TypeOf(&method).Elem(),
		"SelectCase":   reflect.TypeOf(&selectCase).Elem(),
		"SelectDir":    reflect.TypeOf(&selectDir).Elem(),
		"SliceHeader":  reflect.TypeOf(&sliceHeader).Elem(),
		"StringHeader": reflect.TypeOf(&stringHeader).Elem(),
		"StructField":  reflect.TypeOf(&structField).Elem(),
		"StructTag":    reflect.TypeOf(&structTag).Elem(),
		"Type":         reflect.TypeOf(&typ).Elem(),
		"Value":        reflect.TypeOf(&value).Elem(),
		"ValueError":   reflect.TypeOf(&valueError).Elem(),
	}
}

func initRegexp() {
	env.Packages["regexp"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Compile":          reflect.ValueOf(regexp.Compile),
		"CompilePOSIX":     reflect.ValueOf(regexp.CompilePOSIX),
		"Match":            reflect.ValueOf(regexp.Match),
		"MatchReader":      reflect.ValueOf(regexp.MatchReader),
		"MatchString":      reflect.ValueOf(regexp.MatchString),
		"MustCompile":      reflect.ValueOf(regexp.MustCompile),
		"MustCompilePOSIX": reflect.ValueOf(regexp.MustCompilePOSIX),
		"QuoteMeta":        reflect.ValueOf(regexp.QuoteMeta),
	}
	var (
		reg regexp.Regexp
	)
	env.PackageTypes["regexp"] = map[string]reflect.Type{
		"Regexp": reflect.TypeOf(&reg).Elem(),
	}
}

func initSort() {
	env.Packages["sort"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Float64s":          reflect.ValueOf(sort.Float64s),
		"Float64sAreSorted": reflect.ValueOf(sort.Float64sAreSorted),
		"Ints":              reflect.ValueOf(sort.Ints),
		"IntsAreSorted":     reflect.ValueOf(sort.IntsAreSorted),
		"IsSorted":          reflect.ValueOf(sort.IsSorted),
		"Reverse":           reflect.ValueOf(sort.Reverse),
		"Search":            reflect.ValueOf(sort.Search),
		"SearchFloat64s":    reflect.ValueOf(sort.SearchFloat64s),
		"SearchInts":        reflect.ValueOf(sort.SearchInts),
		"SearchStrings":     reflect.ValueOf(sort.SearchStrings),
		"Slice":             reflect.ValueOf(sort.Slice),
		"SliceIsSorted":     reflect.ValueOf(sort.SliceIsSorted),
		"SliceStable":       reflect.ValueOf(sort.SliceStable),
		"Sort":              reflect.ValueOf(sort.Sort),
		"Stable":            reflect.ValueOf(sort.Stable),
		"Strings":           reflect.ValueOf(sort.Strings),
		"StringsAreSorted":  reflect.ValueOf(sort.StringsAreSorted),
	}
	var (
		float64Slice sort.Float64Slice
		intSlice     sort.IntSlice
		iface        sort.Interface
		stringSlice  sort.StringSlice
	)
	env.PackageTypes["sort"] = map[string]reflect.Type{
		"Float64Slice": reflect.TypeOf(&float64Slice).Elem(),
		"IntSlice":     reflect.TypeOf(&intSlice).Elem(),
		"Interface":    reflect.TypeOf(&iface).Elem(),
		"StringSlice":  reflect.TypeOf(&stringSlice).Elem(),
	}
}

func initStrconv() {
	env.Packages["strconv"] = map[string]reflect.Value{
		// define constants
		"IntSize": reflect.ValueOf(strconv.IntSize),

		// define variables
		"ErrRange":  reflect.ValueOf(strconv.ErrRange),
		"ErrSyntax": reflect.ValueOf(strconv.ErrSyntax),

		// define functions
		"AppendBool":               reflect.ValueOf(strconv.AppendBool),
		"AppendFloat":              reflect.ValueOf(strconv.AppendFloat),
		"AppendInt":                reflect.ValueOf(strconv.AppendInt),
		"AppendQuote":              reflect.ValueOf(strconv.AppendQuote),
		"AppendQuoteRune":          reflect.ValueOf(strconv.AppendQuoteRune),
		"AppendQuoteRuneToASCII":   reflect.ValueOf(strconv.AppendQuoteRuneToASCII),
		"AppendQuoteRuneToGraphic": reflect.ValueOf(strconv.AppendQuoteRuneToGraphic),
		"AppendQuoteToASCII":       reflect.ValueOf(strconv.AppendQuoteToASCII),
		"AppendQuoteToGraphic":     reflect.ValueOf(strconv.AppendQuoteToGraphic),
		"AppendUint":               reflect.ValueOf(strconv.AppendUint),
		"Atoi":                     reflect.ValueOf(strconv.Atoi),
		"CanBackquote":             reflect.ValueOf(strconv.CanBackquote),
		"FormatBool":               reflect.ValueOf(strconv.FormatBool),
		"FormatComplex":            reflect.ValueOf(strconv.FormatComplex),
		"FormatFloat":              reflect.ValueOf(strconv.FormatFloat),
		"FormatInt":                reflect.ValueOf(strconv.FormatInt),
		"FormatUint":               reflect.ValueOf(strconv.FormatUint),
		"IsGraphic":                reflect.ValueOf(strconv.IsGraphic),
		"IsPrint":                  reflect.ValueOf(strconv.IsPrint),
		"Itoa":                     reflect.ValueOf(strconv.Itoa),
		"ParseBool":                reflect.ValueOf(strconv.ParseBool),
		"ParseComplex":             reflect.ValueOf(strconv.ParseComplex),
		"ParseFloat":               reflect.ValueOf(strconv.ParseFloat),
		"ParseInt":                 reflect.ValueOf(strconv.ParseInt),
		"ParseUint":                reflect.ValueOf(strconv.ParseUint),
		"Quote":                    reflect.ValueOf(strconv.Quote),
		"QuoteRune":                reflect.ValueOf(strconv.QuoteRune),
		"QuoteRuneToASCII":         reflect.ValueOf(strconv.QuoteRuneToASCII),
		"QuoteRuneToGraphic":       reflect.ValueOf(strconv.QuoteRuneToGraphic),
		"QuoteToASCII":             reflect.ValueOf(strconv.QuoteToASCII),
		"QuoteToGraphic":           reflect.ValueOf(strconv.QuoteToGraphic),
		"Unquote":                  reflect.ValueOf(strconv.Unquote),
		"UnquoteChar":              reflect.ValueOf(strconv.UnquoteChar),
	}
	var (
		numError strconv.NumError
	)
	env.PackageTypes["strconv"] = map[string]reflect.Type{
		"NumError": reflect.TypeOf(&numError).Elem(),
	}
}

func initStrings() {
	env.Packages["strings"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Compare":        reflect.ValueOf(strings.Compare),
		"Contains":       reflect.ValueOf(strings.Contains),
		"ContainsAny":    reflect.ValueOf(strings.ContainsAny),
		"ContainsRune":   reflect.ValueOf(strings.ContainsRune),
		"Count":          reflect.ValueOf(strings.Count),
		"EqualFold":      reflect.ValueOf(strings.EqualFold),
		"Fields":         reflect.ValueOf(strings.Fields),
		"FieldsFunc":     reflect.ValueOf(strings.FieldsFunc),
		"HasPrefix":      reflect.ValueOf(strings.HasPrefix),
		"HasSuffix":      reflect.ValueOf(strings.HasSuffix),
		"Index":          reflect.ValueOf(strings.Index),
		"IndexAny":       reflect.ValueOf(strings.IndexAny),
		"IndexByte":      reflect.ValueOf(strings.IndexByte),
		"IndexFunc":      reflect.ValueOf(strings.IndexFunc),
		"IndexRune":      reflect.ValueOf(strings.IndexRune),
		"Join":           reflect.ValueOf(strings.Join),
		"LastIndex":      reflect.ValueOf(strings.LastIndex),
		"LastIndexAny":   reflect.ValueOf(strings.LastIndexAny),
		"LastIndexByte":  reflect.ValueOf(strings.LastIndexByte),
		"LastIndexFunc":  reflect.ValueOf(strings.LastIndexFunc),
		"Map":            reflect.ValueOf(strings.Map),
		"NewReader":      reflect.ValueOf(strings.NewReader),
		"NewReplacer":    reflect.ValueOf(strings.NewReplacer),
		"Repeat":         reflect.ValueOf(strings.Repeat),
		"Replace":        reflect.ValueOf(strings.Replace),
		"ReplaceAll":     reflect.ValueOf(strings.ReplaceAll),
		"Split":          reflect.ValueOf(strings.Split),
		"SplitAfter":     reflect.ValueOf(strings.SplitAfter),
		"SplitAfterN":    reflect.ValueOf(strings.SplitAfterN),
		"SplitN":         reflect.ValueOf(strings.SplitN),
		"Title":          reflect.ValueOf(strings.Title),
		"ToLower":        reflect.ValueOf(strings.ToLower),
		"ToLowerSpecial": reflect.ValueOf(strings.ToLowerSpecial),
		"ToTitle":        reflect.ValueOf(strings.ToTitle),
		"ToTitleSpecial": reflect.ValueOf(strings.ToTitleSpecial),
		"ToUpper":        reflect.ValueOf(strings.ToUpper),
		"ToUpperSpecial": reflect.ValueOf(strings.ToUpperSpecial),
		"ToValidUTF8":    reflect.ValueOf(strings.ToValidUTF8),
		"Trim":           reflect.ValueOf(strings.Trim),
		"TrimFunc":       reflect.ValueOf(strings.TrimFunc),
		"TrimLeft":       reflect.ValueOf(strings.TrimLeft),
		"TrimLeftFunc":   reflect.ValueOf(strings.TrimLeftFunc),
		"TrimPrefix":     reflect.ValueOf(strings.TrimPrefix),
		"TrimRight":      reflect.ValueOf(strings.TrimRight),
		"TrimRightFunc":  reflect.ValueOf(strings.TrimRightFunc),
		"TrimSpace":      reflect.ValueOf(strings.TrimSpace),
		"TrimSuffix":     reflect.ValueOf(strings.TrimSuffix),
	}
	var (
		builder  strings.Builder
		reader   strings.Reader
		replacer strings.Replacer
	)
	env.PackageTypes["strings"] = map[string]reflect.Type{
		"Builder":  reflect.TypeOf(&builder).Elem(),
		"Reader":   reflect.TypeOf(&reader).Elem(),
		"Replacer": reflect.TypeOf(&replacer).Elem(),
	}
}

func initSync() {
	env.Packages["sync"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewCond": reflect.ValueOf(sync.NewCond),
	}
	var (
		cond      sync.Cond
		locker    sync.Locker
		m         sync.Map
		mutex     sync.Mutex
		once      sync.Once
		pool      sync.Pool
		rWMutex   sync.RWMutex
		waitGroup sync.WaitGroup
	)
	env.PackageTypes["sync"] = map[string]reflect.Type{
		"Cond":      reflect.TypeOf(&cond).Elem(),
		"Locker":    reflect.TypeOf(&locker).Elem(),
		"Map":       reflect.TypeOf(&m).Elem(),
		"Mutex":     reflect.TypeOf(&mutex).Elem(),
		"Once":      reflect.TypeOf(&once).Elem(),
		"Pool":      reflect.TypeOf(&pool).Elem(),
		"RWMutex":   reflect.TypeOf(&rWMutex).Elem(),
		"WaitGroup": reflect.TypeOf(&waitGroup).Elem(),
	}
}

func initSyncAtomic() {
	env.Packages["sync/atomic"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"AddInt32":              reflect.ValueOf(atomic.AddInt32),
		"AddInt64":              reflect.ValueOf(atomic.AddInt64),
		"AddUint32":             reflect.ValueOf(atomic.AddUint32),
		"AddUint64":             reflect.ValueOf(atomic.AddUint64),
		"AddUintptr":            reflect.ValueOf(atomic.AddUintptr),
		"CompareAndSwapInt32":   reflect.ValueOf(atomic.CompareAndSwapInt32),
		"CompareAndSwapInt64":   reflect.ValueOf(atomic.CompareAndSwapInt64),
		"CompareAndSwapPointer": reflect.ValueOf(atomic.CompareAndSwapPointer),
		"CompareAndSwapUint32":  reflect.ValueOf(atomic.CompareAndSwapUint32),
		"CompareAndSwapUint64":  reflect.ValueOf(atomic.CompareAndSwapUint64),
		"CompareAndSwapUintptr": reflect.ValueOf(atomic.CompareAndSwapUintptr),
		"LoadInt32":             reflect.ValueOf(atomic.LoadInt32),
		"LoadInt64":             reflect.ValueOf(atomic.LoadInt64),
		"LoadPointer":           reflect.ValueOf(atomic.LoadPointer),
		"LoadUint32":            reflect.ValueOf(atomic.LoadUint32),
		"LoadUint64":            reflect.ValueOf(atomic.LoadUint64),
		"LoadUintptr":           reflect.ValueOf(atomic.LoadUintptr),
		"StoreInt32":            reflect.ValueOf(atomic.StoreInt32),
		"StoreInt64":            reflect.ValueOf(atomic.StoreInt64),
		"StorePointer":          reflect.ValueOf(atomic.StorePointer),
		"StoreUint32":           reflect.ValueOf(atomic.StoreUint32),
		"StoreUint64":           reflect.ValueOf(atomic.StoreUint64),
		"StoreUintptr":          reflect.ValueOf(atomic.StoreUintptr),
		"SwapInt32":             reflect.ValueOf(atomic.SwapInt32),
		"SwapInt64":             reflect.ValueOf(atomic.SwapInt64),
		"SwapPointer":           reflect.ValueOf(atomic.SwapPointer),
		"SwapUint32":            reflect.ValueOf(atomic.SwapUint32),
		"SwapUint64":            reflect.ValueOf(atomic.SwapUint64),
		"SwapUintptr":           reflect.ValueOf(atomic.SwapUintptr),
	}
	var (
		value atomic.Value
	)
	env.PackageTypes["sync/atomic"] = map[string]reflect.Type{
		"Value": reflect.TypeOf(&value).Elem(),
	}
}

func initTime() {
	env.Packages["time"] = map[string]reflect.Value{
		// define constants
		"ANSIC":       reflect.ValueOf(time.ANSIC),
		"April":       reflect.ValueOf(time.April),
		"August":      reflect.ValueOf(time.August),
		"December":    reflect.ValueOf(time.December),
		"February":    reflect.ValueOf(time.February),
		"Friday":      reflect.ValueOf(time.Friday),
		"Hour":        reflect.ValueOf(time.Hour),
		"January":     reflect.ValueOf(time.January),
		"July":        reflect.ValueOf(time.July),
		"June":        reflect.ValueOf(time.June),
		"Kitchen":     reflect.ValueOf(time.Kitchen),
		"March":       reflect.ValueOf(time.March),
		"May":         reflect.ValueOf(time.May),
		"Microsecond": reflect.ValueOf(time.Microsecond),
		"Millisecond": reflect.ValueOf(time.Millisecond),
		"Minute":      reflect.ValueOf(time.Minute),
		"Monday":      reflect.ValueOf(time.Monday),
		"Nanosecond":  reflect.ValueOf(time.Nanosecond),
		"November":    reflect.ValueOf(time.November),
		"October":     reflect.ValueOf(time.October),
		"RFC1123":     reflect.ValueOf(time.RFC1123),
		"RFC1123Z":    reflect.ValueOf(time.RFC1123Z),
		"RFC3339":     reflect.ValueOf(time.RFC3339),
		"RFC3339Nano": reflect.ValueOf(time.RFC3339Nano),
		"RFC822":      reflect.ValueOf(time.RFC822),
		"RFC822Z":     reflect.ValueOf(time.RFC822Z),
		"RFC850":      reflect.ValueOf(time.RFC850),
		"RubyDate":    reflect.ValueOf(time.RubyDate),
		"Saturday":    reflect.ValueOf(time.Saturday),
		"Second":      reflect.ValueOf(time.Second),
		"September":   reflect.ValueOf(time.September),
		"Stamp":       reflect.ValueOf(time.Stamp),
		"StampMicro":  reflect.ValueOf(time.StampMicro),
		"StampMilli":  reflect.ValueOf(time.StampMilli),
		"StampNano":   reflect.ValueOf(time.StampNano),
		"Sunday":      reflect.ValueOf(time.Sunday),
		"Thursday":    reflect.ValueOf(time.Thursday),
		"Tuesday":     reflect.ValueOf(time.Tuesday),
		"UnixDate":    reflect.ValueOf(time.UnixDate),
		"Wednesday":   reflect.ValueOf(time.Wednesday),

		// define variables
		"Local": reflect.ValueOf(time.Local),
		"UTC":   reflect.ValueOf(time.UTC),

		// define functions
		"After":                  reflect.ValueOf(time.After),
		"AfterFunc":              reflect.ValueOf(time.AfterFunc),
		"Date":                   reflect.ValueOf(time.Date),
		"FixedZone":              reflect.ValueOf(time.FixedZone),
		"LoadLocation":           reflect.ValueOf(time.LoadLocation),
		"LoadLocationFromTZData": reflect.ValueOf(time.LoadLocationFromTZData),
		"NewTicker":              reflect.ValueOf(time.NewTicker),
		"NewTimer":               reflect.ValueOf(time.NewTimer),
		"Now":                    reflect.ValueOf(time.Now),
		"Parse":                  reflect.ValueOf(time.Parse),
		"ParseDuration":          reflect.ValueOf(time.ParseDuration),
		"ParseInLocation":        reflect.ValueOf(time.ParseInLocation),
		"Since":                  reflect.ValueOf(time.Since),
		"Sleep":                  reflect.ValueOf(time.Sleep),
		"Tick":                   reflect.ValueOf(time.Tick),
		"Unix":                   reflect.ValueOf(time.Unix),
		"Until":                  reflect.ValueOf(time.Until),
	}
	var (
		duration   time.Duration
		location   time.Location
		month      time.Month
		parseError time.ParseError
		ticker     time.Ticker
		t          time.Time
		timer      time.Timer
		weekday    time.Weekday
	)
	env.PackageTypes["time"] = map[string]reflect.Type{
		"Duration":   reflect.TypeOf(&duration).Elem(),
		"Location":   reflect.TypeOf(&location).Elem(),
		"Month":      reflect.TypeOf(&month).Elem(),
		"ParseError": reflect.TypeOf(&parseError).Elem(),
		"Ticker":     reflect.TypeOf(&ticker).Elem(),
		"Time":       reflect.TypeOf(&t).Elem(),
		"Timer":      reflect.TypeOf(&timer).Elem(),
		"Weekday":    reflect.TypeOf(&weekday).Elem(),
	}
}

func initUnicode() {
	env.Packages["unicode"] = map[string]reflect.Value{
		// define constants
		"LowerCase":       reflect.ValueOf(unicode.LowerCase),
		"MaxASCII":        reflect.ValueOf(unicode.MaxASCII),
		"MaxCase":         reflect.ValueOf(unicode.MaxCase),
		"MaxLatin1":       reflect.ValueOf(unicode.MaxLatin1),
		"MaxRune":         reflect.ValueOf(unicode.MaxRune),
		"ReplacementChar": reflect.ValueOf(unicode.ReplacementChar),
		"TitleCase":       reflect.ValueOf(unicode.TitleCase),
		"UpperCase":       reflect.ValueOf(unicode.UpperCase),
		"UpperLower":      reflect.ValueOf(unicode.UpperLower),
		"Version":         reflect.ValueOf(unicode.Version),

		// define variables
		"ASCII_Hex_Digit":                    reflect.ValueOf(unicode.ASCII_Hex_Digit),
		"Adlam":                              reflect.ValueOf(unicode.Adlam),
		"Ahom":                               reflect.ValueOf(unicode.Ahom),
		"Anatolian_Hieroglyphs":              reflect.ValueOf(unicode.Anatolian_Hieroglyphs),
		"Arabic":                             reflect.ValueOf(unicode.Arabic),
		"Armenian":                           reflect.ValueOf(unicode.Armenian),
		"Avestan":                            reflect.ValueOf(unicode.Avestan),
		"AzeriCase":                          reflect.ValueOf(unicode.AzeriCase),
		"Balinese":                           reflect.ValueOf(unicode.Balinese),
		"Bamum":                              reflect.ValueOf(unicode.Bamum),
		"Bassa_Vah":                          reflect.ValueOf(unicode.Bassa_Vah),
		"Batak":                              reflect.ValueOf(unicode.Batak),
		"Bengali":                            reflect.ValueOf(unicode.Bengali),
		"Bhaiksuki":                          reflect.ValueOf(unicode.Bhaiksuki),
		"Bidi_Control":                       reflect.ValueOf(unicode.Bidi_Control),
		"Bopomofo":                           reflect.ValueOf(unicode.Bopomofo),
		"Brahmi":                             reflect.ValueOf(unicode.Brahmi),
		"Braille":                            reflect.ValueOf(unicode.Braille),
		"Buginese":                           reflect.ValueOf(unicode.Buginese),
		"Buhid":                              reflect.ValueOf(unicode.Buhid),
		"C":                                  reflect.ValueOf(unicode.C),
		"Canadian_Aboriginal":                reflect.ValueOf(unicode.Canadian_Aboriginal),
		"Carian":                             reflect.ValueOf(unicode.Carian),
		"CaseRanges":                         reflect.ValueOf(unicode.CaseRanges),
		"Categories":                         reflect.ValueOf(unicode.Categories),
		"Caucasian_Albanian":                 reflect.ValueOf(unicode.Caucasian_Albanian),
		"Cc":                                 reflect.ValueOf(unicode.Cc),
		"Cf":                                 reflect.ValueOf(unicode.Cf),
		"Chakma":                             reflect.ValueOf(unicode.Chakma),
		"Cham":                               reflect.ValueOf(unicode.Cham),
		"Cherokee":                           reflect.ValueOf(unicode.Cherokee),
		"Co":                                 reflect.ValueOf(unicode.Co),
		"Common":                             reflect.ValueOf(unicode.Common),
		"Coptic":                             reflect.ValueOf(unicode.Coptic),
		"Cs":                                 reflect.ValueOf(unicode.Cs),
		"Cuneiform":                          reflect.ValueOf(unicode.Cuneiform),
		"Cypriot":                            reflect.ValueOf(unicode.Cypriot),
		"Cyrillic":                           reflect.ValueOf(unicode.Cyrillic),
		"Dash":                               reflect.ValueOf(unicode.Dash),
		"Deprecated":                         reflect.ValueOf(unicode.Deprecated),
		"Deseret":                            reflect.ValueOf(unicode.Deseret),
		"Devanagari":                         reflect.ValueOf(unicode.Devanagari),
		"Diacritic":                          reflect.ValueOf(unicode.Diacritic),
		"Digit":                              reflect.ValueOf(unicode.Digit),
		"Dogra":                              reflect.ValueOf(unicode.Dogra),
		"Duployan":                           reflect.ValueOf(unicode.Duployan),
		"Egyptian_Hieroglyphs":               reflect.ValueOf(unicode.Egyptian_Hieroglyphs),
		"Elbasan":                            reflect.ValueOf(unicode.Elbasan),
		"Elymaic":                            reflect.ValueOf(unicode.Elymaic),
		"Ethiopic":                           reflect.ValueOf(unicode.Ethiopic),
		"Extender":                           reflect.ValueOf(unicode.Extender),
		"FoldCategory":                       reflect.ValueOf(unicode.FoldCategory),
		"FoldScript":                         reflect.ValueOf(unicode.FoldScript),
		"Georgian":                           reflect.ValueOf(unicode.Georgian),
		"Glagolitic":                         reflect.ValueOf(unicode.Glagolitic),
		"Gothic":                             reflect.ValueOf(unicode.Gothic),
		"Grantha":                            reflect.ValueOf(unicode.Grantha),
		"GraphicRanges":                      reflect.ValueOf(unicode.GraphicRanges),
		"Greek":                              reflect.ValueOf(unicode.Greek),
		"Gujarati":                           reflect.ValueOf(unicode.Gujarati),
		"Gunjala_Gondi":                      reflect.ValueOf(unicode.Gunjala_Gondi),
		"Gurmukhi":                           reflect.ValueOf(unicode.Gurmukhi),
		"Han":                                reflect.ValueOf(unicode.Han),
		"Hangul":                             reflect.ValueOf(unicode.Hangul),
		"Hanifi_Rohingya":                    reflect.ValueOf(unicode.Hanifi_Rohingya),
		"Hanunoo":                            reflect.ValueOf(unicode.Hanunoo),
		"Hatran":                             reflect.ValueOf(unicode.Hatran),
		"Hebrew":                             reflect.ValueOf(unicode.Hebrew),
		"Hex_Digit":                          reflect.ValueOf(unicode.Hex_Digit),
		"Hiragana":                           reflect.ValueOf(unicode.Hiragana),
		"Hyphen":                             reflect.ValueOf(unicode.Hyphen),
		"IDS_Binary_Operator":                reflect.ValueOf(unicode.IDS_Binary_Operator),
		"IDS_Trinary_Operator":               reflect.ValueOf(unicode.IDS_Trinary_Operator),
		"Ideographic":                        reflect.ValueOf(unicode.Ideographic),
		"Imperial_Aramaic":                   reflect.ValueOf(unicode.Imperial_Aramaic),
		"Inherited":                          reflect.ValueOf(unicode.Inherited),
		"Inscriptional_Pahlavi":              reflect.ValueOf(unicode.Inscriptional_Pahlavi),
		"Inscriptional_Parthian":             reflect.ValueOf(unicode.Inscriptional_Parthian),
		"Javanese":                           reflect.ValueOf(unicode.Javanese),
		"Join_Control":                       reflect.ValueOf(unicode.Join_Control),
		"Kaithi":                             reflect.ValueOf(unicode.Kaithi),
		"Kannada":                            reflect.ValueOf(unicode.Kannada),
		"Katakana":                           reflect.ValueOf(unicode.Katakana),
		"Kayah_Li":                           reflect.ValueOf(unicode.Kayah_Li),
		"Kharoshthi":                         reflect.ValueOf(unicode.Kharoshthi),
		"Khmer":                              reflect.ValueOf(unicode.Khmer),
		"Khojki":                             reflect.ValueOf(unicode.Khojki),
		"Khudawadi":                          reflect.ValueOf(unicode.Khudawadi),
		"L":                                  reflect.ValueOf(unicode.L),
		"Lao":                                reflect.ValueOf(unicode.Lao),
		"Latin":                              reflect.ValueOf(unicode.Latin),
		"Lepcha":                             reflect.ValueOf(unicode.Lepcha),
		"Letter":                             reflect.ValueOf(unicode.Letter),
		"Limbu":                              reflect.ValueOf(unicode.Limbu),
		"Linear_A":                           reflect.ValueOf(unicode.Linear_A),
		"Linear_B":                           reflect.ValueOf(unicode.Linear_B),
		"Lisu":                               reflect.ValueOf(unicode.Lisu),
		"Ll":                                 reflect.ValueOf(unicode.Ll),
		"Lm":                                 reflect.ValueOf(unicode.Lm),
		"Lo":                                 reflect.ValueOf(unicode.Lo),
		"Logical_Order_Exception":            reflect.ValueOf(unicode.Logical_Order_Exception),
		"Lower":                              reflect.ValueOf(unicode.Lower),
		"Lt":                                 reflect.ValueOf(unicode.Lt),
		"Lu":                                 reflect.ValueOf(unicode.Lu),
		"Lycian":                             reflect.ValueOf(unicode.Lycian),
		"Lydian":                             reflect.ValueOf(unicode.Lydian),
		"M":                                  reflect.ValueOf(unicode.M),
		"Mahajani":                           reflect.ValueOf(unicode.Mahajani),
		"Makasar":                            reflect.ValueOf(unicode.Makasar),
		"Malayalam":                          reflect.ValueOf(unicode.Malayalam),
		"Mandaic":                            reflect.ValueOf(unicode.Mandaic),
		"Manichaean":                         reflect.ValueOf(unicode.Manichaean),
		"Marchen":                            reflect.ValueOf(unicode.Marchen),
		"Mark":                               reflect.ValueOf(unicode.Mark),
		"Masaram_Gondi":                      reflect.ValueOf(unicode.Masaram_Gondi),
		"Mc":                                 reflect.ValueOf(unicode.Mc),
		"Me":                                 reflect.ValueOf(unicode.Me),
		"Medefaidrin":                        reflect.ValueOf(unicode.Medefaidrin),
		"Meetei_Mayek":                       reflect.ValueOf(unicode.Meetei_Mayek),
		"Mende_Kikakui":                      reflect.ValueOf(unicode.Mende_Kikakui),
		"Meroitic_Cursive":                   reflect.ValueOf(unicode.Meroitic_Cursive),
		"Meroitic_Hieroglyphs":               reflect.ValueOf(unicode.Meroitic_Hieroglyphs),
		"Miao":                               reflect.ValueOf(unicode.Miao),
		"Mn":                                 reflect.ValueOf(unicode.Mn),
		"Modi":                               reflect.ValueOf(unicode.Modi),
		"Mongolian":                          reflect.ValueOf(unicode.Mongolian),
		"Mro":                                reflect.ValueOf(unicode.Mro),
		"Multani":                            reflect.ValueOf(unicode.Multani),
		"Myanmar":                            reflect.ValueOf(unicode.Myanmar),
		"N":                                  reflect.ValueOf(unicode.N),
		"Nabataean":                          reflect.ValueOf(unicode.Nabataean),
		"Nandinagari":                        reflect.ValueOf(unicode.Nandinagari),
		"Nd":                                 reflect.ValueOf(unicode.Nd),
		"New_Tai_Lue":                        reflect.ValueOf(unicode.New_Tai_Lue),
		"Newa":                               reflect.ValueOf(unicode.Newa),
		"Nko":                                reflect.ValueOf(unicode.Nko),
		"Nl":                                 reflect.ValueOf(unicode.Nl),
		"No":                                 reflect.ValueOf(unicode.No),
		"Noncharacter_Code_Point":            reflect.ValueOf(unicode.Noncharacter_Code_Point),
		"Number":                             reflect.ValueOf(unicode.Number),
		"Nushu":                              reflect.ValueOf(unicode.Nushu),
		"Nyiakeng_Puachue_Hmong":             reflect.ValueOf(unicode.Nyiakeng_Puachue_Hmong),
		"Ogham":                              reflect.ValueOf(unicode.Ogham),
		"Ol_Chiki":                           reflect.ValueOf(unicode.Ol_Chiki),
		"Old_Hungarian":                      reflect.ValueOf(unicode.Old_Hungarian),
		"Old_Italic":                         reflect.ValueOf(unicode.Old_Italic),
		"Old_North_Arabian":                  reflect.ValueOf(unicode.Old_North_Arabian),
		"Old_Permic":                         reflect.ValueOf(unicode.Old_Permic),
		"Old_Persian":                        reflect.ValueOf(unicode.Old_Persian),
		"Old_Sogdian":                        reflect.ValueOf(unicode.Old_Sogdian),
		"Old_South_Arabian":                  reflect.ValueOf(unicode.Old_South_Arabian),
		"Old_Turkic":                         reflect.ValueOf(unicode.Old_Turkic),
		"Oriya":                              reflect.ValueOf(unicode.Oriya),
		"Osage":                              reflect.ValueOf(unicode.Osage),
		"Osmanya":                            reflect.ValueOf(unicode.Osmanya),
		"Other":                              reflect.ValueOf(unicode.Other),
		"Other_Alphabetic":                   reflect.ValueOf(unicode.Other_Alphabetic),
		"Other_Default_Ignorable_Code_Point": reflect.ValueOf(unicode.Other_Default_Ignorable_Code_Point),
		"Other_Grapheme_Extend":              reflect.ValueOf(unicode.Other_Grapheme_Extend),
		"Other_ID_Continue":                  reflect.ValueOf(unicode.Other_ID_Continue),
		"Other_ID_Start":                     reflect.ValueOf(unicode.Other_ID_Start),
		"Other_Lowercase":                    reflect.ValueOf(unicode.Other_Lowercase),
		"Other_Math":                         reflect.ValueOf(unicode.Other_Math),
		"Other_Uppercase":                    reflect.ValueOf(unicode.Other_Uppercase),
		"P":                                  reflect.ValueOf(unicode.P),
		"Pahawh_Hmong":                       reflect.ValueOf(unicode.Pahawh_Hmong),
		"Palmyrene":                          reflect.ValueOf(unicode.Palmyrene),
		"Pattern_Syntax":                     reflect.ValueOf(unicode.Pattern_Syntax),
		"Pattern_White_Space":                reflect.ValueOf(unicode.Pattern_White_Space),
		"Pau_Cin_Hau":                        reflect.ValueOf(unicode.Pau_Cin_Hau),
		"Pc":                                 reflect.ValueOf(unicode.Pc),
		"Pd":                                 reflect.ValueOf(unicode.Pd),
		"Pe":                                 reflect.ValueOf(unicode.Pe),
		"Pf":                                 reflect.ValueOf(unicode.Pf),
		"Phags_Pa":                           reflect.ValueOf(unicode.Phags_Pa),
		"Phoenician":                         reflect.ValueOf(unicode.Phoenician),
		"Pi":                                 reflect.ValueOf(unicode.Pi),
		"Po":                                 reflect.ValueOf(unicode.Po),
		"Prepended_Concatenation_Mark":       reflect.ValueOf(unicode.Prepended_Concatenation_Mark),
		"PrintRanges":                        reflect.ValueOf(unicode.PrintRanges),
		"Properties":                         reflect.ValueOf(unicode.Properties),
		"Ps":                                 reflect.ValueOf(unicode.Ps),
		"Psalter_Pahlavi":                    reflect.ValueOf(unicode.Psalter_Pahlavi),
		"Punct":                              reflect.ValueOf(unicode.Punct),
		"Quotation_Mark":                     reflect.ValueOf(unicode.Quotation_Mark),
		"Radical":                            reflect.ValueOf(unicode.Radical),
		"Regional_Indicator":                 reflect.ValueOf(unicode.Regional_Indicator),
		"Rejang":                             reflect.ValueOf(unicode.Rejang),
		"Runic":                              reflect.ValueOf(unicode.Runic),
		"S":                                  reflect.ValueOf(unicode.S),
		"STerm":                              reflect.ValueOf(unicode.STerm),
		"Samaritan":                          reflect.ValueOf(unicode.Samaritan),
		"Saurashtra":                         reflect.ValueOf(unicode.Saurashtra),
		"Sc":                                 reflect.ValueOf(unicode.Sc),
		"Scripts":                            reflect.ValueOf(unicode.Scripts),
		"Sentence_Terminal":                  reflect.ValueOf(unicode.Sentence_Terminal),
		"Sharada":                            reflect.ValueOf(unicode.Sharada),
		"Shavian":                            reflect.ValueOf(unicode.Shavian),
		"Siddham":                            reflect.ValueOf(unicode.Siddham),
		"SignWriting":                        reflect.ValueOf(unicode.SignWriting),
		"Sinhala":                            reflect.ValueOf(unicode.Sinhala),
		"Sk":                                 reflect.ValueOf(unicode.Sk),
		"Sm":                                 reflect.ValueOf(unicode.Sm),
		"So":                                 reflect.ValueOf(unicode.So),
		"Soft_Dotted":                        reflect.ValueOf(unicode.Soft_Dotted),
		"Sogdian":                            reflect.ValueOf(unicode.Sogdian),
		"Sora_Sompeng":                       reflect.ValueOf(unicode.Sora_Sompeng),
		"Soyombo":                            reflect.ValueOf(unicode.Soyombo),
		"Space":                              reflect.ValueOf(unicode.Space),
		"Sundanese":                          reflect.ValueOf(unicode.Sundanese),
		"Syloti_Nagri":                       reflect.ValueOf(unicode.Syloti_Nagri),
		"Symbol":                             reflect.ValueOf(unicode.Symbol),
		"Syriac":                             reflect.ValueOf(unicode.Syriac),
		"Tagalog":                            reflect.ValueOf(unicode.Tagalog),
		"Tagbanwa":                           reflect.ValueOf(unicode.Tagbanwa),
		"Tai_Le":                             reflect.ValueOf(unicode.Tai_Le),
		"Tai_Tham":                           reflect.ValueOf(unicode.Tai_Tham),
		"Tai_Viet":                           reflect.ValueOf(unicode.Tai_Viet),
		"Takri":                              reflect.ValueOf(unicode.Takri),
		"Tamil":                              reflect.ValueOf(unicode.Tamil),
		"Tangut":                             reflect.ValueOf(unicode.Tangut),
		"Telugu":                             reflect.ValueOf(unicode.Telugu),
		"Terminal_Punctuation":               reflect.ValueOf(unicode.Terminal_Punctuation),
		"Thaana":                             reflect.ValueOf(unicode.Thaana),
		"Thai":                               reflect.ValueOf(unicode.Thai),
		"Tibetan":                            reflect.ValueOf(unicode.Tibetan),
		"Tifinagh":                           reflect.ValueOf(unicode.Tifinagh),
		"Tirhuta":                            reflect.ValueOf(unicode.Tirhuta),
		"Title":                              reflect.ValueOf(unicode.Title),
		"TurkishCase":                        reflect.ValueOf(unicode.TurkishCase),
		"Ugaritic":                           reflect.ValueOf(unicode.Ugaritic),
		"Unified_Ideograph":                  reflect.ValueOf(unicode.Unified_Ideograph),
		"Upper":                              reflect.ValueOf(unicode.Upper),
		"Vai":                                reflect.ValueOf(unicode.Vai),
		"Variation_Selector":                 reflect.ValueOf(unicode.Variation_Selector),
		"Wancho":                             reflect.ValueOf(unicode.Wancho),
		"Warang_Citi":                        reflect.ValueOf(unicode.Warang_Citi),
		"White_Space":                        reflect.ValueOf(unicode.White_Space),
		"Yi":                                 reflect.ValueOf(unicode.Yi),
		"Z":                                  reflect.ValueOf(unicode.Z),
		"Zanabazar_Square":                   reflect.ValueOf(unicode.Zanabazar_Square),
		"Zl":                                 reflect.ValueOf(unicode.Zl),
		"Zp":                                 reflect.ValueOf(unicode.Zp),
		"Zs":                                 reflect.ValueOf(unicode.Zs),

		// define functions
		"In":         reflect.ValueOf(unicode.In),
		"Is":         reflect.ValueOf(unicode.Is),
		"IsControl":  reflect.ValueOf(unicode.IsControl),
		"IsDigit":    reflect.ValueOf(unicode.IsDigit),
		"IsGraphic":  reflect.ValueOf(unicode.IsGraphic),
		"IsLetter":   reflect.ValueOf(unicode.IsLetter),
		"IsLower":    reflect.ValueOf(unicode.IsLower),
		"IsMark":     reflect.ValueOf(unicode.IsMark),
		"IsNumber":   reflect.ValueOf(unicode.IsNumber),
		"IsOneOf":    reflect.ValueOf(unicode.IsOneOf),
		"IsPrint":    reflect.ValueOf(unicode.IsPrint),
		"IsPunct":    reflect.ValueOf(unicode.IsPunct),
		"IsSpace":    reflect.ValueOf(unicode.IsSpace),
		"IsSymbol":   reflect.ValueOf(unicode.IsSymbol),
		"IsTitle":    reflect.ValueOf(unicode.IsTitle),
		"IsUpper":    reflect.ValueOf(unicode.IsUpper),
		"SimpleFold": reflect.ValueOf(unicode.SimpleFold),
		"To":         reflect.ValueOf(unicode.To),
		"ToLower":    reflect.ValueOf(unicode.ToLower),
		"ToTitle":    reflect.ValueOf(unicode.ToTitle),
		"ToUpper":    reflect.ValueOf(unicode.ToUpper),
	}
	var (
		caseRange   unicode.CaseRange
		range16     unicode.Range16
		range32     unicode.Range32
		rangeTable  unicode.RangeTable
		specialCase unicode.SpecialCase
	)
	env.PackageTypes["unicode"] = map[string]reflect.Type{
		"CaseRange":   reflect.TypeOf(&caseRange).Elem(),
		"Range16":     reflect.TypeOf(&range16).Elem(),
		"Range32":     reflect.TypeOf(&range32).Elem(),
		"RangeTable":  reflect.TypeOf(&rangeTable).Elem(),
		"SpecialCase": reflect.TypeOf(&specialCase).Elem(),
	}
}

func initUnicodeUTF16() {
	env.Packages["unicode/utf16"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Decode":      reflect.ValueOf(utf16.Decode),
		"DecodeRune":  reflect.ValueOf(utf16.DecodeRune),
		"Encode":      reflect.ValueOf(utf16.Encode),
		"EncodeRune":  reflect.ValueOf(utf16.EncodeRune),
		"IsSurrogate": reflect.ValueOf(utf16.IsSurrogate),
	}
	var ()
	env.PackageTypes["unicode/utf16"] = map[string]reflect.Type{}
}

func initUnicodeUTF8() {
	env.Packages["unicode/utf8"] = map[string]reflect.Value{
		// define constants
		"MaxRune":   reflect.ValueOf(utf8.MaxRune),
		"RuneError": reflect.ValueOf(utf8.RuneError),
		"RuneSelf":  reflect.ValueOf(utf8.RuneSelf),
		"UTFMax":    reflect.ValueOf(utf8.UTFMax),

		// define variables

		// define functions
		"DecodeLastRune":         reflect.ValueOf(utf8.DecodeLastRune),
		"DecodeLastRuneInString": reflect.ValueOf(utf8.DecodeLastRuneInString),
		"DecodeRune":             reflect.ValueOf(utf8.DecodeRune),
		"DecodeRuneInString":     reflect.ValueOf(utf8.DecodeRuneInString),
		"EncodeRune":             reflect.ValueOf(utf8.EncodeRune),
		"FullRune":               reflect.ValueOf(utf8.FullRune),
		"FullRuneInString":       reflect.ValueOf(utf8.FullRuneInString),
		"RuneCount":              reflect.ValueOf(utf8.RuneCount),
		"RuneCountInString":      reflect.ValueOf(utf8.RuneCountInString),
		"RuneLen":                reflect.ValueOf(utf8.RuneLen),
		"RuneStart":              reflect.ValueOf(utf8.RuneStart),
		"Valid":                  reflect.ValueOf(utf8.Valid),
		"ValidRune":              reflect.ValueOf(utf8.ValidRune),
		"ValidString":            reflect.ValueOf(utf8.ValidString),
	}
	var ()
	env.PackageTypes["unicode/utf8"] = map[string]reflect.Type{}
}
