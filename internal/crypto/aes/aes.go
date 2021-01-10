package aes

import (
	"errors"
	"fmt"

	"project/internal/crypto/rand"
)

// about AES valid key size.
const (
	Key128Bit = 16
	Key192Bit = 24
	Key256Bit = 32
)

const (
	// IVSize is the AES IV size.
	IVSize = 16

	// BlockSize is the AES block size in bytes.
	BlockSize = 16
)

// errors about encrypt and decrypt.
var (
	ErrEmptyData          = errors.New("empty data")
	ErrInvalidCipherData  = errors.New("invalid aes cipher data")
	ErrInvalidPaddingSize = errors.New("invalid aes padding size")
	ErrInvalidIVSize      = errors.New("invalid iv size")
)

// AES is a aes encrypter, it can encrypt and decrypt data.
type AES interface {
	// Encrypt is used to encrypt data, it will generate iv
	// and append it in the front of output byte slice.
	Encrypt(data []byte) ([]byte, error)

	// EncryptWithIV is used to encrypt data with given iv, it
	// will not append it in the front of output byte slice.
	EncryptWithIV(data, iv []byte) ([]byte, error)

	// Decrypt is used to decrypt data, it will use iv in
	// the front of input byte slice.
	Decrypt(data []byte) ([]byte, error)

	// DecryptWithIV is used to decrypt data with given iv, it
	// will not use iv it in the front of input byte slice.
	DecryptWithIV(data, iv []byte) ([]byte, error)

	// Key is used to get aes key.
	Key() []byte
}

// GenerateIV is used to generate iv.
func GenerateIV() ([]byte, error) {
	iv := make([]byte, IVSize)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate iv: %s", err)
	}
	return iv, nil
}
