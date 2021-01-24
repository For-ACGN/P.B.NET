package lsb

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
)

// errors about Reader, Writer, Encrypter and Decrypter.
var (
	// ErrNegativePosition is a error about negative position.
	ErrNegativePosition = errors.New("seek: negative position")

	// ErrInvalidOffset is a error about invalid offset.
	ErrInvalidOffset = errors.New("offset is larger than capacity")

	// ErrNoEnoughCapacity is a error that image can not write data.
	ErrNoEnoughCapacity = errors.New("image has no enough capacity for write")

	// ErrImgTooSmall is a error that means this image can't encrypt data.
	ErrImgTooSmall = errors.New("image rectangle is too small")
)

// supported lsb Writer and Reader modes.
const (
	_ Mode = iota
	PNGWithNRGBA32
	PNGWithNRGBA64
)

// supported lsb Encrypter and Decrypter algorithms.
const (
	_ Algorithm = iota
	AESWithCTR
)

// Mode is the lsb Writer and Reader mode.
type Mode uint32

func (m Mode) String() string {
	switch m {
	case PNGWithNRGBA32:
		return "PNG-NRGBA32"
	case PNGWithNRGBA64:
		return "PNG-NRGBA64"
	default:
		return fmt.Sprintf("unknown mode: %d", m)
	}
}

// Algorithm is the lsb Encrypter and Decrypter algorithm.
type Algorithm uint32

func (alg Algorithm) String() string {
	switch alg {
	case AESWithCTR:
		return "AES-CTR"
	default:
		return fmt.Sprintf("unknown algorithm: %d", alg)
	}
}

// Writer is the lsb writer interface, use Seek to write data to different area.
type Writer interface {
	// Write is used to write data to this image.
	Write(b []byte) (int, error)

	// Encode is used to encode image to writer.
	Encode(w io.Writer) error

	// Seek sets the offset for the next Write to offset.
	Seek(offset int64, whence int) (int64, error)

	// Reset is used to reset writer.
	Reset()

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can write to this image.
	Cap() int64

	// Mode is used to get the writer mode.
	Mode() Mode
}

// Reader is the lsb reader interface, use Seek to read data from different area.
type Reader interface {
	// Read is used to read data from this image.
	Read(b []byte) (int, error)

	// Seek sets the offset for the next Read to offset.
	Seek(offset int64, whence int) (int64, error)

	// Reset is used to reset reader.
	Reset()

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can read from this image.
	Cap() int64

	// Mode is used to get the reader mode.
	Mode() Mode
}

// Encrypter is the lsb encrypter interface, use Seek to encrypt data to different area.
type Encrypter interface {
	// Write is used to encrypt data and write it to under image.
	Write(b []byte) (int, error)

	// Encode is used to encode image to writer, if success, it will reset writer.
	Encode(w io.Writer) error

	// Seek sets the offset for the next Write to offset.
	Seek(offset int64, whence int) (int64, error)

	// Reset is used to reset under writer and key, if key is nil, only reset writer.
	Reset(key []byte) error

	// Key is used to get the aes key.
	Key() []byte

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can encrypt to this image.
	Cap() int64

	// Mode is used to get the mode about the under Writer.
	Mode() Mode

	// Algorithm is used to get the algorithm.
	Algorithm() Algorithm
}

// Decrypter is the lsb decrypter interface, use Seek to decrypt data from different area.
type Decrypter interface {
	// Read is used to read data from under image and decrypt it.
	Read(b []byte) (int, error)

	// Seek sets the offset for the next Read to offset.
	Seek(offset int64, whence int) (int64, error)

	// Reset is used to reset under writer and key, if key is nil, only reset reader.
	Reset(key []byte) error

	// Key is used to get the aes key.
	Key() []byte

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can decrypt from this image.
	Cap() int64

	// Mode is used to get the mode about the under Reader.
	Mode() Mode

	// Algorithm is used to get the algorithm.
	Algorithm() Algorithm
}

var decoders = map[string]func(io.Reader) (image.Image, error){
	"png":  png.Decode,
	"jpeg": jpeg.Decode,
	"gif":  gif.Decode,
}

// LoadImage is used to load image from reader.
func LoadImage(r io.Reader, ext string) (image.Image, error) {
	// use the decoder with the file name extension
	decoder, ok := decoders[strings.ToLower(ext)]
	if ok {
		return decoder(r)
	}
	// try all the decoders that supported
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	for _, decoder = range decoders {
		_, err = reader.Seek(0, io.SeekStart)
		if err != nil {
			panic(fmt.Sprintf("lsb: reset image reader: %s", err))
		}
		img, err := decoder(reader)
		if err == nil {
			return img, nil
		}
	}
	if ext != "" {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
	return nil, errors.New("unsupported image format")
}

// NewWriter is used to create a lsb writer with mode.
func NewWriter(mode Mode, img image.Image) (Writer, error) {
	var (
		writer Writer
		err    error
	)
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		writer, err = NewPNGWriter(img, mode)
	default:
		return nil, fmt.Errorf("failed to create lsb writer with %s", mode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create lsb writer: %s", err)
	}
	return writer, nil
}

// NewReader is used to create a lsb reader with mode.
func NewReader(mode Mode, r io.Reader) (Reader, error) {
	var (
		reader Reader
		err    error
	)
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		reader, err = NewPNGReader(r)
	default:
		return nil, fmt.Errorf("failed to create lsb reader with %s", mode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create lsb reader: %s", err)
	}
	return reader, nil
}

// NewEncrypter is used to create a lsb encrypter with algorithm.
func NewEncrypter(writer Writer, alg Algorithm, key []byte) (Encrypter, error) {
	var (
		enc Encrypter
		err error
	)
	switch alg {
	case AESWithCTR:
		enc, err = NewCTREncrypter(writer, key)
	default:
		return nil, fmt.Errorf("failed to create lsb encrypter with %s", alg)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create lsb encrypter: %s", err)
	}
	return enc, nil
}

// NewDecrypter is used to create a lsb decrypter with algorithm.
func NewDecrypter(reader Reader, alg Algorithm, key []byte) (Decrypter, error) {
	var (
		dec Decrypter
		err error
	)
	switch alg {
	case AESWithCTR:
		dec, err = NewCTRDecrypter(reader, key)
	default:
		return nil, fmt.Errorf("failed to create lsb decrypter with %s", alg)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create lsb decrypter: %s", err)
	}
	return dec, nil
}
