package aes

import (
	"errors"
)

// about AES valid key size.
const (
	Key128Bit int = 16
	Key192Bit int = 24
	Key256Bit int = 32
)

const (
	// IVSize is the AES IV size.
	IVSize int = 16

	// BlockSize is the AES block size in bytes.
	BlockSize int = 16
)

// errors about encrypt and decrypt.
var (
	ErrEmptyData          = errors.New("empty data")
	ErrInvalidCipherData  = errors.New("invalid aes cipher data")
	ErrInvalidPaddingSize = errors.New("invalid aes padding size")
)

// Encrypter is a aes encrypter, it can encrypt and decrypt data.
type Encrypter interface {
	// Encrypt is used to encrypt data.
	Encrypt(data []byte) ([]byte, error)

	// Decrypt is used to decrypt data.
	Decrypt(data []byte) ([]byte, error)

	// Key is used to get aes key.
	Key() []byte
}
