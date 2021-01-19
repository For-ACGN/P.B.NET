package lsb

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/aes"
	"project/internal/patch/monkey"
	"project/internal/security"
	"project/internal/testsuite"
)

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

func TestNewPNGWriter(t *testing.T) {
	img := testGeneratePNG(160, 90)

	writer, err := NewPNGWriter(img, 0)
	require.EqualError(t, err, "png writer with unknown mode: 0")
	require.Nil(t, writer)
}

func TestPNGWriterWithInvalidMode(t *testing.T) {
	img := testGeneratePNG(160, 90)
	writer, err := NewPNGWriter(img, PNGWithNRGBA32)
	require.NoError(t, err)
	writer.mode = Invalid

	t.Run("Write", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		_, _ = writer.Write([]byte{0})
	})

	t.Run("Encode", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		_ = writer.Encode(nil)
	})

	t.Run("Reset", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		writer.Reset()
	})

	testsuite.IsDestroyed(t, writer)
}

func TestNewPNGReader(t *testing.T) {
	t.Run("invalid image", func(t *testing.T) {
		reader, err := NewPNGReader(nil)
		require.Error(t, err)
		require.Nil(t, reader)
	})

	t.Run("unsupported png format", func(t *testing.T) {
		data, err := os.ReadFile("testdata/black.png")
		require.NoError(t, err)

		reader, err := NewPNGReader(data)
		require.EqualError(t, err, "unsupported png format: *image.RGBA")
		require.Nil(t, reader)
	})
}

func TestPNGReaderWithInvalidMode(t *testing.T) {
	img := testGeneratePNGBytes(t, 160, 90)
	reader, err := NewPNGReader(img)
	require.NoError(t, err)
	reader.mode = Invalid

	t.Run("Read", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		_, _ = reader.Read(make([]byte, 1024))
	})

	testsuite.IsDestroyed(t, reader)
}

func TestNewPNGEncrypter(t *testing.T) {
	t.Run("unknown mode", func(t *testing.T) {
		img := testGeneratePNG(160, 90)

		encrypter, err := NewPNGEncrypter(img, 0, nil)
		require.EqualError(t, err, "png writer with unknown mode: 0")
		require.Nil(t, encrypter)
	})

	t.Run("too small image", func(t *testing.T) {
		img := testGeneratePNG(1, 2)

		encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, nil)
		require.Equal(t, err, ErrImgTooSmall)
		require.Nil(t, encrypter)
	})

	t.Run("failed to reset", func(t *testing.T) {
		img := testGeneratePNG(160, 90)
		invalidKey := make([]byte, 8)

		encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, invalidKey)
		require.Error(t, err)
		require.Nil(t, encrypter)
	})
}

func TestPNGEncrypter_Write(t *testing.T) {
	img := testGeneratePNG(160, 90)
	Key := make([]byte, aes.Key256Bit)

	encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, Key)
	require.NoError(t, err)
	encrypter.writer = new(mockWriter)

	_, err = encrypter.Write([]byte{0})
	require.Equal(t, mockError, err)
}

func TestPNGEncrypter_Encode(t *testing.T) {
	img := testGeneratePNG(160, 90)
	Key := make([]byte, aes.Key256Bit)

	encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, Key)
	require.NoError(t, err)

	_, err = encrypter.Write([]byte{0})
	require.NoError(t, err)

	encrypter.writer = new(mockWriter)

	err = encrypter.Encode(nil)
	require.Equal(t, mockError, err)
}

func TestPNGEncrypter_SetOffset(t *testing.T) {
	img := testGeneratePNG(160, 90)
	Key := make([]byte, aes.Key256Bit)
	encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, Key)
	require.NoError(t, err)

	t.Run("failed to write header", func(t *testing.T) {
		writer := encrypter.writer
		defer func() { encrypter.writer = writer }()

		_, err = encrypter.Write([]byte{0})
		require.NoError(t, err)

		encrypter.writer = new(mockWriter)

		err = encrypter.SetOffset(1)
		require.Equal(t, mockError, err)
	})

	t.Run("failed to generate IV", func(t *testing.T) {
		patch := func() ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(aes.GenerateIV, patch)
		defer pg.Unpatch()

		err = encrypter.SetOffset(1)
		monkey.IsMonkeyError(t, err)
	})

	t.Run("failed to set stream", func(t *testing.T) {
		patch := func() ([]byte, error) {
			return make([]byte, 8), nil
		}
		pg := monkey.Patch(aes.GenerateIV, patch)
		defer pg.Unpatch()

		err = encrypter.SetOffset(1)
		require.Equal(t, aes.ErrInvalidIVSize, err)
	})

	testsuite.IsDestroyed(t, encrypter)
}

func TestPNGEncrypter_writeHeader(t *testing.T) {
	img := testGeneratePNG(160, 90)
	Key := make([]byte, aes.Key256Bit)
	encrypter, err := NewPNGEncrypter(img, PNGWithNRGBA32, Key)
	require.NoError(t, err)

	t.Run("failed to encrypt size buffer", func(t *testing.T) {
		iv := encrypter.iv
		defer func() { encrypter.iv = iv }()

		encrypter.iv = security.NewBytes(make([]byte, 4))

		_, err = encrypter.Write([]byte{0})
		require.NoError(t, err)

		defer testsuite.DeferForPanic(t)
		_ = encrypter.SetOffset(1)
	})

	t.Run("failed to set offset", func(t *testing.T) {
		offset := encrypter.offset
		defer func() { encrypter.offset = offset }()

		encrypter.offset = -1024

		_, err = encrypter.Write([]byte{0})
		require.NoError(t, err)

		defer testsuite.DeferForPanic(t)
		_ = encrypter.SetOffset(1)
	})

	testsuite.IsDestroyed(t, encrypter)
}

func TestNewPNGDecrypter(t *testing.T) {
	t.Run("invalid image", func(t *testing.T) {
		_, err := NewPNGDecrypter(nil, nil)
		require.Error(t, err)
	})

	t.Run("too small image", func(t *testing.T) {
		img := testGeneratePNGBytes(t, 1, 2)

		_, err := NewPNGDecrypter(img, nil)
		require.Equal(t, ErrImgTooSmall, err)
	})

	t.Run("failed to reset", func(t *testing.T) {
		img := testGeneratePNGBytes(t, 160, 90)
		invalidKey := make([]byte, 8)

		_, err := NewPNGDecrypter(img, invalidKey)
		require.Error(t, err)
	})
}

func TestPNGDecrypter_Read(t *testing.T) {
	img := testGeneratePNGBytes(t, 160, 90)
	key := make([]byte, aes.Key256Bit)
	decrypter, err := NewPNGDecrypter(img, key)
	require.NoError(t, err)
	decrypter.reader = new(mockReader)

	t.Run("failed to validate", func(t *testing.T) {
		_, err = decrypter.Read(nil)
		require.Error(t, err)
	})

	t.Run("failed to read cipher data", func(t *testing.T) {
		decrypter.size = 128

		_, err = decrypter.Read(make([]byte, 16))
		require.Error(t, err)
	})

	testsuite.IsDestroyed(t, decrypter)
}
