package aes

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
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

func TestCBCEncrypt(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		_, err := CBCEncrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	testdata := make([]byte, 64)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCEncrypt(testdata, nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("failed to generate random iv", func(t *testing.T) {
		patch := func([]byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(rand.Read, patch)
		defer pg.Unpatch()

		_, err := CBCEncrypt(testdata, test128BitKey)
		monkey.IsExistMonkeyError(t, err)
	})
}

func TestCBCDecrypt(t *testing.T) {
	t.Run("invalid cipher data", func(t *testing.T) {
		_, err := CBCDecrypt(nil, test128BitKey)
		require.Equal(t, ErrInvalidCipherData, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCDecrypt(make([]byte, 64), nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})
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
