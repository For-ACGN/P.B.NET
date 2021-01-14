package aes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
	"project/internal/random"
	"project/internal/testsuite"
)

var (
	test128BitKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	test192BitKey = append(test128BitKey, []byte{17, 18, 19, 20, 21, 22, 23, 24}...)
	test256BitKey = bytes.Repeat(test128BitKey, 2)
)

var tests = [...]*struct {
	name   string
	newAES func(key []byte) (AES, error)
}{
	{"CBC", NewCBC},
	{"CTR", NewCTR},
}

func generateBytes() []byte {
	return random.Bytes(512 + random.Int(1024))
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aes, err := test.newAES(key)
			require.NoError(t, err)

			t.Run("without iv", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					testdata := generateBytes()
					testdataCp := append([]byte{}, testdata...)

					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)

					require.Equal(t, testdataCp, testdata)
					require.NotEqual(t, testdata, cipherData)
				}

				for i := 0; i < 20; i++ {
					testdata := generateBytes()

					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)

					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					plainData, err = aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
			})

			t.Run("with iv", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					iv, err := GenerateIV()
					require.NoError(t, err)

					testdata := generateBytes()
					testdataCp := append([]byte{}, testdata...)

					cipherData, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)

					require.Equal(t, testdataCp, testdata)
					require.NotEqual(t, testdata, cipherData)
				}

				for i := 0; i < 20; i++ {
					iv, err := GenerateIV()
					require.NoError(t, err)

					testdata := generateBytes()

					cipherData, err := aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)

					plainData, err := aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					plainData, err = aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)
				}
			})

			t.Run("one way", func(t *testing.T) {
				testdata1 := generateBytes()
				testdata2 := generateBytes()
				iv, err := GenerateIV()
				require.NoError(t, err)

				aes1, err := test.newAES(key)
				require.NoError(t, err)
				aes2, err := test.newAES(key)
				require.NoError(t, err)

				cipherData1, err := aes1.Encrypt(testdata1)
				require.NoError(t, err)
				cipherData2, err := aes1.Encrypt(testdata2)
				require.NoError(t, err)

				plainData1, err := aes2.Decrypt(cipherData1)
				require.NoError(t, err)
				require.Equal(t, testdata1, plainData1)
				plainData2, err := aes2.Decrypt(cipherData2)
				require.NoError(t, err)
				require.Equal(t, testdata2, plainData2)

				plainData2, err = aes2.Decrypt(cipherData2)
				require.NoError(t, err)
				require.Equal(t, testdata2, plainData2)
				plainData1, err = aes2.Decrypt(cipherData1)
				require.NoError(t, err)
				require.Equal(t, testdata1, plainData1)

				cipherData1, err = aes1.EncryptWithIV(testdata1, iv)
				require.NoError(t, err)
				cipherData2, err = aes1.EncryptWithIV(testdata2, iv)
				require.NoError(t, err)

				plainData1, err = aes2.DecryptWithIV(cipherData1, iv)
				require.NoError(t, err)
				require.Equal(t, testdata1, plainData1)
				plainData2, err = aes2.DecryptWithIV(cipherData2, iv)
				require.NoError(t, err)
				require.Equal(t, testdata2, plainData2)

				plainData2, err = aes2.DecryptWithIV(cipherData2, iv)
				require.NoError(t, err)
				require.Equal(t, testdata2, plainData2)
				plainData1, err = aes2.DecryptWithIV(cipherData1, iv)
				require.NoError(t, err)
				require.Equal(t, testdata1, plainData1)
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
				aes, err := test.newAES(test128BitKey)
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

					plainData, err = aes.Decrypt(cipherData)
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

					plainData, err = aes.DecryptWithIV(cipherData, iv)
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
					aes, err = test.newAES(test128BitKey)
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

					plainData, err = aes.Decrypt(cipherData)
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

					plainData, err = aes.DecryptWithIV(cipherData, iv)
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
				aes, err := test.newAES(test128BitKey)
				require.NoError(t, err)

				testsuite.RunMultiTimes(100, func() {
					cipherData, err := aes.Encrypt(testdata)
					require.NoError(t, err)

					plainData, err := aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					plainData, err = aes.Decrypt(cipherData)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

					cipherData, err = aes.EncryptWithIV(testdata, iv)
					require.NoError(t, err)

					plainData, err = aes.DecryptWithIV(cipherData, iv)
					require.NoError(t, err)
					require.Equal(t, testdata, plainData)

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
