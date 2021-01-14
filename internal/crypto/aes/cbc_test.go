package aes

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
)

func TestCBC(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) { testCBC(t, test128BitKey) })
	t.Run("192 bit key", func(t *testing.T) { testCBC(t, test192BitKey) })
	t.Run("256 bit key", func(t *testing.T) { testCBC(t, test256BitKey) })
}

func testCBC(t *testing.T, key []byte) {
	t.Run("without iv", func(t *testing.T) {
		testdata := testGenerateBytes()
		testdataCp := append([]byte{}, testdata...)

		cipherData, err := CBCEncrypt(testdata, key)
		require.NoError(t, err)

		require.Equal(t, testdataCp, testdata)
		require.NotEqual(t, testdata, cipherData)

		plainData, err := CBCDecrypt(cipherData, key)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	})

	t.Run("with iv", func(t *testing.T) {
		testdata := testGenerateBytes()
		testdataCp := append([]byte{}, testdata...)
		iv, err := GenerateIV()
		require.NoError(t, err)

		cipherData, err := CBCEncryptWithIV(testdata, key, iv)
		require.NoError(t, err)

		require.Equal(t, testdataCp, testdata)
		require.NotEqual(t, testdata, cipherData)

		plainData, err := CBCDecryptWithIV(cipherData, key, iv)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	})
}

func TestCBCEncrypt(t *testing.T) {
	testdata := make([]byte, 64)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCEncrypt(testdata, nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CBCEncrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("failed to generate iv", func(t *testing.T) {
		patch := func([]byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(rand.Read, patch)
		defer pg.Unpatch()

		_, err := CBCEncrypt(testdata, test128BitKey)
		monkey.IsExistMonkeyError(t, err)
	})
}

func TestCBCEncryptWithIV(t *testing.T) {
	iv, err := GenerateIV()
	require.NoError(t, err)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCEncryptWithIV(nil, nil, iv)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CBCEncryptWithIV(nil, test128BitKey, iv)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid iv size", func(t *testing.T) {
		_, err := CBCEncryptWithIV(make([]byte, 64), test128BitKey, make([]byte, 8))
		require.Equal(t, ErrInvalidIVSize, err)
	})
}

func TestCBCDecrypt(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCDecrypt(make([]byte, 64), nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CBCDecrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid cipher data", func(t *testing.T) {
		_, err := CBCDecrypt(make([]byte, 7), test128BitKey)
		require.Equal(t, ErrInvalidCipherData, err)

		_, err = CBCDecrypt(make([]byte, 63), test128BitKey)
		require.Equal(t, ErrInvalidCipherData, err)
	})

	t.Run("invalid padding size", func(t *testing.T) {
		_, err := CBCDecrypt(make([]byte, 64), test128BitKey)
		require.Equal(t, ErrInvalidPaddingSize, err)
	})
}

func TestCBCDecryptWithIV(t *testing.T) {
	iv, err := GenerateIV()
	require.NoError(t, err)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCDecryptWithIV(nil, nil, iv)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CBCDecryptWithIV(nil, test128BitKey, iv)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid cipher data", func(t *testing.T) {
		_, err := CBCDecryptWithIV(make([]byte, 7), test128BitKey, iv)
		require.Equal(t, ErrInvalidCipherData, err)

		_, err = CBCDecryptWithIV(make([]byte, 63), test128BitKey, iv)
		require.Equal(t, ErrInvalidCipherData, err)
	})

	t.Run("invalid iv size", func(t *testing.T) {
		_, err := CBCDecryptWithIV(make([]byte, 64), test128BitKey, make([]byte, 8))
		require.Equal(t, ErrInvalidIVSize, err)
	})

	t.Run("invalid padding size", func(t *testing.T) {
		_, err := CBCDecryptWithIV(make([]byte, 64), test128BitKey, iv)
		require.Equal(t, ErrInvalidPaddingSize, err)
	})
}

func TestNewCBC(t *testing.T) {
	cbc, err := NewCBC(nil)
	require.Error(t, err)
	require.Nil(t, cbc)
	_, ok := err.(aes.KeySizeError)
	require.True(t, ok)
}

func BenchmarkCBC_Encrypt(b *testing.B) {
	b.Run("64 Bytes ", func(b *testing.B) { benchmarkCBCEncrypt(b, 64) })
	b.Run("256 Bytes", func(b *testing.B) { benchmarkCBCEncrypt(b, 256) })
	b.Run("1 KB     ", func(b *testing.B) { benchmarkCBCEncrypt(b, 1024) })
	b.Run("4 KB     ", func(b *testing.B) { benchmarkCBCEncrypt(b, 4*1024) })
	b.Run("16 KB    ", func(b *testing.B) { benchmarkCBCEncrypt(b, 16*1024) })
	b.Run("128 KB   ", func(b *testing.B) { benchmarkCBCEncrypt(b, 128*1024) })
	b.Run("1 MB     ", func(b *testing.B) { benchmarkCBCEncrypt(b, 1024*1024) })
}

func benchmarkCBCEncrypt(b *testing.B, size int) {
	data := bytes.Repeat([]byte{1}, size)

	cbc, err := NewCBC(test256BitKey)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = cbc.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkCBC_Decrypt(b *testing.B) {
	b.Run("64 Bytes ", func(b *testing.B) { benchmarkCBCDecrypt(b, 64) })
	b.Run("256 Bytes", func(b *testing.B) { benchmarkCBCDecrypt(b, 256) })
	b.Run("1 KB     ", func(b *testing.B) { benchmarkCBCDecrypt(b, 1024) })
	b.Run("4 KB     ", func(b *testing.B) { benchmarkCBCDecrypt(b, 4*1024) })
	b.Run("16 KB    ", func(b *testing.B) { benchmarkCBCDecrypt(b, 16*1024) })
	b.Run("128 KB   ", func(b *testing.B) { benchmarkCBCDecrypt(b, 128*1024) })
	b.Run("1 MB     ", func(b *testing.B) { benchmarkCBCDecrypt(b, 1024*1024) })
}

func benchmarkCBCDecrypt(b *testing.B, size int) {
	data := bytes.Repeat([]byte{1}, size)

	cbc, err := NewCBC(test256BitKey)
	require.NoError(b, err)

	cipherData, err := cbc.Encrypt(data)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = cbc.Decrypt(cipherData)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}
