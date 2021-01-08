package aes

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
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

func TestNewCTR(t *testing.T) {
	ctr, err := NewCTR(nil)
	require.Error(t, err)
	require.Nil(t, ctr)
	_, ok := err.(aes.KeySizeError)
	require.True(t, ok)
}

func TestCTR_Encrypt(t *testing.T) {
	ctr, err := NewCTR(test128BitKey)
	require.NoError(t, err)

	t.Run("empty data", func(t *testing.T) {
		_, err := ctr.Encrypt(nil)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("failed to generate random iv", func(t *testing.T) {
		patch := func([]byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(rand.Read, patch)
		defer pg.Unpatch()

		_, err := ctr.Encrypt(make([]byte, 64))
		monkey.IsExistMonkeyError(t, err)
	})
}

func TestCTR_Decrypt(t *testing.T) {
	ctr, err := NewCTR(test128BitKey)
	require.NoError(t, err)

	_, err = ctr.Decrypt(nil)
	require.Equal(t, ErrInvalidCipherData, err)
}

func TestCTR_Parallel(t *testing.T) {
	testdata := generateBytes()

	t.Run("part", func(t *testing.T) {
		ctr, err := NewCTR(test128BitKey)
		require.NoError(t, err)

		enc := func() {
			_, err := ctr.Encrypt(testdata)
			require.NoError(t, err)
		}
		dec := func() {
			cipherData, err := ctr.Encrypt(testdata)
			require.NoError(t, err)
			plainData, err := ctr.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, testdata, plainData)
		}
		key := func() {
			key := ctr.Key()
			require.Equal(t, test128BitKey, key)
		}
		testsuite.RunParallel(100, nil, nil, enc, dec, key)

		testsuite.IsDestroyed(t, ctr)
	})

	t.Run("whole", func(t *testing.T) {
		var ctr *CTR

		init := func() {
			var err error
			ctr, err = NewCTR(test128BitKey)
			require.NoError(t, err)
		}
		enc := func() {
			_, err := ctr.Encrypt(testdata)
			require.NoError(t, err)
		}
		dec := func() {
			cipherData, err := ctr.Encrypt(testdata)
			require.NoError(t, err)
			plainData, err := ctr.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, testdata, plainData)
		}
		key := func() {
			key := ctr.Key()
			require.Equal(t, test128BitKey, key)
		}
		testsuite.RunParallel(100, init, nil, enc, dec, key)

		testsuite.IsDestroyed(t, ctr)
	})

	t.Run("multi", func(t *testing.T) {
		ctr, err := NewCTR(test128BitKey)
		require.NoError(t, err)

		testsuite.RunMultiTimes(100, func() {
			cipherData, err := ctr.Encrypt(testdata)
			require.NoError(t, err)
			plainData, err := ctr.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, testdata, plainData)

			key := ctr.Key()
			require.Equal(t, test128BitKey, key)
		})

		testsuite.IsDestroyed(t, ctr)
	})
}

func BenchmarkCTR_Encrypt(b *testing.B) {
	b.Run("64 Bytes", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 64)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("256 Bytes", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 256)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("1 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 1024)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("4 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 4*1024)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("16 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 16*1024)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("128 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 128*1024)
		benchmarkCTREncrypt(b, data)
	})

	b.Run("1 MB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 1024*1024)
		benchmarkCTREncrypt(b, data)
	})
}

func benchmarkCTREncrypt(b *testing.B, data []byte) {
	ctr, err := NewCTR(test256BitKey)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ctr.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkCTR_Decrypt(b *testing.B) {
	b.Run("64 Bytes", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 64)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("256 Bytes", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 256)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("1 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 1024)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("4 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 4*1024)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("16 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 16*1024)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("128 KB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 128*1024)
		benchmarkCTRDecrypt(b, data)
	})

	b.Run("1 MB", func(b *testing.B) {
		data := bytes.Repeat([]byte{0}, 1024*1024)
		benchmarkCTRDecrypt(b, data)
	})
}

func benchmarkCTRDecrypt(b *testing.B, data []byte) {
	ctr, err := NewCTR(test256BitKey)
	require.NoError(b, err)

	cipherData, err := ctr.Encrypt(data)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ctr.Decrypt(cipherData)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}
