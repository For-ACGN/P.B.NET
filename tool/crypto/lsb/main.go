package main

import (
	"compress/flate"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"project/internal/crypto/aes"
	"project/internal/crypto/lsb"
	"project/internal/system"
)

var (
	encrypt  bool
	decrypt  bool
	lsbMode  uint
	lsbAlg   uint
	imgPath  string
	text     string
	filePath string
	offset   int64
	key      string
	output   string
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.CommandLine.Usage = printUsage

	flag.BoolVar(&encrypt, "enc", false, "encrypt data to a original image")
	flag.BoolVar(&decrypt, "dec", false, "decrypt data from a encrypted image")
	flag.UintVar(&lsbMode, "mode", 0, "specify lsb writer or reader mode")
	flag.UintVar(&lsbAlg, "alg", 0, "specify lsb encrypter or decrypter algorithm")
	flag.StringVar(&imgPath, "img", "", "original or encrypted image file path")
	flag.StringVar(&text, "text", "", "text message that will be encrypted")
	flag.StringVar(&filePath, "file", "", "file that will be encrypted")
	flag.Int64Var(&offset, "offset", 0, "set offset for encrypter or decrypter")
	flag.StringVar(&key, "key", "lsb", "password for encrypt or decrypt data")
	flag.StringVar(&output, "output", "", "output encrypted image or secret file path")
	flag.Parse()
}

func printUsage() {
	exe, err := system.ExecutableName()
	system.CheckError(err)
	const format = `
supported modes:
    1: PNG-NRGBA32
    2: PNG-NRGBA64
  
supported algorithms:
    1: AES-CTR

usage:
  [encrypt]
    text mode: %s -enc -img "raw.png" -text "secret" -key "lsb" -output "enc.png"
    file mode: %s -enc -img "raw.png" -file "se.txt" -key "lsb" -output "enc.png"

  [decrypt]
    text mode: %s -dec -img "enc.png" -key "lsb"
    file mode: %s -dec -img "enc.png" -key "lsb" -output "secret.txt"
`
	fmt.Printf(format[1:]+"\n", exe, exe, exe, exe)
	flag.PrintDefaults()
}

func main() {
	// check arguments
	if len(os.Args) == 1 {
		printUsage()
		return
	}
	// load image file
	img, err := os.Open(imgPath)
	system.CheckError(err)
	defer func() { _ = img.Close() }()
	// use default mode
	mode := lsb.Mode(lsbMode)
	if mode == 0 {
		mode = lsb.PNGWithNRGBA32
	}
	// use default algorithm
	alg := lsb.Algorithm(lsbAlg)
	if alg == 0 {
		alg = lsb.AESWithCTR
	}
	switch {
	case encrypt:
		encryptImage(mode, alg, img)
	case decrypt:
		decryptImage(mode, alg, img)
	default:
		printUsage()
	}
}

func encryptImage(mode lsb.Mode, alg lsb.Algorithm, imgFile io.Reader) {
	ext := filepath.Ext(imgPath)
	img, err := lsb.LoadImage(imgFile, ext)
	system.CheckError(err)
	// create data reader
	var reader io.Reader
	if filePath != "" {
		file, err := os.Open(filePath)
		system.CheckError(err)
		defer func() { _ = file.Close() }()
		reader = file
	} else {
		reader = strings.NewReader(text)
	}
	// create lsb encrypter
	writer, err := lsb.NewWriter(mode, img)
	system.CheckError(err)
	aesKey := calculateAESKey()
	encrypter, err := lsb.NewEncrypter(writer, alg, aesKey)
	system.CheckError(err)
	// set offset
	if offset != 0 {
		err = encrypter.SetOffset(offset)
		system.CheckError(err)
	}
	// compress plain data and encrypted to image
	dw, err := flate.NewWriter(encrypter, flate.BestCompression)
	system.CheckError(err)
	_, err = io.Copy(dw, reader)
	system.CheckError(err)
	err = dw.Close()
	system.CheckError(err)
	// save output image file
	if output == "" {
		output = "enc.png"
	}
	outputFile, err := os.Create(output)
	system.CheckError(err)
	defer func() { _ = outputFile.Close() }()
	err = encrypter.Encode(outputFile)
	system.CheckError(err)
}

func decryptImage(mode lsb.Mode, alg lsb.Algorithm, img io.Reader) {
	// create lsb decrypter
	reader, err := lsb.NewReader(mode, img)
	system.CheckError(err)
	aesKey := calculateAESKey()
	decrypter, err := lsb.NewDecrypter(reader, alg, aesKey)
	system.CheckError(err)
	// set offset
	if offset != 0 {
		err = decrypter.SetOffset(offset)
		system.CheckError(err)
	}
	// set output writer
	var (
		dst io.Writer
		buf []byte
	)
	if output != "" {
		file, err := os.Create(output)
		system.CheckError(err)
		defer func() { _ = file.Close() }()
		dst = file
		buf = make([]byte, 1024*1024)
	} else {
		dst = os.Stdout
		buf = make([]byte, 1024)
	}
	// decrypt and decompress plain data
	dr := flate.NewReader(decrypter)
	_, err = io.CopyBuffer(dst, dr, buf)
	system.CheckError(err)
	err = dr.Close()
	system.CheckError(err)
}

func calculateAESKey() []byte {
	pwd := []byte(key)
	salt := []byte("lsb")
	return pbkdf2.Key(pwd, salt, 4096, aes.Key256Bit, sha256.New)
}
