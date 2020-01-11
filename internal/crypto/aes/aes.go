package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// valid AES key size
const (
	Key128Bit = 16
	Key192Bit = 24
	Key256Bit = 32
)

// AES information
const (
	IVSize    = 16
	BlockSize = 16
)

// errors
var (
	ErrInvalidIVSize      = errors.New("invalid iv size")
	ErrInvalidCipherData  = errors.New("invalid cipher data")
	ErrEmptyData          = errors.New("empty data")
	ErrInvalidPaddingSize = errors.New("invalid padding size")
)

// CBC is a AES CBC PKCS#5 encrypter
type CBC struct {
	key   []byte
	iv    []byte
	block cipher.Block
}

// NewCBC is used create a AES CBC PKCS#5 encrypter
func NewCBC(key, iv []byte) (*CBC, error) {
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := &CBC{
		key:   make([]byte, len(key)),
		iv:    make([]byte, IVSize),
		block: block,
	}
	copy(cbc.key, key)
	copy(cbc.iv, iv)
	return cbc, nil
}

// Encrypt is used to encrypt plain data
func (c *CBC) Encrypt(plainData []byte) ([]byte, error) {
	plainDataSize := len(plainData)
	if plainDataSize == 0 {
		return nil, ErrEmptyData
	}
	paddingSize := BlockSize - plainDataSize%BlockSize
	totalSize := plainDataSize + paddingSize
	plain := make([]byte, totalSize)
	copy(plain, plainData)
	padding := byte(paddingSize)
	for i := 0; i < paddingSize; i++ {
		plain[plainDataSize+i] = padding
	}
	encrypter := cipher.NewCBCEncrypter(c.block, c.iv)
	cipherData := make([]byte, totalSize)
	encrypter.CryptBlocks(cipherData, plain)
	return cipherData, nil
}

// Decrypt is used to decrypt cipher data
func (c *CBC) Decrypt(cipherData []byte) ([]byte, error) {
	cipherDataSize := len(cipherData)
	if cipherDataSize == 0 {
		return nil, ErrEmptyData
	}
	if cipherDataSize < BlockSize {
		return nil, ErrInvalidCipherData
	}
	if cipherDataSize%BlockSize != 0 {
		return nil, ErrInvalidCipherData
	}
	decrypter := cipher.NewCBCDecrypter(c.block, c.iv)
	plainData := make([]byte, cipherDataSize)
	decrypter.CryptBlocks(plainData, cipherData)
	plainDataSize := len(plainData)
	paddingSize := int(plainData[plainDataSize-1])
	offset := plainDataSize - paddingSize
	if offset < 0 {
		return nil, ErrInvalidPaddingSize
	}
	return plainData[:offset], nil
}

// KeyIV is used to get AES Key and IV
func (c *CBC) KeyIV() ([]byte, []byte) {
	key := make([]byte, len(c.key))
	iv := make([]byte, IVSize)
	copy(key, c.key)
	copy(iv, c.iv)
	return key, iv
}

// CBCEncrypt is used to encrypt plain data
func CBCEncrypt(plainData, key, iv []byte) ([]byte, error) {
	plainDataSize := len(plainData)
	if plainDataSize == 0 {
		return nil, ErrEmptyData
	}
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	paddingSize := BlockSize - plainDataSize%BlockSize
	totalSize := plainDataSize + paddingSize
	plain := make([]byte, totalSize)
	copy(plain, plainData)
	padding := byte(paddingSize)
	for i := 0; i < paddingSize; i++ {
		plain[plainDataSize+i] = padding
	}
	encrypter := cipher.NewCBCEncrypter(block, iv)
	cipherData := make([]byte, totalSize)
	encrypter.CryptBlocks(cipherData, plain)
	return cipherData, nil
}

// CBCDecrypt is used to decrypt cipher data
func CBCDecrypt(cipherData, key, iv []byte) ([]byte, error) {
	cipherDataSize := len(cipherData)
	if cipherDataSize == 0 {
		return nil, ErrEmptyData
	}
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if cipherDataSize < BlockSize {
		return nil, ErrInvalidCipherData
	}
	if cipherDataSize%BlockSize != 0 {
		return nil, ErrInvalidCipherData
	}
	decrypter := cipher.NewCBCDecrypter(block, iv)
	plainData := make([]byte, cipherDataSize)
	decrypter.CryptBlocks(plainData, cipherData)
	plainDataSize := len(plainData)
	paddingSize := int(plainData[plainDataSize-1])
	offset := plainDataSize - paddingSize
	if offset < 0 {
		return nil, ErrInvalidPaddingSize
	}
	return plainData[:offset], nil
}
