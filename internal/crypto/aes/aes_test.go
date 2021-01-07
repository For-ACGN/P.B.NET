package aes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func generateBytes() []byte {
	testdata := make([]byte, 63)
	for i := 0; i < 63; i++ {
		testdata[i] = byte(i)
	}
	return testdata
}

func TestCTR(t *testing.T) {
	key128 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	key192 := append(key128, []byte{17, 18, 19, 20, 21, 22, 23, 24}...)
	key256 := bytes.Repeat(key128, 2)

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
	}

	t.Run("key 128bit", func(t *testing.T) {
		testFn(key128)
	})

	t.Run("key 192bit", func(t *testing.T) {
		testFn(key192)
	})

	t.Run("key 256bit", func(t *testing.T) {
		testFn(key256)
	})
}

func TestAES(t *testing.T) {
	key128 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	key256 := bytes.Repeat(key128, 2)
	iv := []byte{11, 12, 13, 14, 15, 16, 17, 18, 19, 10, 111, 112, 113, 114, 115, 116}

	// encrypt & decrypt
	f := func(key []byte) {
		testdata := generateBytes()

		cipherData, err := CBCEncrypt(testdata, key, iv)
		require.NoError(t, err)
		require.Equal(t, generateBytes(), testdata)
		require.NotEqual(t, testdata, cipherData)

		plainData, err := CBCDecrypt(cipherData, key, iv)
		require.NoError(t, err)
		require.Equal(t, testdata, plainData)
	}

	t.Run("key 128bit", func(t *testing.T) {
		f(key128)
	})

	t.Run("key 256bit", func(t *testing.T) {
		f(key256)
	})

	t.Run("no data", func(t *testing.T) {
		_, err := CBCEncrypt(nil, key128, iv)
		require.Equal(t, ErrEmptyData, err)

		_, err = CBCDecrypt(nil, key128, iv)
		require.Equal(t, ErrEmptyData, err)
	})

	data := bytes.Repeat([]byte{255}, 32)

	t.Run("invalid key", func(t *testing.T) {
		_, err := CBCEncrypt(data, nil, iv)
		require.Error(t, err)

		_, err = CBCDecrypt(data, nil, iv)
		require.Error(t, err)
	})

	t.Run("invalid iv", func(t *testing.T) {
		_, err := CBCEncrypt(data, key128, nil)
		require.Equal(t, ErrInvalidIVSize, err)

		_, err = CBCDecrypt(data, key128, nil)
		require.Equal(t, ErrInvalidIVSize, err)
	})

	t.Run("ErrInvalidCipherData", func(t *testing.T) {
		_, err := CBCDecrypt(bytes.Repeat([]byte{0}, 13), key128, iv)
		require.Equal(t, ErrInvalidCipherData, err)

		_, err = CBCDecrypt(bytes.Repeat([]byte{0}, 63), key128, iv)
		require.Equal(t, ErrInvalidCipherData, err)
	})

	t.Run("ErrInvalidPaddingSize", func(t *testing.T) {
		_, err := CBCDecrypt(bytes.Repeat([]byte{0}, 64), key128, iv)
		require.Equal(t, ErrInvalidPaddingSize, err)
	})
}

func TestCBC(t *testing.T) {
	key128 := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	key256 := bytes.Repeat(key128, 2)
	iv := []byte{11, 12, 13, 14, 15, 16, 17, 18, 19, 10, 111, 112, 113, 114, 115, 116}

	t.Run("test key", func(t *testing.T) {
		_, err := NewCBC(bytes.Repeat([]byte{0}, Key128Bit), iv)
		require.NoError(t, err)

		_, err = NewCBC(bytes.Repeat([]byte{0}, Key192Bit), iv)
		require.NoError(t, err)

		_, err = NewCBC(bytes.Repeat([]byte{0}, Key256Bit), iv)
		require.NoError(t, err)
	})

	// encrypt & decrypt
	f := func(key []byte) {
		cbc, err := NewCBC(key, iv)
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
	}

	t.Run("key 128bit", func(t *testing.T) {
		f(key128)
	})

	t.Run("key 256bit", func(t *testing.T) {
		f(key256)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, err := NewCBC(nil, iv)
		require.Error(t, err)
	})

	t.Run("invalid iv", func(t *testing.T) {
		_, err := NewCBC(key128, nil)
		require.Error(t, err)
	})

	t.Run("no data", func(t *testing.T) {
		cbc, err := NewCBC(key128, iv)
		require.NoError(t, err)

		_, err = cbc.Encrypt(nil)
		require.Equal(t, ErrEmptyData, err)

		_, err = cbc.Decrypt(nil)
		require.Equal(t, ErrEmptyData, err)
	})

	t.Run("invalid data", func(t *testing.T) {
		cbc, err := NewCBC(key128, iv)
		require.NoError(t, err)

		_, err = cbc.Decrypt(bytes.Repeat([]byte{0}, 13))
		require.Equal(t, ErrInvalidCipherData, err)

		_, err = cbc.Decrypt(bytes.Repeat([]byte{0}, 63))
		require.Equal(t, ErrInvalidCipherData, err)

		_, err = cbc.Decrypt(bytes.Repeat([]byte{0}, 64))
		require.Equal(t, ErrInvalidPaddingSize, err)
	})

	t.Run("key iv", func(t *testing.T) {
		cbc, err := NewCBC(key128, iv)
		require.NoError(t, err)

		k, v := cbc.KeyIV()
		require.Equal(t, key128, k)
		require.Equal(t, iv, v)
	})
}

