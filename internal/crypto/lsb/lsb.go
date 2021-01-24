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
	// ErrInvalidOffset is a error about invalid offset.
	ErrInvalidOffset = errors.New("offset is larger than capacity that can read/write")

	// ErrNoEnoughCapacity is a error that image can not write data.
	ErrNoEnoughCapacity = errors.New("image has no enough capacity for write")

	// ErrImgTooSmall is a error that means this image can't encrypt data.
	ErrImgTooSmall = errors.New("image rectangle is too small")
)

// supported lsb modes.
const (
	Invalid Mode = iota
	PNGWithNRGBA32
	PNGWithNRGBA64
)

// Mode is the lsb mode.
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

// Writer is the LSB writer interface.
type Writer interface {
	// Write is used to write data to this image.
	Write(b []byte) (int, error)

	// Encode is used to encode image to writer.
	Encode(w io.Writer) error

	// SetOffset is used to set pointer about position.
	SetOffset(v int64) error

	// Reset is used to reset writer.
	Reset()

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can write to this image.
	Cap() int64

	// Mode is used to get the writer mode.
	Mode() Mode
}

// Reader is the LSB reader interface.
type Reader interface {
	// Read is used to read data from this image.
	Read(b []byte) (int, error)

	// SetOffset is used to set pointer about position.
	SetOffset(v int64) error

	// Reset is used to reset reader.
	Reset()

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can read from this image.
	Cap() int64

	// Mode is used to get the reader mode.
	Mode() Mode
}

// Encrypter is the LSB encrypter interface.
type Encrypter interface {
	// Write is used to encrypt data and write it to under image.
	Write(b []byte) (int, error)

	// Encode is used to encode image to writer, if success, it will reset writer.
	Encode(w io.Writer) error

	// SetOffset is used to set pointer about data start area.
	SetOffset(v int64) error

	// Reset is used to reset under writer and key, if key is nil, only reset writer.
	Reset(key []byte) error

	// Key is used to get the aes key.
	Key() []byte

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can encrypt to this image.
	Cap() int64

	// Mode is used to get the encrypter mode.
	Mode() Mode
}

// Decrypter is the LSB decrypter interface.
type Decrypter interface {
	// Read is used to read data from under image and decrypt it.
	Read(b []byte) (int, error)

	// SetOffset is used to set pointer about data start area.
	SetOffset(v int64) error

	// Reset is used to reset under writer and key, if key is nil, only reset reader.
	Reset(key []byte) error

	// Key is used to get the aes key.
	Key() []byte

	// Image is used to get the original image.
	Image() image.Image

	// Cap is used to calculate the capacity that can decrypt from this image.
	Cap() int64

	// Mode is used to get the decrypter mode.
	Mode() Mode
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
	for _, decoder := range decoders {
		_, err = reader.Seek(0, io.SeekStart)
		if err != nil {
			panic("lsb: internal error")
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
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		return NewPNGWriter(img, mode)
	default:
		return nil, errors.New(mode.String())
	}
}

// NewReader is used to create a lsb reader with mode.
func NewReader(mode Mode, reader io.Reader) (Reader, error) {
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		return NewPNGReader(reader)
	default:
		return nil, errors.New(mode.String())
	}
}

// NewEncrypter is used to create a lsb encrypter with mode.
func NewEncrypter(mode Mode, img image.Image, key []byte) (Encrypter, error) {
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		return NewCTREncrypter(img, mode, key)
	default:
		return nil, errors.New(mode.String())
	}
}

// NewDecrypter is used to create a lsb decrypter with mode.
func NewDecrypter(mode Mode, reader io.Reader, key []byte) (Decrypter, error) {
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
		return NewCTRDecrypter(reader, key)
	default:
		return nil, errors.New(mode.String())
	}
}
