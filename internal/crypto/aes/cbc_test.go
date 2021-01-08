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

func TestCBC(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) {
		testCBC(t, test128BitKey)
	})

	t.Run("192 bit key", func(t *testing.T) {
		testCBC(t, test192BitKey)
	})

	t.Run("256 bit key", func(t *testing.T) {
		testCBC(t, test256BitKey)
	})
}

func testCBC(t *testing.T, key []byte) {
	cbc, err := NewCBC(key)
	require.NoError(t, err)

	testdata := generateBytes()

	for i := 0; i < 10; i++ {
		cipherData, err := cbc.Encrypt(testdata)
		require.NoError(t, err)

		require.Equal(t, generateBytes(), testdata)
		require.NotEqual(t, testdata, cipherData)
	}

	cipherData, err := cbc.Encrypt(testdata)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		plainData, err := cbc.Decrypt(cipherData)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	}

	require.Equal(t, key, cbc.Key())
}
