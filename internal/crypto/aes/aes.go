package aes

import (
	"errors"
)

// about valid AES key size.
const (
	Key128Bit int = 16
	Key192Bit int = 24
	Key256Bit int = 32
)

// about AES information.
const (
	IVSize    int = 16
	BlockSize int = 16
)

// errors about encrypt and decrypt.
var (
	ErrEmptyData          = errors.New("empty data")
	ErrInvalidCipherData  = errors.New("invalid aes cipher data")
	ErrInvalidPaddingSize = errors.New("invalid aes padding size")
)
