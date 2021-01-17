package aes

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/convert"
	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
	"project/internal/random"
)

func TestCTR(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) { testCTR(t, test128BitKey) })
	t.Run("192 bit key", func(t *testing.T) { testCTR(t, test192BitKey) })
	t.Run("256 bit key", func(t *testing.T) { testCTR(t, test256BitKey) })
}

func testCTR(t *testing.T, key []byte) {
	t.Run("without iv", func(t *testing.T) {
		testdata := testGenerateBytes()
		testdataCp := append([]byte{}, testdata...)

		cipherData, err := CTREncrypt(testdata, key)
		require.NoError(t, err)

		require.Equal(t, testdataCp, testdata)
		require.NotEqual(t, testdata, cipherData)

		plainData, err := CTRDecrypt(cipherData, key)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	})

	t.Run("with iv", func(t *testing.T) {
		testdata := testGenerateBytes()
		testdataCp := append([]byte{}, testdata...)
		iv, err := GenerateIV()
		require.NoError(t, err)

		cipherData, err := CTREncryptWithIV(testdata, key, iv)
		require.NoError(t, err)

		require.Equal(t, testdataCp, testdata)
		require.NotEqual(t, testdata, cipherData)

		plainData, err := CTRDecryptWithIV(cipherData, key, iv)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	})
}

func TestCTREncrypt(t *testing.T) {
	testdata := make([]byte, 64)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CTREncrypt(testdata, nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CTREncrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("failed to generate iv", func(t *testing.T) {
		patch := func([]byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(rand.Read, patch)
		defer pg.Unpatch()

		_, err := CTREncrypt(testdata, test128BitKey)
		monkey.IsExistMonkeyError(t, err)
	})
}

func TestCTREncryptWithIV(t *testing.T) {
	iv, err := GenerateIV()
	require.NoError(t, err)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CTREncryptWithIV(nil, nil, iv)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CTREncryptWithIV(nil, test128BitKey, iv)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid iv size", func(t *testing.T) {
		_, err := CTREncryptWithIV(make([]byte, 64), test128BitKey, make([]byte, 8))
		require.Equal(t, ErrInvalidIVSize, err)
	})
}

func TestCTRDecrypt(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		_, err := CTRDecrypt(make([]byte, 64), nil)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CTRDecrypt(nil, test128BitKey)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid cipher data", func(t *testing.T) {
		_, err := CTRDecrypt(make([]byte, 7), test128BitKey)
		require.Equal(t, ErrInvalidCipherData, err)
	})
}

func TestCTRDecryptWithIV(t *testing.T) {
	iv, err := GenerateIV()
	require.NoError(t, err)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CTRDecryptWithIV(nil, nil, iv)
		require.Error(t, err)
		_, ok := err.(aes.KeySizeError)
		require.True(t, ok)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := CTRDecryptWithIV(nil, test128BitKey, iv)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid iv size", func(t *testing.T) {
		_, err := CTRDecryptWithIV(make([]byte, 64), test128BitKey, make([]byte, 8))
		require.Equal(t, ErrInvalidIVSize, err)
	})
}

func TestNewCTR(t *testing.T) {
	ctr, err := NewCTR(nil)
	require.Error(t, err)
	require.Nil(t, ctr)
	_, ok := err.(aes.KeySizeError)
	require.True(t, ok)
}

func TestCTR_SetStream(t *testing.T) {
	ctr, err := NewCTR(test256BitKey)
	require.NoError(t, err)
	err = ctr.SetStream(nil)
	require.Equal(t, ErrInvalidIVSize, err)
}

func TestCTR_XORKeyStream(t *testing.T) {
	testdata1 := testGenerateBytes()
	testdata2 := testGenerateBytes()
	testdata1Len := len(testdata1)
	testdata2Len := len(testdata2)
	testdata := convert.MergeBytes(testdata1, testdata2)
	iv, err := GenerateIV()
	require.NoError(t, err)

	ctr1, err := NewCTR(test256BitKey)
	require.NoError(t, err)
	err = ctr1.SetStream(iv)
	require.NoError(t, err)

	ctr2, err := NewCTR(test256BitKey)
	require.NoError(t, err)
	err = ctr2.SetStream(iv)
	require.NoError(t, err)

	cipherData1 := make([]byte, testdata1Len)
	ctr1.XORKeyStream(cipherData1, testdata1)
	cipherData2 := make([]byte, testdata2Len)
	ctr1.XORKeyStream(cipherData2, testdata2)
	cipherData := convert.MergeBytes(cipherData1, cipherData2)

	rv := 64 + random.Intn(64)
	plainData1 := make([]byte, testdata1Len-rv)
	ctr2.XORKeyStream(plainData1, cipherData[:testdata1Len-rv])
	plainData2 := make([]byte, testdata2Len+rv)
	ctr2.XORKeyStream(plainData2, cipherData[testdata1Len-rv:])
	plainData := convert.MergeBytes(plainData1, plainData2)

	require.Equal(t, testdata, plainData)
}

func BenchmarkCTR_Encrypt(b *testing.B) {
	b.Run("64 Bytes ", func(b *testing.B) { benchmarkCTREncrypt(b, 64) })
	b.Run("256 Bytes", func(b *testing.B) { benchmarkCTREncrypt(b, 256) })
	b.Run("1 KB     ", func(b *testing.B) { benchmarkCTREncrypt(b, 1024) })
	b.Run("4 KB     ", func(b *testing.B) { benchmarkCTREncrypt(b, 4*1024) })
	b.Run("16 KB    ", func(b *testing.B) { benchmarkCTREncrypt(b, 16*1024) })
	b.Run("128 KB   ", func(b *testing.B) { benchmarkCTREncrypt(b, 128*1024) })
	b.Run("1 MB     ", func(b *testing.B) { benchmarkCTREncrypt(b, 1024*1024) })
}

func benchmarkCTREncrypt(b *testing.B, size int) {
	data := bytes.Repeat([]byte{1}, size)

	ctr, err := NewCTR(test256BitKey)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = ctr.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkCTR_Decrypt(b *testing.B) {
	b.Run("64 Bytes ", func(b *testing.B) { benchmarkCTRDecrypt(b, 64) })
	b.Run("256 Bytes", func(b *testing.B) { benchmarkCTRDecrypt(b, 256) })
	b.Run("1 KB     ", func(b *testing.B) { benchmarkCTRDecrypt(b, 1024) })
	b.Run("4 KB     ", func(b *testing.B) { benchmarkCTRDecrypt(b, 4*1024) })
	b.Run("16 KB    ", func(b *testing.B) { benchmarkCTRDecrypt(b, 16*1024) })
	b.Run("128 KB   ", func(b *testing.B) { benchmarkCTRDecrypt(b, 128*1024) })
	b.Run("1 MB     ", func(b *testing.B) { benchmarkCTRDecrypt(b, 1024*1024) })
}

func benchmarkCTRDecrypt(b *testing.B, size int) {
	data := bytes.Repeat([]byte{1}, size)

	ctr, err := NewCTR(test256BitKey)
	require.NoError(b, err)

	cipherData, err := ctr.Encrypt(data)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = ctr.Decrypt(cipherData)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}
