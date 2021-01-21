package main

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"project/internal/crypto/aes"
	"project/internal/crypto/lsb"
	"project/internal/system"
)

var (
	encrypt    bool
	decrypt    bool
	textMode   bool
	binaryMode bool
	lsbMode    lsb.Mode
	imgPath    string
	data       string
	offset     int64
	password   string
	output     string
)

func init() {
	var mode uint

	flag.CommandLine.SetOutput(os.Stdout)
	flag.CommandLine.Usage = printUsage

	flag.BoolVar(&encrypt, "enc", false, "encrypt data to a png file")
	flag.BoolVar(&decrypt, "dec", false, "decrypt data from a png file")
	flag.BoolVar(&textMode, "text", false, "use text mode")
	flag.BoolVar(&binaryMode, "bin", false, "use binary mode")
	flag.UintVar(&mode, "mode", 0, "set lsb mode (see internal/crypto/lsb/lsb.go)")
	flag.StringVar(&imgPath, "img", "", "raw or encrypted image file path")
	flag.StringVar(&data, "data", "", "text message or binary file path for encrypt")
	flag.Int64Var(&offset, "offset", 0, "set offset for encrypter or decrypter")
	flag.StringVar(&password, "pwd", "lsb", "password for encrypt or decrypt data")
	flag.StringVar(&output, "output", "", "output file path")
	flag.Parse()

	// set default lsb mode
	lsbMode = lsb.Mode(mode)
	if lsbMode == 0 {
		lsbMode = lsb.PNGWithNRGBA32
	}
}

func printUsage() {
	exe, err := system.ExecutableName()
	system.CheckError(err)
	const format = `
usage:

 [encrypt]
   text mode:   %s -enc -text -img "raw.png" -data "secret" -pwd "pass"
   binary mode: %s -enc -bin -img "raw.png" -data "secret.txt" -pwd "pass"

 [decrypt]
   text mode:   %s -dec -text -img "enc.png" -pwd "pass"
   binary mode: %s -dec -bin -img "enc.png" -pwd "pass"

`
	fmt.Printf(format[1:], exe, exe, exe, exe)
	flag.PrintDefaults()
}

func main() {
	switch {
	case encrypt:
		encryptData()
	case decrypt:
		decryptData()
	default:
		printUsage()
	}
}

func encryptData() {
	// create plain data reader
	var reader io.Reader
	switch {
	case textMode:
		reader = strings.NewReader(data)
	case binaryMode:
		file, err := os.Open(data) // #nosec
		system.CheckError(err)
		defer func() { _ = file.Close() }()
		reader = file
	default:
		system.PrintError("select text or binary mode")
	}
	// read image
	switch lsbMode {
	case lsb.PNGWithNRGBA32:
		file, err := os.Open(imgPath)
		system.CheckError(err)
		defer func() { _ = file.Close() }()

	}

	// compress plain data
	buf := bytes.NewBuffer(make([]byte, 0, len(reader)/2))
	writer, err := flate.NewWriter(buf, flate.BestCompression)
	system.CheckError(err)
	_, err = writer.Write(reader)
	system.CheckError(err)
	err = writer.Close()
	system.CheckError(err)
	// encrypt
	key, iv := generateAESKeyIV()
	pngEnc, err := lsb.EncryptToPNG(img, buf.Bytes(), key, iv)
	system.CheckError(err)
	// write file
	if output == "" {
		output = "enc.png"
	}
	err = system.WriteFile(output, pngEnc)
	system.CheckError(err)
}

func decryptData() {
	// check mode first
	switch {
	case textMode, binaryMode:
	default:
		fmt.Println("select text or binary mode")
	}
	png := readPNG()
	key, iv := generateAESKeyIV()
	plainData, err := lsb.DecryptFromPNG(png, key, iv)
	system.CheckError(err)
	// decompress plain data
	reader := flate.NewReader(bytes.NewReader(plainData))
	buf := bytes.NewBuffer(make([]byte, 0, len(plainData)*2))
	_, err = buf.ReadFrom(reader)
	system.CheckError(err)
	err = reader.Close()
	system.CheckError(err)
	// handle data
	switch {
	case textMode:
		fmt.Println(buf)
	case binaryMode:
		if output == "" {
			output = "file.txt"
		}
		err = system.WriteFile(output, buf.Bytes())
		system.CheckError(err)
	}
}

func generateAESKeyIV() ([]byte, []byte) {
	hash := sha256.Sum256([]byte(password))
	return hash[:], hash[:aes.IVSize]
}
