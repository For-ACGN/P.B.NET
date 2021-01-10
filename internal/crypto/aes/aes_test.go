package aes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

var (
	test128BitKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	test192BitKey = append(test128BitKey, []byte{17, 18, 19, 20, 21, 22, 23, 24}...)
	test256BitKey = bytes.Repeat(test128BitKey, 2)
)

var tests = [...]*struct {
	name string
	fn   func(key []byte) (AES, error)
}{
	{"CBC", NewCBC},
	{"CTR", NewCTR},
}

func generateBytes() []byte {
	testdata := make([]byte, 63)
	for i := 0; i < 63; i++ {
		testdata[i] = byte(i)
	}
	return testdata
}

func TestGenerateIV(t *testing.T) {
	patch := func([]byte) (int, error) {
		return 0, monkey.Error
	}
	pg := monkey.Patch(rand.Read, patch)
	defer pg.Unpatch()

	iv, err := GenerateIV()
	monkey.IsExistMonkeyError(t, err)
	require.Nil(t, iv)
}

func TestAES(t *testing.T) {
	t.Run("128 bit key", func(t *testing.T) { testAES(t, test128BitKey) })
	t.Run("192 bit key", func(t *testing.T) { testAES(t, test192BitKey) })
	t.Run("256 bit key", func(t *testing.T) { testAES(t, test256BitKey) })
}

func testAES(t *testing.T, key []byte) {
	testdata := generateBytes()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aes, err := test.fn(key)
			require.NoError(t, err)

			t.Run("without iv", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)

					require.Equal(t, generateBytes(), testdata)
					require.NotEqual(t, testdata, cipherData)
				}

				cipherData, err := aes.Encrypt(testdata)
				require.NoError(t, err)
				for i := 0; i < 20; i++ {
					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
			})

			t.Run("with iv", func(t *testing.T) {
				iv, err := GenerateIV()
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					cipherData, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)

					require.Equal(t, generateBytes(), testdata)
					require.NotEqual(t, testdata, cipherData)
				}

				cipherData, err := aes.EncryptWithIV(testdata, iv)
				require.NoError(t, err)
				for i := 0; i < 20; i++ {
					plainData, err := aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
			})

			require.Equal(t, key, aes.Key())

			testsuite.IsDestroyed(t, aes)
		})
	}
}

func TestAES_Parallel(t *testing.T) {
	testdata := generateBytes()
	iv, err := GenerateIV()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("part", func(t *testing.T) {
				aes, err := test.fn(test128BitKey)
				require.NoError(t, err)

				enc := func() {
					_, err := aes.Encrypt(testdata)
					require.NoError(t, err)
				}
				dec := func() {
					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)
					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
				encWithIV := func() {
					_, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)
				}
				decWithIV := func() {
					cipherData, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)
					plainData, err := aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
				key := func() {
					key := aes.Key()
					require.Equal(t, test128BitKey, key)
				}
				testsuite.RunParallel(100, nil, nil, enc, dec, encWithIV, decWithIV, key)

				testsuite.IsDestroyed(t, aes)
			})

			t.Run("whole", func(t *testing.T) {
				var aes AES

				init := func() {
					var err error
					aes, err = test.fn(test128BitKey)
					require.NoError(t, err)
				}
				enc := func() {
					_, err := aes.Encrypt(testdata)
					require.NoError(t, err)
				}
				dec := func() {
					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)
					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
				encWithIV := func() {
					_, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)
				}
				decWithIV := func() {
					cipherData, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)
					plainData, err := aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
				key := func() {
					key := aes.Key()
					require.Equal(t, test128BitKey, key)
				}
				testsuite.RunParallel(100, init, nil, enc, dec, encWithIV, decWithIV, key)

				testsuite.IsDestroyed(t, aes)
			})

			t.Run("multi", func(t *testing.T) {
				aes, err := test.fn(test128BitKey)
				require.NoError(t, err)

				testsuite.RunMultiTimes(100, func() {
					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)
					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					cipherData, err = aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)
					plainData, err = aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					key := aes.Key()
					require.Equal(t, test128BitKey, key)
				})

				testsuite.IsDestroyed(t, aes)
			})
		})
	}
}
