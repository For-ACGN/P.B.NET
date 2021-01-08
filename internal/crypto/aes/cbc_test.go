package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAESCBC(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) {
		testAESCBC(t, test128BitKey)
	})

	t.Run("192 bit key", func(t *testing.T) {
		testAESCBC(t, test192BitKey)
	})

	t.Run("256 bit key", func(t *testing.T) {
		testAESCBC(t, test256BitKey)
	})
}

func testAESCBC(t *testing.T, key []byte) {
	testdata := generateBytes()

	cipherData, err := CBCEncrypt(testdata, key)
	require.NoError(t, err)
	require.Equal(t, generateBytes(), testdata)
	require.NotEqual(t, testdata, cipherData)

	plainData, err := CBCDecrypt(cipherData, key)
	require.NoError(t, err)
	require.Equal(t, testdata, plainData)
}
