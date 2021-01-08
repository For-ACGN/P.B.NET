package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"project/internal/crypto/rand"
	"project/internal/security"
)

// CBCEncrypt is used to encrypt plain data with cipher block chaining
// mode with PKCS#7, Output is [IV + cipher data].
func CBCEncrypt(data, key []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// make buffer
	paddingSize := BlockSize - l%BlockSize
	output := make([]byte, IVSize+l+paddingSize)
	// generate random iv
	iv := output[:IVSize]
	_, err = rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random iv: %s", err)
	}
	// copy plain data and padding data
	copy(output[IVSize:], data)
	padding := byte(paddingSize)
	for i := 0; i < paddingSize; i++ {
		output[IVSize+l+i] = padding
	}
	// encrypt plain data
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(output[IVSize:], output[IVSize:])
	return output, nil
}

// CBCDecrypt is used to decrypt cipher data with cipher block chaining
// mode with PKCS#7, Input data is [IV + cipher data].
func CBCDecrypt(data, key []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if l-IVSize < BlockSize {
		return nil, ErrInvalidCipherData
	}
	if (l-IVSize)%BlockSize != 0 {
		return nil, ErrInvalidCipherData
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, l-IVSize)
	// decrypt cipher data
	decrypter := cipher.NewCBCDecrypter(block, data[:IVSize])
	decrypter.CryptBlocks(output, data[IVSize:])
	// remove padding data
	outputSize := len(output)
	paddingSize := int(output[outputSize-1])
	offset := outputSize - paddingSize
	if offset < 0 {
		return nil, ErrInvalidPaddingSize
	}
	return output[:offset], nil
}

// CBC is the AES encrypter with cipher block chaining mode.
type CBC struct {
	key   *security.Bytes
	block cipher.Block
}

// NewCBC is used create a AES CBC PKCS#5 encrypter.
func NewCBC(key []byte) (*CBC, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := CBC{
		key:   security.NewBytes(key),
		block: block,
	}
	return &cbc, nil
}

// Encrypt is used to encrypt plain data. Output is [IV + cipher data].
func (cbc *CBC) Encrypt(data []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	// make buffer
	paddingSize := BlockSize - l%BlockSize
	output := make([]byte, IVSize+l+paddingSize)
	// generate random iv
	iv := output[:IVSize]
	_, err := rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random iv: %s", err)
	}
	// copy plain data and padding data
	copy(output[IVSize:], data)
	padding := byte(paddingSize)
	for i := 0; i < paddingSize; i++ {
		output[IVSize+l+i] = padding
	}
	// encrypt plain data
	encrypter := cipher.NewCBCEncrypter(cbc.block, iv)
	encrypter.CryptBlocks(output[IVSize:], output[IVSize:])
	return output, nil
}

// Decrypt is used to decrypt cipher data. Input data is [IV + cipher data].
func (cbc *CBC) Decrypt(data []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if l-IVSize < BlockSize {
		return nil, ErrInvalidCipherData
	}
	if (l-IVSize)%BlockSize != 0 {
		return nil, ErrInvalidCipherData
	}
	output := make([]byte, l-IVSize)
	// decrypt cipher data
	decrypter := cipher.NewCBCDecrypter(cbc.block, data[:IVSize])
	decrypter.CryptBlocks(output, data[IVSize:])
	// remove padding data
	outputSize := len(output)
	paddingSize := int(output[outputSize-1])
	offset := outputSize - paddingSize
	if offset < 0 {
		return nil, ErrInvalidPaddingSize
	}
	return output[:offset], nil
}

// Key is used to get AES Key.
func (cbc *CBC) Key() []byte {
	key := cbc.key.Get()
	defer cbc.key.Put(key)
	// copy it, usually cover it after use.
	cp := make([]byte, len(key))
	copy(cp, key)
	return cp
}
