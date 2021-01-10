package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"project/internal/crypto/rand"
	"project/internal/security"
)

// CBCEncrypt is used to encrypt plain data with cipher block chaining
// mode with PKCS#7. Output is [IV + cipher data].
func CBCEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cbcEncrypt(block, data)
}

func cbcEncrypt(block cipher.Block, data []byte) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to generate iv: %s", err)
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

// CBCEncryptWithIV is used to encrypt plain data with cipher block chaining
// mode with PKCS#7. Output is cipher data, not include iv.
func CBCEncryptWithIV(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cbcEncryptWithIV(block, data, iv)
}

func cbcEncryptWithIV(block cipher.Block, data, iv []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	// make buffer
	paddingSize := BlockSize - l%BlockSize
	output := make([]byte, l+paddingSize)
	// copy plain data and padding data
	copy(output, data)
	padding := byte(paddingSize)
	for i := 0; i < paddingSize; i++ {
		output[l+i] = padding
	}
	// encrypt plain data
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(output, output)
	return output, nil
}

// CBCDecrypt is used to decrypt cipher data with cipher block chaining
// mode with PKCS#7. Input data is [IV + cipher data].
func CBCDecrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cbcDecrypt(block, data)
}

func cbcDecrypt(block cipher.Block, data []byte) ([]byte, error) {
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
	iv := data[:IVSize]
	// decrypt cipher data
	decrypter := cipher.NewCBCDecrypter(block, iv)
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

// CBCDecryptWithIV is used to decrypt cipher data with cipher block chaining
// mode with PKCS#7, Input data is cipher data.
func CBCDecryptWithIV(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cbcDecryptWithIV(block, data, iv)
}

func cbcDecryptWithIV(block cipher.Block, data, iv []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if l < BlockSize {
		return nil, ErrInvalidCipherData
	}
	if l%BlockSize != 0 {
		return nil, ErrInvalidCipherData
	}
	output := make([]byte, l)
	// decrypt cipher data
	decrypter := cipher.NewCBCDecrypter(block, iv)
	decrypter.CryptBlocks(output, data)
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
func NewCBC(key []byte) (AES, error) {
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
	return cbcEncrypt(cbc.block, data)
}

// EncryptWithIV used to encrypt data with given iv. Output is cipher data.
func (cbc *CBC) EncryptWithIV(data, iv []byte) ([]byte, error) {
	return cbcEncryptWithIV(cbc.block, data, iv)
}

// Decrypt is used to decrypt cipher data. Input data is [IV + cipher data].
func (cbc *CBC) Decrypt(data []byte) ([]byte, error) {
	return cbcDecrypt(cbc.block, data)
}

// DecryptWithIV is used to decrypt cipher data with given iv. Input data is cipher data.
func (cbc *CBC) DecryptWithIV(data, iv []byte) ([]byte, error) {
	return cbcDecryptWithIV(cbc.block, data, iv)
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
