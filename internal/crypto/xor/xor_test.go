package xor

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestXOR(t *testing.T) {
	plainData := testsuite.Bytes()
	key := []byte{1, 2, 3, 4}

	cipherData := XOR(plainData, key)
	require.NotEqual(t, plainData, cipherData)

	data := XOR(cipherData, key)
	require.Equal(t, plainData, data)
}

func TestBuffer(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		XOR(nil, nil)
	})

	t.Run("invalid data size", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		Buffer(make([]byte, 5), nil, nil)
	})

	t.Run("invalid key size", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		Buffer(make([]byte, 5), make([]byte, 5), nil)
	})
}
