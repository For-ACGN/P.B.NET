package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"project/internal/crypto/rand"
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
