package lsb

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/random"
	"project/internal/testsuite"
)

var tests = [...]*struct {
	mode      Mode
	newWriter func(img image.Image) (Writer, error)
	newReader func(img []byte) (Reader, error)
}{
	{
		PNGWithNRGBA32,
		func(img image.Image) (Writer, error) {
			return NewPNGWriter(img, PNGWithNRGBA32)
		},
		func(img []byte) (Reader, error) {
			return NewPNGReader(img)
		},
	},
	{
		PNGWithNRGBA64,
		func(img image.Image) (Writer, error) {
			return NewPNGWriter(img, PNGWithNRGBA64)
		},
		func(img []byte) (Reader, error) {
			return NewPNGReader(img)
		},
	},
}

func TestMode_String(t *testing.T) {
	for _, test := range tests {
		fmt.Println(test.mode)
	}
	fmt.Println(Mode(1234578))
}

func TestWriterAndReader(t *testing.T) {
	t.Run("black", func(t *testing.T) { testWriterAndReader(t, "black") })
	t.Run("white", func(t *testing.T) { testWriterAndReader(t, "white") })
}

func testWriterAndReader(t *testing.T, name string) {
	file, err := os.Open(fmt.Sprintf("testdata/%s.png", name))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()
	img, err := png.Decode(file)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.mode.String(), func(t *testing.T) {
			t.Run("Common", func(t *testing.T) {
				testdata1 := random.Bytes(256 + random.Int(256))
				testdata2 := random.Bytes(512 + random.Int(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// write data
				writer, err := test.newWriter(img)
				require.NoError(t, err)

				n, err := writer.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = writer.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				// read data
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				rv := 64 + random.Int(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result := append(buf1, buf2...)

				expected := append(testdata1, testdata2...)
				require.Equal(t, expected, result)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})

			t.Run("Common Full", func(t *testing.T) {
				// write data
				writer, err := test.newWriter(img)
				require.NoError(t, err)

				testdata := random.Bytes(int(writer.Cap()))
				rv := 128 + random.Int(128)
				testdata1 := testdata[:rv]
				testdata2 := testdata[rv:]
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				n, err := writer.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = writer.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				// read data
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				rv = 32 + random.Int(32)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result := append(buf1, buf2...)

				require.Equal(t, testdata, result)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})

			t.Run("Reset", func(t *testing.T) {
				testdata1 := random.Bytes(256 + random.Int(256))
				testdata2 := random.Bytes(512 + random.Int(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// write data
				writer, err := test.newWriter(img)
				require.NoError(t, err)

				// reset writer
				n, err := writer.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				writer.Reset()

				n, err = writer.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				// read data
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				// reset reader
				rv := 64 + random.Int(64)
				buf1 := make([]byte, testdata2Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				reader.Reset()

				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result := append(buf1, buf2...)

				require.Equal(t, testdata2, result)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})

			t.Run("SetOffset", func(t *testing.T) {
				testdata1 := random.Bytes(256 + random.Int(256))
				testdata2 := random.Bytes(512 + random.Int(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)
				offset := int64(512 + random.Int(128))

				// write data
				writer, err := test.newWriter(img)
				require.NoError(t, err)

				n, err := writer.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				err = writer.SetOffset(offset)
				require.NoError(t, err)

				n, err = writer.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				// read data
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				rv := 64 + random.Int(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result1 := append(buf1, buf2...)

				err = reader.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Int(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result2 := append(buf1, buf2...)

				expected := append(testdata1, testdata2...)
				result := append(result1, result2...)

				require.Equal(t, expected, result)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})

			t.Run("SetOffset Full", func(t *testing.T) {
				// write data
				writer, err := test.newWriter(img)
				require.NoError(t, err)

				testdata := random.Bytes(int(writer.Cap()))
				rv := 128 + random.Int(128)
				offset := int64(256 + random.Int(128))
				testdata1 := testdata[:rv]
				testdata2 := testdata[offset:]
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				n, err := writer.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				err = writer.SetOffset(offset)
				require.NoError(t, err)

				n, err = writer.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				// read data
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				rv = 64 + random.Int(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result1 := append(buf1, buf2...)

				err = reader.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Int(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				result2 := append(buf1, buf2...)

				expected := append(testdata1, testdata2...)
				result := append(result1, result2...)

				require.Equal(t, expected, result)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})
		})
	}
}

func TestEncrypterAndDecrypter(t *testing.T) {

}
