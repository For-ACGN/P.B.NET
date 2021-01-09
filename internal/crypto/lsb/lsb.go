package lsb

import (
	"image"
	"io"
)

// Writer is the LSB writer interface.
type Writer interface {
	// StorageSize is used to calculate the size that can write to this image.
	StorageSize() uint64

	// Image is used to get the inner image.
	Image() image.Image

	// Write is used to write data to this image.
	Write(data []byte) (int, error)

	// Encode is used to encode image to writer.
	Encode(w io.Writer) error
}

// Reader is the LSB reader interface.
type Reader interface {
	// StorageSize is used to calculate the size that can read from this image.
	StorageSize() uint64

	// Image is used to get the inner image.
	Image() image.Image

	// Read is used to read data from this image.
	Read(data []byte) (int, error)
}

// Encrypter is the LSB encrypter interface.
type Encrypter interface {
	// EncryptTo is used to encrypt data with AES-CTR and write to io.Writer.
	EncryptTo(w io.Writer, data, key []byte) error

	// EncryptFrom is used to read data from io.Reader then encrypt data
	// with AES-CTR and write to io.Writer.
	EncryptFrom(w io.Writer, r io.Reader, key []byte) error

	// Encrypt is used to encrypt data with AES-CTR and write to byte slice.
	Encrypt(data, key []byte) ([]byte, error)
}

// Decrypter is the LSB decrypter interface.
type Decrypter interface {
	// DecryptTo is used to decrypt data with AES-CTR and write to io.Writer.
	DecryptTo(w io.Writer, key []byte) error

	// Decrypt is used to decrypt data with AES-CTR and write to byte slice.
	Decrypt(key []byte) ([]byte, error)
}
