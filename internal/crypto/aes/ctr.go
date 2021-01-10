package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"project/internal/crypto/rand"
	"project/internal/security"
)

// CTREncrypt is used to encrypt plain data with counter mode. Output is [IV + cipher data].
func CTREncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ctrEncrypt(block, data)
}

func ctrEncrypt(block cipher.Block, data []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	// make buffer
	output := make([]byte, IVSize+l)
	// generate random iv
	iv := output[:IVSize]
	_, err := rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate iv: %s", err)
	}
	// encrypt plain data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output[IVSize:], data)
	return output, nil
}

// CTREncryptWithIV is used to encrypt plain data with counter mode. Output is cipher data.
func CTREncryptWithIV(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ctrEncryptWithIV(block, data, iv)
}

func ctrEncryptWithIV(block cipher.Block, data, iv []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	// make buffer
	output := make([]byte, l)
	// encrypt plain data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output, data)
	return output, nil
}

// CTRDecrypt is used to decrypt cipher data with counter mode. Input is [IV + cipher data].
func CTRDecrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ctrDecrypt(block, data)
}

func ctrDecrypt(block cipher.Block, data []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if l < IVSize+1 {
		return nil, ErrInvalidCipherData
	}
	output := make([]byte, l-IVSize)
	iv := data[:IVSize]
	// decrypt cipher data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output, data[IVSize:])
	return output, nil
}

// CTRDecryptWithIV is used to decrypt cipher data with counter mode. Input is cipher data.
func CTRDecryptWithIV(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ctrDecryptWithIV(block, data, iv)
}

func ctrDecryptWithIV(block cipher.Block, data, iv []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	if len(iv) != IVSize {
		return nil, ErrInvalidIVSize
	}
	output := make([]byte, l)
	// decrypt cipher data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output, data)
	return output, nil
}

// CTR is the AES encrypter with counter mode.
type CTR struct {
	key   *security.Bytes
	block cipher.Block
}

// NewCTR is used to create a AES encrypter with counter mode.
func NewCTR(key []byte) (AES, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ctr := CTR{
		key:   security.NewBytes(key),
		block: block,
	}
	return &ctr, nil
}

// Encrypt is used to encrypt plain data. Output is [IV + cipher data].
func (ctr *CTR) Encrypt(data []byte) ([]byte, error) {
	return ctrEncrypt(ctr.block, data)
}

// EncryptWithIV used to encrypt data with given iv. Output is cipher data.
func (ctr *CTR) EncryptWithIV(data, iv []byte) ([]byte, error) {
	return ctrEncryptWithIV(ctr.block, data, iv)
}

// Decrypt is used to decrypt cipher data. Input data is [IV + cipher data].
func (ctr *CTR) Decrypt(data []byte) ([]byte, error) {
	return ctrDecrypt(ctr.block, data)
}

// DecryptWithIV is used to decrypt cipher data with given iv. Input data is cipher data.
func (ctr *CTR) DecryptWithIV(data, iv []byte) ([]byte, error) {
	return ctrDecryptWithIV(ctr.block, data, iv)
}

// Key is used to get AES Key.
func (ctr *CTR) Key() []byte {
	key := ctr.key.Get()
	defer ctr.key.Put(key)
	// copy it, usually cover it after use.
	cp := make([]byte, len(key))
	copy(cp, key)
	return cp
}