func TestCBC_Parallel(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	iv := []byte{11, 12, 13, 14, 15, 16, 17, 18, 19, 10, 111, 112, 113, 114, 115, 116}

	data := bytes.Repeat([]byte{1, 2, 3, 4}, 10)

	t.Run("part", func(t *testing.T) {
		cbc, err := NewCBC(key, iv)
		require.NoError(t, err)

		enc := func() {
			_, err := cbc.Encrypt(data)
			require.NoError(t, err)
		}
		dec := func() {
			cipherData, err := cbc.Encrypt(data)
			require.NoError(t, err)
			plainData, err := cbc.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, data, plainData)
		}
		keyIV := func() {
			k, i := cbc.KeyIV()
			require.Equal(t, key, k)
			require.Equal(t, iv, i)
		}
		testsuite.RunParallel(100, nil, nil, enc, dec, keyIV)
	})

	t.Run("whole", func(t *testing.T) {
		var cbc *CBC

		init := func() {
			var err error
			cbc, err = NewCBC(key, iv)
			require.NoError(t, err)
		}
		enc := func() {
			_, err := cbc.Encrypt(data)
			require.NoError(t, err)
		}
		dec := func() {
			cipherData, err := cbc.Encrypt(data)
			require.NoError(t, err)
			plainData, err := cbc.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, data, plainData)
		}
		keyIV := func() {
			k, i := cbc.KeyIV()
			require.Equal(t, key, k)
			require.Equal(t, iv, i)
		}
		testsuite.RunParallel(100, init, nil, enc, dec, keyIV)
	})

	t.Run("multi", func(t *testing.T) {
		cbc, err := NewCBC(key, iv)
		require.NoError(t, err)

		testsuite.RunMultiTimes(100, func() {
			cipherData, err := cbc.Encrypt(data)
			require.NoError(t, err)
			plainData, err := cbc.Decrypt(cipherData)
			require.NoError(t, err)
			require.Equal(t, data, plainData)

			k, i := cbc.KeyIV()
			require.Equal(t, key, k)
			require.Equal(t, iv, i)
		})

		testsuite.IsDestroyed(t, cbc)
	})
}

func BenchmarkCBC_Encrypt(b *testing.B) {
	data := bytes.Repeat([]byte{0}, 64)

	b.Run("128bit", func(b *testing.B) {
		key := bytes.Repeat([]byte{0}, 16)

		benchmarkCBCEncrypt(b, data, key)
	})

	b.Run("256bit", func(b *testing.B) {
		key := bytes.Repeat([]byte{0}, 32)

		benchmarkCBCEncrypt(b, data, key)
	})
}

func benchmarkCBCEncrypt(b *testing.B, data, key []byte) {
	iv := bytes.Repeat([]byte{0}, IVSize)
	cbc, err := NewCBC(key, iv)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cbc.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkCBC_Decrypt(b *testing.B) {
	iv := bytes.Repeat([]byte{0}, IVSize)

	b.Run("128bit", func(b *testing.B) {
		key := bytes.Repeat([]byte{0}, 16)
		cipherData, err := CBCEncrypt(bytes.Repeat([]byte{0}, 64), key, iv)
		require.NoError(b, err)

		benchmarkCBCDecrypt(b, cipherData, key, iv)
	})

	b.Run("256bit", func(b *testing.B) {
		key := bytes.Repeat([]byte{0}, 32)
		cipherData, err := CBCEncrypt(bytes.Repeat([]byte{0}, 64), key, iv)
		require.NoError(b, err)

		benchmarkCBCDecrypt(b, cipherData, key, iv)
	})
}

func benchmarkCBCDecrypt(b *testing.B, data, key, iv []byte) {
	cbc, err := NewCBC(key, iv)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cbc.Decrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}
