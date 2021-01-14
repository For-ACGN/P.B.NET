package lsb

import (
	"errors"
	"fmt"
	"image"
	"io"
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
	_ Mode = iota
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
