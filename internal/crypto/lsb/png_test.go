package lsb

import (
	"bytes"
	"image"
	"image/png"
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
