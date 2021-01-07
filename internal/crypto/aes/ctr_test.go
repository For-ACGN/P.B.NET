package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAESCTR(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) {
		testAESCTR(t, test128BitKey)
	})

	t.Run("192 bit key", func(t *testing.T) {
		testAESCTR(t, test192BitKey)
	})

	t.Run("256 bit key ", func(t *testing.T) {
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

func TestCTR(t *testing.T) {
	// encrypt & decrypt
	testFn := func(key []byte) {
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

	t.Run("key 128bit", func(t *testing.T) {
		testFn(test128BitKey)
	})

	t.Run("key 192bit", func(t *testing.T) {
		testFn(test192BitKey)
	})

	t.Run("key 256bit", func(t *testing.T) {
		testFn(test256BitKey)
	})
}
