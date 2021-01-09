package lsb

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/aes"
	"project/internal/patch/monkey"
	"project/internal/random"
	"project/internal/testsuite"
)

func TestCalculateStorageSize(t *testing.T) {
	for _, testdata := range [...]*struct {
		width  int
		height int
		output int
	}{
		{width: 100, height: 200, output: 19951},
		{width: 20, height: 3, output: 15},
		{width: 20, height: 2, output: 0},
		{width: 8, height: 4, output: 0},
		{width: 0, height: 0, output: 0},
		{width: 1, height: 1, output: 0},
	} {
		rect := image.Rect(0, 0, testdata.width, testdata.height)
		size := CalculateStorageSize(rect)
		require.Equal(t, testdata.output, size)
	}
}

func TestLSB_White(t *testing.T) {
	testLSB(t, "white")
}

func TestLSB_Black(t *testing.T) {
	testLSB(t, "black")
}

func testLSB(t *testing.T, name string) {
	pic, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.png", name))
	require.NoError(t, err)

	key := bytes.Repeat([]byte{1}, aes.Key256Bit)
	iv := bytes.Repeat([]byte{2}, aes.IVSize)
	plainData := bytes.Repeat(pic[:128], 5)

	picEnc, err := EncryptToPNG(pic, plainData, key, iv)
	require.NoError(t, err)

	dec, err := DecryptFromPNG(picEnc, key, iv)
	require.NoError(t, err)

	require.Equal(t, plainData, dec)

	// look the different about two pictures
	//
	// filename := fmt.Sprintf("testdata/%s_enc.png", name)
	// err = system.WriteFile(filename, picEnc)
	// require.NoError(t, err)
}

func testGeneratePNG(width, height int) *image.NRGBA64 {
	rect := image.Rect(0, 0, width, height)
	return image.NewNRGBA64(rect)
}

func testGeneratePNGBytes(t *testing.T, width, height int) []byte {
	img := testGeneratePNG(width, height)
	buf := bytes.NewBuffer(make([]byte, 0, width*height/4))
	err := png.Encode(buf, img)
	require.NoError(t, err)
	return buf.Bytes()
}

func TestEncryptToPNG(t *testing.T) {
	t.Run("invalid png data", func(t *testing.T) {
		img, err := EncryptToPNG(nil, nil, nil, nil)
		require.Error(t, err)
		require.Nil(t, img)
	})

	t.Run("failed to encrypt", func(t *testing.T) {
		pic := testGeneratePNGBytes(t, 160, 90)
		img, err := EncryptToPNG(pic, nil, nil, nil)
		require.Error(t, err)
		require.Nil(t, img)
	})

	t.Run("failed to encode", func(t *testing.T) {
		// must before patch, because testGeneratePNGBytes call png.Encode
		pic := testGeneratePNGBytes(t, 160, 90)

		encoder := new(png.Encoder)
		patch := func(interface{}, io.Writer, image.Image) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(encoder, "Encode", patch)
		defer pg.Unpatch()

		plainData := []byte{1, 2, 3, 4}
		key := random.Bytes(aes.Key256Bit)
		iv := random.Bytes(aes.IVSize)

		img, err := EncryptToPNG(pic, plainData, key, iv)
		monkey.IsMonkeyError(t, err)
		require.Nil(t, img)
	})
}

func TestDecryptFromPNG(t *testing.T) {
	t.Run("invalid png data", func(t *testing.T) {
		plainData, err := DecryptFromPNG(nil, nil, nil)
		require.Error(t, err)
		require.Nil(t, plainData)
	})

	t.Run("invalid png", func(t *testing.T) {
		rect := image.Rect(0, 0, 160, 90)
		img := image.NewRGBA(rect)
		buf := bytes.NewBuffer(make([]byte, 0, 128))
		err := png.Encode(buf, img)
		require.NoError(t, err)

		plainData, err := DecryptFromPNG(buf.Bytes(), nil, nil)
		require.Error(t, err)
		require.Nil(t, plainData)
	})
}

