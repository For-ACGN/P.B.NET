package aes

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
)

func TestAESCTR(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) {
		testAESCTR(t, test128BitKey)
	})

	t.Run("192 bit key", func(t *testing.T) {
		testAESCTR(t, test192BitKey)
	})

	t.Run("256 bit key", func(t *testing.T) {
		testAESCTR(t, test256BitKey)
	})
}

func testAESCTR(t *testing.T, key []byte) {
	testdata := generateBytes()

	cipherData, err := CTREncrypt(testdata, key)
	require.NoError(t, err)
	require.Equal(t, generateBytes(), testdata)
	require.NotEqual(t, testdata, cipherData)

	plainData, err := CTRDecrypt(cipherData, key)
	require.NoError(t, err)
	require.Equal(t, testdata, plainData)
}

func TestCTREncrypt(t *testing.T) {
	testdata := make([]byte, 64)

	t.Run("empty data", func(t *testing.T) {
		_, err := CTREncrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, err := CTREncrypt(testdata, nil)
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

		_, err := CTREncrypt(testdata, test128BitKey)
		monkey.IsExistMonkeyError(t, err)
	})
}

func TestCTRDecrypt(t *testing.T) {
	t.Run("invalid cipher data", func(t *testing.T) {
		_, err := CTRDecrypt(nil, test128BitKey)
		require.Equal(t, ErrInvalidCipherData, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, err := CTRDecrypt(make([]byte, 64), nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})
}

func TestCTR(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) {
		testCTR(t, test128BitKey)
	})

	t.Run("192 bit key", func(t *testing.T) {
		testCTR(t, test192BitKey)
	})

	t.Run("256 bit key", func(t *testing.T) {
		testCTR(t, test256BitKey)
	})
}

func testCTR(t *testing.T, key []byte) {
	ctr, err := NewCTR(key)
	require.NoError(t, err)

	testdata := generateBytes()

	for i := 0; i < 10; i++ {
		cipherData, err := ctr.Encrypt(testdata)
		require.NoError(t, err)

		require.Equal(t, generateBytes(), testdata)
		require.NotEqual(t, testdata, cipherData)
	}

	cipherData, err := ctr.Encrypt(testdata)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		plainData, err := ctr.Decrypt(cipherData)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	}

	require.Equal(t, key, ctr.Key())
}
