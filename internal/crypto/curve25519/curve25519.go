package curve25519

import (
	"golang.org/x/crypto/curve25519"
)

// ScalarSize is the size of the scalar input to X25519.
const ScalarSize = 32

// Basepoint is the canonical Curve25519 generator.
var Basepoint []byte

var basePoint = [32]byte{
	9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

func init() {
	Basepoint = basePoint[:]
}

// X25519 returns the result of the scalar multiplication (scalar * point),
// according to RFC 7748, Section 5. scalar, point and the return value are
// slices of 32 bytes.
//
// scalar can be generated at random, for example with crypto/rand. point should
// be either Basepoint or the output of another X25519 call.
//
// If point is Basepoint (but not if it's a different slice with the same
// contents) a precomputed implementation might be used for performance.
func X25519(in, base []byte) ([]byte, error) {
	return curve25519.X25519(in, base)
}

// X25519Base use Basepoint as base argument.
func X25519Base(in []byte) ([]byte, error) {
	return X25519(in, Basepoint)
}
