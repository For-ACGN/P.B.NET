package main

import (
	"crypto/hmac"
	"crypto/md5"  // #nosec
	"crypto/sha1" // #nosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"project/internal/convert"
	"project/internal/system"
)

var algorithmIdx = []string{
	"md5",
	"sha1",
	"sha224",
	"sha256",
	"sha384",
	"sha512",
}

var algorithms = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha224": sha256.New224,
	"sha256": sha256.New,
	"sha384": sha512.New384,
	"sha512": sha512.New,
}

var (
	algo string
	text string
	file string
	key  string
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.CommandLine.Usage = printUsage

	flag.StringVar(&algo, "algo", "", "specify hash algorithm")
	flag.StringVar(&text, "text", "", "input text message")
	flag.StringVar(&file, "file", "", "input file path")
	flag.StringVar(&key, "key", "", "key for hmac algorithm")
	flag.Parse()

	// add hmac algorithm
	algList := make([]string, len(algorithmIdx))
	keyBytes := []byte(key)
	for i, a := range algorithmIdx {
		fn, ok := algorithms[a]
		if !ok {
			panic(fmt.Sprintf("algorithm %s is not exist", a))
		}
		algorithms["hmac-"+a] = func() hash.Hash {
			return hmac.New(fn, keyBytes)
		}
		algList[i] = "hmac-" + a
	}
	algorithmIdx = append(algorithmIdx, algList...)
}

func printUsage() {
	exe, err := system.ExecutableName()
	system.CheckError(err)
	const format = `
supported algorithms:

  md5, sha1, sha224, sha256, sha384, sha512,
  hmac-md5, hmac-sha1, hmac-sha224, hmac-sha256
  hmac-sha384, hmac-sha512

usage:

  %s -algo "sha256" -text "hello"
  %s -algo "sha256" -file "f.txt"
  %s -algo "hmac-sha256" -text "hello" -key "hash"
  %s -algo "hmac-sha256" -file "f.txt" -key "hash"

  If -algo is empty, use all supported algorithms.

arguments:
`
	const n = 4
	exes := make([]interface{}, n)
	for i := 0; i < n; i++ {
		exes[i] = exe
	}
	fmt.Printf(format[1:]+"\n", exes...)
	flag.PrintDefaults()
}

func main() {
	// check arguments
	if len(os.Args) == 1 {
		printUsage()
		return
	}
	// specify algorithm
	fn, ok := algorithms[algo]
	if !ok && algo != "" {
		system.PrintErrorf("unsupported algorithm: %s", algo)
	}
	// check key if use hmac
	if strings.Contains(algo, "hmac") && len(key) == 0 {
		fmt.Printf("[warning] use %s algorithm, but key is empty\n\n", algo)
	}
	// create data reader
	var (
		rs     io.ReadSeeker
		buffer []byte
	)
	if file != "" {
		f, err := os.Open(file)
		system.CheckError(err)
		defer func() { _ = f.Close() }()
		rs = f
		buffer = make([]byte, 1024*1024)
	} else {
		rs = strings.NewReader(text)
		buffer = make([]byte, 1024)
	}
	// use specific algorithm
	if fn != nil {
		h := fn()
		_, err := io.CopyBuffer(h, rs, buffer)
		system.CheckError(err)
		digest := h.Sum(nil)
		printAlgorithmAndDigest(algo, digest)
		return
	}
	// use all algorithms
	for _, a := range algorithmIdx {
		h := algorithms[a]()
		_, err := io.CopyBuffer(h, rs, buffer)
		system.CheckError(err)
		digest := h.Sum(nil)
		printAlgorithmAndDigest(a, digest)
		_, err = rs.Seek(0, io.SeekStart)
		system.CheckError(err)
	}
}

func printAlgorithmAndDigest(algo string, digest []byte) {
	hexStr := hex.EncodeToString(digest)
	hexStrUp := strings.ToUpper(hexStr)
	base64Str := base64.StdEncoding.EncodeToString(digest)

	prefix := strings.Repeat(" ", 11)
	hexLow := convert.SdumpStringWithPL(hexStr, prefix, 64)
	hexUp := convert.SdumpStringWithPL(hexStrUp, prefix, 64)
	base64St := convert.SdumpStringWithPL(base64Str, prefix, 64)
	bytesStr := convert.SdumpBytesWithPL(digest, prefix, 8)

	fmt.Println("algorithm:", algo)
	fmt.Printf("  hex-low: %s\n\n", convert.RemoveFirstPrefix(hexLow, prefix))
	fmt.Printf("  hex-up:  %s\n\n", convert.RemoveFirstPrefix(hexUp, prefix))
	fmt.Printf("  base64:  %s\n\n", convert.RemoveFirstPrefix(base64St, prefix))
	fmt.Printf("  bytes:   %s\n\n", convert.RemoveFirstPrefix(bytesStr, prefix))

	if strings.Contains(algo, "md5") {
		fmt.Println("  ------------------------------------------------")
		fmt.Println("  hex-low 16:", hexStr[8:24])
		fmt.Println("  hex-up  16:", hexStrUp[8:24])
		fmt.Println("  ------------------------------------------------")
		fmt.Println()
	}
}