func TestEncrypt(t *testing.T) {
	t.Run("size > storage", func(t *testing.T) {
		img := testGeneratePNG(10, 10)
		plainData := make([]byte, 1024)

		pic, err := Encrypt(img, plainData, nil, nil)
		require.Error(t, err)
		require.Nil(t, pic)
	})

	t.Run("size > 4GB", func(t *testing.T) {
		img := testsuite.NewMockImage()
		img.SetMaxPoint(math.MaxInt32, math.MaxInt32)

		// create fake slice to make slice.Len too large
		plainData := make([]byte, 1024)
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&plainData)) // #nosec
		sliceHeader.Len = math.MaxInt32 + 1

		pic, err := Encrypt(img, plainData, nil, nil)
		require.Error(t, err)
		require.Nil(t, pic)
	})

	t.Run("alpha", func(t *testing.T) {
		img := testsuite.NewMockImage()
		img.SetPixel(1, 1, color.NRGBA64{
			R: 65535,
			G: 65535,
			B: 65535,
			A: 65521,
		})

		plainData := []byte{1, 2, 3, 4}
		key := random.Bytes(aes.Key256Bit)
		iv := random.Bytes(aes.IVSize)

		pic, err := Encrypt(img, plainData, key, iv)
		require.NoError(t, err)
		require.NotNil(t, pic)
	})
}

func TestDecrypt(t *testing.T) {
	key := random.Bytes(aes.Key256Bit)
	iv := random.Bytes(aes.IVSize)

	t.Run("bounds-ok", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			width  int
			height int
			size   int
		}{
			{width: 100, height: 200, size: 19951},
			{width: 100, height: 100, size: 9951},
			{width: 20, height: 3, size: 15},
		} {
			pic := testGeneratePNGBytes(t, testdata.width, testdata.height)
			plainData := random.Bytes(testdata.size)

			picEnc, err := EncryptToPNG(pic, plainData, key, iv)
			require.NoError(t, err)
			dec, err := DecryptFromPNG(picEnc, key, iv)
			require.NoError(t, err)

			require.Equal(t, plainData, dec)
		}
	})

	t.Run("bounds-failed", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			width  int
			height int
			size   int
		}{
			{width: 100, height: 200, size: 19951 + 1},
			{width: 100, height: 100, size: 9951 + 1},
			{width: 20, height: 3, size: 15 + 1},
		} {
			pic := testGeneratePNGBytes(t, testdata.width, testdata.height)
			plainData := random.Bytes(testdata.size)

			picEnc, err := EncryptToPNG(pic, plainData, key, iv)
			require.Error(t, err)
			require.Nil(t, picEnc)
		}
	})

	t.Run("invalid image size", func(t *testing.T) {
		img := testGeneratePNG(5, 5)

		plainData, err := Decrypt(img, key, iv)
		require.Error(t, err)
		require.Nil(t, plainData)
	})

	t.Run("invalid size in header", func(t *testing.T) {
		// set header first byte [85, 0, 0, 0]
		img := testGeneratePNG(100, 100)
		img.SetNRGBA64(0, 0, color.NRGBA64{
			R: 1,
			G: 1,
			B: 1,
			A: 1,
		})

		plainData, err := Decrypt(img, key, iv)
		require.Error(t, err)
		require.Nil(t, plainData)
	})

	t.Run("invalid cipher data", func(t *testing.T) {
		img := testGeneratePNG(100, 100)

		plainData, err := Decrypt(img, key, iv)
		require.Error(t, err)
		require.Nil(t, plainData)
	})

	t.Run("invalid hash", func(t *testing.T) {
		// set header first byte [0, 0, 32, 0]
		img := testGeneratePNG(100, 100)
		img.SetNRGBA64(0, 2, color.NRGBA64{
			R: 0,
			G: 256, // 0010 0000 -> 32
			B: 0,
			A: 0,
		})

		plainData, err := Decrypt(img, key, iv)
		require.Error(t, err)
		require.Nil(t, plainData)
	})

	t.Run("internal error", func(t *testing.T) {
		x := 0
		y := 0
		img := testGeneratePNG(1, 1)

		defer testsuite.DeferForPanic(t)
		readNRGBA64(img, make([]byte, 1024), &x, &y)
	})
}

func TestFuzz(t *testing.T) {
	for i := 0; i < 10; i++ {
		width := 30 + random.Int(300)
		height := 10 + random.Int(100)
		size := CalculateStorageSize(image.Rect(0, 0, width, height))

		pic := testGeneratePNGBytes(t, width, height)
		key := random.Bytes(aes.Key256Bit)
		iv := random.Bytes(aes.IVSize)

		// ok
		for _, size := range [...]int{
			size,
			size - 1,
			size - random.Int(100),
		} {
			plainData := random.Bytes(size)

			picEnc, err := EncryptToPNG(pic, plainData, key, iv)
			require.NoError(t, err)
			dec, err := DecryptFromPNG(picEnc, key, iv)
			require.NoError(t, err)

			require.Equal(t, plainData, dec)
		}

		// failed
		plainData := random.Bytes(size + 1)

		picEnc, err := EncryptToPNG(pic, plainData, key, iv)
		require.Error(t, err)
		require.Nil(t, picEnc)
	}
}
