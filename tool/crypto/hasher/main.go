package main

import (
	"crypto/hmac"
	"crypto/md5"  // #nosec
	"crypto/sha1" // #nosec
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"project/internal/system"
)

var (
	alg  string
	text bool
	bin  bool
	data string
	key  []byte
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.CommandLine.Usage = printUsage

	var keyStr string

	flag.StringVar(&alg, "alg", "", "select hash algorithm")
	flag.BoolVar(&text, "text", false, "use text mode")
	flag.BoolVar(&bin, "bin", false, "use binary mode")
	flag.StringVar(&data, "data", "hello", "text message or binary file path")
	flag.StringVar(&keyStr, "key", "", "key for hmac algorithm")
	flag.Parse()

	key = []byte(keyStr)
}

func printUsage() {
	exe, err := system.ExecutableName()
	system.CheckError(err)
	const format = `
supported algorithm:
  md5, sha1, sha256, sha512,
  hmac-md5, hmac-sha1, hmac-sha256, hmac-sha512

usage:
  %s -alg "sha256" -text -data "secret"
  %s -alg "sha256" -bin -data "secret.txt"

  %s -alg "hmac-sha256" -text -data "secret" -key "hash"
  %s -alg "hmac-sha256" -bin -data "secret.txt" -key "hash"

  If -alg is empty, use all supported algorithm.
`
	fmt.Printf(format[1:], exe, exe, exe, exe)
	flag.PrintDefaults()
}

var algorithms = map[string]func() hash.Hash{
	"md5":         md5.New,
	"sha1":        sha1.New,
	"sha256":      sha256.New,
	"sha512":      sha512.New,
	"hmac-md5":    func() hash.Hash { return hmac.New(md5.New, key) },
	"hmac-sha1":   func() hash.Hash { return hmac.New(sha1.New, key) },
	"hmac-sha256": func() hash.Hash { return hmac.New(sha256.New, key) },
	"hmac-sha512": func() hash.Hash { return hmac.New(sha512.New, key) },
}

func main() {
	// select algorithm
	fn, ok := algorithms[alg]
	if !ok && alg != "" {
		system.PrintErrorf("unsupported algorithm: %s", alg)
	}
	// create data reader
	var reader io.Reader
	switch {
	case text:
		reader = strings.NewReader(data)
	case bin:
		file, err := os.Open(data)
		system.CheckError(err)
		defer func() { _ = file.Close() }()
		reader = file
	default:
		printUsage()
		return
	}
}
