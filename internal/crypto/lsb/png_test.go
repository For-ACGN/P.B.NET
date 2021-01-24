package lsb

import (
	"image"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func testGeneratePNG(width, height int) *image.NRGBA64 {
	rect := image.Rect(0, 0, width, height)
	return image.NewNRGBA64(rect)
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
	writer.mode = 0

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
		img := testsuite.NewMockConnWithReadError()

		reader, err := NewPNGReader(img)
		require.Error(t, err)
		require.Nil(t, reader)
	})

	t.Run("unsupported png format", func(t *testing.T) {
		file, err := os.Open("testdata/black.png")
		require.NoError(t, err)
		defer func() { _ = file.Close() }()

		reader, err := NewPNGReader(file)
		require.EqualError(t, err, "unsupported png format: *image.RGBA")
		require.Nil(t, reader)
	})
}

func TestPNGReaderWithInvalidMode(t *testing.T) {
	r := testGeneratePNGReader(t, 160, 90)
	reader, err := NewPNGReader(r)
	require.NoError(t, err)
	reader.mode = 0

	t.Run("Read", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		_, _ = reader.Read(make([]byte, 1024))
	})

	testsuite.IsDestroyed(t, reader)
}
