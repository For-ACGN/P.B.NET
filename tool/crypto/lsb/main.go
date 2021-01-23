package main

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
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

	flag.BoolVar(&encrypt, "enc", false, "encrypt data to a image")
	flag.BoolVar(&decrypt, "dec", false, "decrypt data from a image")
	flag.UintVar(&lsbMode, "mode", 0, "specify lsb mode (see internal/crypto/lsb/lsb.go)")
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
	imgFile, err := os.Open(imgPath)
	system.CheckError(err)
	defer func() { _ = imgFile.Close() }()
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
	// use default mode
	mode := lsb.Mode(lsbMode)
	if mode == 0 {
		mode = lsb.PNGWithNRGBA32
	}
	switch {
	case encrypt:
		encryptImage(mode, img, reader)
	case decrypt:
		decryptImage(mode, img)
	default:
		printUsage()
	}
}

func encryptImage(mode lsb.Mode, img image.Image, src io.Reader) {
	// create lsb encrypter
	aesKey := calculateAESKey()
	var (
		encrypter lsb.Encrypter
		err       error
	)
	switch mode {
	case lsb.PNGWithNRGBA32, lsb.PNGWithNRGBA64:
		encrypter, err = lsb.NewPNGEncrypter(img, mode, aesKey)
		system.CheckError(err)
	default:
		system.PrintError(mode)
	}
	// set offset
	if offset != 0 {
		err = encrypter.SetOffset(offset)
		system.CheckError(err)
	}
	// compress plain data and encrypted to image
	writer, err := flate.NewWriter(encrypter, flate.BestCompression)
	system.CheckError(err)
	_, err = io.Copy(writer, src)
	system.CheckError(err)
	err = writer.Close()
	system.CheckError(err)
	// write file
	if output == "" {
		output = "enc.png"
	}
	outputFile, err := os.Create(output)
	system.CheckError(err)
	defer func() { _ = outputFile.Close() }()
	err = encrypter.Encode(outputFile)
	system.CheckError(err)
}

func decryptImage(mode lsb.Mode, img image.Image) {
	// check mode first
	switch {
	case textMode, binMode:
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
	case binMode:
		if output == "" {
			output = "file.txt"
		}
		err = system.WriteFile(output, buf.Bytes())
		system.CheckError(err)
	}
}

func calculateAESKey() []byte {
	pwd := []byte(key)
	salt := []byte("lsb")
	return pbkdf2.Key(pwd, salt, 4096, aes.Key256Bit, sha256.New)
}
