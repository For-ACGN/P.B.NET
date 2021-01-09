package lsb

import (
	"fmt"
	"image"
	"io"
)

// supported lsb modes.
const (
	PNGWithNRGBA32 Mode = iota
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
	// StorageSize is used to calculate the size that can write to this image.
	StorageSize() uint64

	// Image is used to get the inner image that will be encoded.
	Image() image.Image

	// Write is used to write data to this image.
	Write(data []byte) (int, error)

	// Encode is used to encode image to writer.
	Encode(w io.Writer) error

	// Reset is used to reset writer.
	Reset()

	// Mode is used to get the writer mode.
	Mode() Mode
}

// Reader is the LSB reader interface.
type Reader interface {
	// StorageSize is used to calculate the size that can read from this image.
	StorageSize() uint64

	// Image is used to get the inner image.
	Image() image.Image

	// Read is used to read data from this image.
	Read(data []byte) (int, error)

	// Mode is used to get the reader mode.
	Mode() Mode
}

// Encrypter is the LSB encrypter interface.
type Encrypter interface {
	// StorageSize is used to calculate the size that can write to this image.
	StorageSize() uint64

	// EncryptTo is used to encrypt data with AES-CTR and write to io.Writer.
	EncryptTo(w io.Writer, data, key []byte) error

	// EncryptFrom is used to read data from io.Reader then encrypt data
	// with AES-CTR and write to io.Writer.
	EncryptFrom(w io.Writer, r io.Reader, key []byte) error

	// Encrypt is used to encrypt data with AES-CTR and write to byte slice.
	Encrypt(data, key []byte) ([]byte, error)

	// Mode is used to get the encrypter mode.
	Mode() Mode
}

// Decrypter is the LSB decrypter interface.
type Decrypter interface {
	// DecryptTo is used to decrypt data with AES-CTR and write to io.Writer.
	DecryptTo(w io.Writer, key []byte) error

	// Decrypt is used to decrypt data with AES-CTR and write to byte slice.
	Decrypt(key []byte) ([]byte, error)

	// Mode is used to get the decrypter mode.
	Mode() Mode
}
