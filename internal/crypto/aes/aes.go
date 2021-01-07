package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"project/internal/security"
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
	ErrEmptyData         = errors.New("empty data")
	ErrInvalidCipherData = errors.New("invalid aes cipher data")

	ErrInvalidIVSize      = errors.New("invalid aes iv size")
	ErrInvalidPaddingSize = errors.New("invalid aes padding size")
)

// CBC is a AES CBC PKCS#5 encrypter.
type CBC struct {
	key   *security.Bytes
	iv    *security.Bytes
	block cipher.Block
}

// NewCBC is used create a AES CBC PKCS#5 encrypter.
func NewCBC(key, iv []byte) (*CBC, error) {
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := &CBC{
		key:   security.NewBytes(key),
		iv:    security.NewBytes(iv),
		block: block,
	}
	return cbc, nil
}

// Encrypt is used to encrypt plain data.
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
	// get iv
	iv := c.iv.Get()
	defer c.iv.Put(iv)
	// encrypt
	encrypter := cipher.NewCBCEncrypter(c.block, iv)
	cipherData := make([]byte, totalSize)
	encrypter.CryptBlocks(cipherData, plain)
	return cipherData, nil
}

// Decrypt is used to decrypt cipher data.
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
	// get iv
	iv := c.iv.Get()
	defer c.iv.Put(iv)
	// decrypt
	decrypter := cipher.NewCBCDecrypter(c.block, iv)
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

// KeyIV is used to get AES Key and IV.
func (c *CBC) KeyIV() ([]byte, []byte) {
	// get key and iv
	key := c.key.Get()
	defer c.key.Put(key)
	iv := c.iv.Get()
	defer c.iv.Put(iv)
	// copy it, usually cover it after use.
	keyCp := make([]byte, len(key))
	ivCp := make([]byte, IVSize)
	copy(keyCp, key)
	copy(ivCp, iv)
	return keyCp, ivCp
}

// CBCEncrypt is used to encrypt plain data.
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

// CBCDecrypt is used to decrypt cipher data.
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
