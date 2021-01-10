package ed25519

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"project/internal/crypto/rand"
)

const (
	// PublicKeySize is the size, in bytes, of public keys as used in this package.
	PublicKeySize = 32

	// PrivateKeySize is the size, in bytes, of private keys as used in this package.
	PrivateKeySize = 64

	// SignatureSize is the size, in bytes, of signatures generated and verified by this package.
	SignatureSize = 64

	// SeedSize is the size, in bytes, of private key seeds. used by RFC 8032.
	SeedSize = 32
)

// Errors about ImportPrivateKey and ImportPublicKey.
var (
	ErrInvalidPrivateKeySize = errors.New("invalid ed25519 private key size")
	ErrInvalidPublicKeySize  = errors.New("invalid ed25519 public key size")
)

// GenerateKey is used to generate private key.
func GenerateKey() (ed25519.PrivateKey, error) {
	seed := make([]byte, SeedSize)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed: %s", err)
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

// GetPublicKey is used to get public key from private key.
func GetPublicKey(key ed25519.PrivateKey) ed25519.PublicKey {
	return key.Public().(ed25519.PublicKey)
}

// ImportPrivateKey is used to import private key from bytes.
func ImportPrivateKey(key []byte) (ed25519.PrivateKey, error) {
	if len(key) != PrivateKeySize {
		return nil, ErrInvalidPrivateKeySize
	}
	pri := make([]byte, PrivateKeySize)
	copy(pri, key)
	return pri, nil
}

// ImportPublicKey is used to import public key from bytes.
func ImportPublicKey(key []byte) (ed25519.PublicKey, error) {
	if len(key) != PublicKeySize {
		return nil, ErrInvalidPublicKeySize
	}
	pub := make([]byte, PublicKeySize)
	copy(pub, key)
	return pub, nil
}

// Sign signs the message with private key and returns a signature
// It will panic if len(privateKey) is not PrivateKeySize.
func Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// Verify reports whether signature is a valid signature of message by
// public key. It will panic if len(publicKey) is not PublicKeySize.
func Verify(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}
