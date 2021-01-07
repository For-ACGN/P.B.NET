package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"sync"

	"project/internal/crypto/rand"
	"project/internal/security"
)

// ctrIVPool is used to cache iv byte slice for CTREncrypt and CTRDecrypt.
var ctrIVPool sync.Pool

func init() {
	ctrIVPool.New = func() interface{} {
		iv := make([]byte, IVSize)
		return &iv
	}
}

// CTREncrypt is used to encrypt plain data with counter mode. Output is [IV + cipher data].
func CTREncrypt(data, key []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// generate random iv
	ivPtr := ctrIVPool.Get().(*[]byte)
	defer ctrIVPool.Put(ivPtr)
	iv := *ivPtr
	_, err = rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random iv: %s", err)
	}
	// make buffer
	output := make([]byte, IVSize+l)
	copy(output, iv)
	// encrypt plain data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output[IVSize:], data)
	return output, nil
}

// CTRDecrypt is used to decrypt cipher data with counter mode. Input data is [IV + cipher data].
func CTRDecrypt(data, key []byte) ([]byte, error) {
	l := len(data)
	if l < IVSize+1 {
		return nil, ErrInvalidCipherData
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := data[:IVSize]
	output := make([]byte, l-IVSize)
	// decrypt cipher data
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(output, data[IVSize:])
	return output, nil
}

// CTR is the AES encryptor with counter mode.
type CTR struct {
	key    *security.Bytes
	block  cipher.Block
	ivPool sync.Pool
}

// NewCTR is used to create a AES encryptor with counter mode.
func NewCTR(key []byte) (*CTR, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ctr := CTR{
		key:   security.NewBytes(key),
		block: block,
	}
	ctr.ivPool.New = func() interface{} {
		iv := make([]byte, IVSize)
		return &iv
	}
	return &ctr, nil
}

// Encrypt is used to encrypt plain data. Output is [IV + cipher data].
func (ctr *CTR) Encrypt(data []byte) ([]byte, error) {
	l := len(data)
	if l == 0 {
		return nil, ErrEmptyData
	}
	// generate random iv
	ivPtr := ctr.ivPool.Get().(*[]byte)
	defer ctr.ivPool.Put(ivPtr)
	iv := *ivPtr
	_, err := rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random iv: %s", err)
	}
	// make buffer
	output := make([]byte, IVSize+l)
	copy(output, iv)
	// encrypt plain data
	stream := cipher.NewCTR(ctr.block, iv)
	stream.XORKeyStream(output[IVSize:], data)
	return output, nil
}

// Decrypt is used to decrypt cipher data. Input data is [IV + cipher data].
func (ctr *CTR) Decrypt(data []byte) ([]byte, error) {
	l := len(data)
	if l < IVSize+1 {
		return nil, ErrInvalidCipherData
	}
	iv := data[:IVSize]
	output := make([]byte, l-IVSize)
	// decrypt cipher data
	stream := cipher.NewCTR(ctr.block, iv)
	stream.XORKeyStream(output, data[IVSize:])
	return output, nil
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
