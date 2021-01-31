package lsb

import (
	"bytes"
	"image"
	"image/png"
	"io"
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
		r := testsuite.NewMockConnWithReadError()

		reader, err := NewPNGReader(r)
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
	img := testGeneratePNG(160, 90)
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	err := png.Encode(buf, img)
	require.NoError(t, err)

	reader, err := NewPNGReader(buf)
	require.NoError(t, err)
	reader.mode = 0

	t.Run("Read", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)
		_, _ = reader.Read(make([]byte, 1024))
	})

	testsuite.IsDestroyed(t, reader)
}

func BenchmarkNewPNGWriter(b *testing.B) {
	img := testGeneratePNG(1920, 1080)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := NewPNGWriter(img, PNGWithNRGBA32)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkPNGWriter_Write(b *testing.B) {
	img := testGeneratePNG(1920, 1080)
	writer, err := NewPNGWriter(img, PNGWithNRGBA32)
	require.NoError(b, err)

	data := make([]byte, 2048)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = writer.Write(data)
		if err != nil {
			b.Fatal(err)
		}

		// not use b.StopTimer
		writer.pngCommon.Reset()
	}

	b.StopTimer()
}

func BenchmarkNewPNGReader(b *testing.B) {
	img := testGeneratePNG(1920, 1080)
	buf := bytes.NewBuffer(make([]byte, 0, 1920*1080/2))
	err := png.Encode(buf, img)
	require.NoError(b, err)
	reader := bytes.NewReader(buf.Bytes())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = NewPNGReader(reader)
		if err != nil {
			b.Fatal(err)
		}

		_, err = reader.Seek(0, io.SeekStart)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkPNGReader_Read(b *testing.B) {
	img := testGeneratePNG(1920, 1080)
	buf := bytes.NewBuffer(make([]byte, 0, 1920*1080/2))
	err := png.Encode(buf, img)
	require.NoError(b, err)
	reader, err := NewPNGReader(buf)
	require.NoError(b, err)

	buffer := make([]byte, 2048)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = reader.Read(buffer)
		if err != nil {
			b.Fatal(err)
		}

		// not use b.StopTimer
		reader.pngCommon.Reset()
	}

	b.StopTimer()
}
