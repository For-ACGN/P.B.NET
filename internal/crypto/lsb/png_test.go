package lsb

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

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
	writer.mode = InvalidMode

	t.Run("Write", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		_, err = writer.Write([]byte{0})
		require.Error(t, err)
	})

	t.Run("Encode", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		err = writer.Encode(nil)
		require.Error(t, err)
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
	reader.mode = InvalidMode

	t.Run("Read", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		_, err = reader.Read(make([]byte, 1024))
		require.Error(t, err)
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

}

func TestPNGEncrypter_Encode(t *testing.T) {

}

func TestPNGEncrypter_SetOffset(t *testing.T) {

}
