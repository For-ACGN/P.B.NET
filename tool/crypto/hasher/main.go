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

	"project/internal/system"
)

var algorithms = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha224": sha256.New224,
	"sha256": sha256.New,
	"sha384": sha512.New384,
	"sha512": sha512.New,
}

var (
	alg  string
	text string
	file string
	key  []byte
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.CommandLine.Usage = printUsage

	var keyStr string

	flag.StringVar(&alg, "alg", "", "select hash algorithm")
	flag.StringVar(&text, "text", "hello", "input text message")
	flag.StringVar(&file, "file", "", "input file")
	flag.StringVar(&keyStr, "key", "", "key for hmac algorithm")
	flag.Parse()

	// add hmac algorithm
	key = []byte(keyStr)
	for a, fn := range algorithms {
		f := fn
		algorithms["hmac-"+a] = func() hash.Hash {
			return hmac.New(f, key)
		}
	}
}

func printUsage() {
	exe, err := system.ExecutableName()
	system.CheckError(err)
	const format = `
supported algorithm:
  md5, sha1, sha224, sha256, sha384, sha512,
  hmac-md5, hmac-sha1, hmac-sha224, hmac-sha256
  hmac-sha384, hmac-sha512

usage:
  %s -alg "sha256" -text "hello"
  %s -alg "sha256" -file "f.txt"

  %s -alg "hmac-sha256" -text "hello" -key "hash"
  %s -alg "hmac-sha256" -file "f.txt" -key "hash"

  If -alg is empty, use all supported algorithm.
`
	fmt.Printf(format[1:]+"\n", exe, exe, exe, exe)
	flag.PrintDefaults()
}

func main() {
	// select algorithm
	fn, ok := algorithms[alg]
	if !ok && alg != "" {
		system.PrintErrorf("unsupported algorithm: %s", alg)
	}
	// check key if use hmac
	if strings.Contains(alg, "hmac") && len(key) == 0 {
		fmt.Printf("[warning] use %s algorithm, but key is empty\n\n", alg)
	}
	// create data reader
	var rs io.ReadSeeker
	if file != "" {
		f, err := os.Open(file)
		system.CheckError(err)
		defer func() { _ = f.Close() }()
		rs = f
	} else {
		rs = strings.NewReader(text)
	}
	// use single algorithm
	if fn != nil {
		h := fn()
		_, err := io.Copy(h, rs)
		system.CheckError(err)
		digest := h.Sum(nil)
		printAlgorithmAndDigest(alg, digest)
		return
	}
	// use all algorithms
	for alg, fn := range algorithms {
		h := fn()
		_, err := io.Copy(h, rs)
		system.CheckError(err)
		digest := h.Sum(nil)
		printAlgorithmAndDigest(alg, digest)
	}
}

func printAlgorithmAndDigest(alg string, digest []byte) {
	hexStr := hex.EncodeToString(digest)
	hexStrUp := strings.ToUpper(hexStr)

	fmt.Println("algorithm:", alg)
	fmt.Printf("  raw:     %v\n", digest)
	fmt.Printf("  hex-low: %s\n", hexStr)
	fmt.Printf("  hex-up:  %s\n", hexStrUp)
	fmt.Printf("  base64:  %s\n", base64.StdEncoding.EncodeToString(digest))

	if strings.Contains(alg, "md5") {
		fmt.Println()
		fmt.Printf("  hex-low 16: %s\n", hexStr[8:24])
		fmt.Printf("  hex-up  16: %s\n", hexStrUp[8:24])
	}

	fmt.Println()
}
