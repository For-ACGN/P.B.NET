package lsb

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/random"
	"project/internal/testsuite"
)

var tests = [...]*struct {
	mode         Mode
	newWriter    func(img image.Image) (Writer, error)
	newReader    func(img []byte) (Reader, error)
	newEncrypter func(img image.Image, key []byte) (Encrypter, error)
	newDecrypter func(img, key []byte) (Decrypter, error)
}{
	{
		PNGWithNRGBA32,
		func(img image.Image) (Writer, error) {
			return NewPNGWriter(img, PNGWithNRGBA32)
		},
		func(img []byte) (Reader, error) {
			return NewPNGReader(img)
		},
		func(img image.Image, key []byte) (Encrypter, error) {
			return NewPNGEncrypter(img, PNGWithNRGBA32, key)
		},
		func(img, key []byte) (Decrypter, error) {
			return NewPNGDecrypter(img, key)
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
		func(img image.Image, key []byte) (Encrypter, error) {
			return NewPNGEncrypter(img, PNGWithNRGBA64, key)
		},
		func(img, key []byte) (Decrypter, error) {
			return NewPNGDecrypter(img, key)
		},
	},
}

// test png image is 160*90
const testImageFullSize = 160 * 90

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

				// write zero
				n, err = writer.Write(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// already full
				n, err = writer.Write(testdata1)
				require.Equal(t, ErrNoEnoughCapacity, err)
				require.Equal(t, 0, n)

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

				// read zero
				n, err = reader.Read(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// read EOF
				buf := make([]byte, testImageFullSize+1)
				n, err = reader.Read(buf)
				require.Equal(t, io.EOF, err)
				require.Equal(t, 0, n)

				// read remaining
				reader.Reset()

				n, err = reader.Read(buf)
				require.NoError(t, err)
				require.Equal(t, testImageFullSize, n)
				require.Equal(t, testdata, buf[:n])

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

				// write zero
				n, err = writer.Write(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// already full
				n, err = writer.Write(testdata1)
				require.Equal(t, ErrNoEnoughCapacity, err)
				require.Equal(t, 0, n)

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

				// can't use append(testdata1, testdata2...), otherwise it
				// will change testdata2 and influence "read remaining" sub test
				expected := append([]byte{}, testdata1...)
				expected = append(expected, testdata2...)
				result := append(result1, result2...)
				require.Equal(t, expected, result)

				// read zero
				n, err = reader.Read(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// read EOF
				buf := make([]byte, testImageFullSize+1)
				n, err = reader.Read(buf)
				require.Equal(t, io.EOF, err)
				require.Equal(t, 0, n)

				// read remaining
				err = reader.SetOffset(offset)
				require.NoError(t, err)

				n, err = reader.Read(buf)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)
				require.Equal(t, testdata2, buf[:n])

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

			t.Run("SetOffset Invalid", func(t *testing.T) {
				testdata := random.Bytes(128)

				writer, err := test.newWriter(img)
				require.NoError(t, err)

				err = writer.SetOffset(-1)
				require.Equal(t, ErrInvalidOffset, err)

				err = writer.SetOffset(math.MaxInt64)
				require.Equal(t, ErrInvalidOffset, err)

				n, err := writer.Write(testdata)
				require.NoError(t, err)
				require.Equal(t, 128, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = writer.Encode(output)
				require.NoError(t, err)

				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				err = reader.SetOffset(-1)
				require.Equal(t, ErrInvalidOffset, err)

				err = reader.SetOffset(math.MaxInt64)
				require.Equal(t, ErrInvalidOffset, err)

				buf := make([]byte, 128)
				_, err = io.ReadFull(reader, buf)
				require.NoError(t, err)

				require.Equal(t, testdata, buf)

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			})
		})
	}
}

func TestEncrypterAndDecrypter(t *testing.T) {
	t.Run("black", func(t *testing.T) { testEncrypterAndDecrypter(t, "black") })
	t.Run("white", func(t *testing.T) { testEncrypterAndDecrypter(t, "white") })
}

func testEncrypterAndDecrypter(t *testing.T, name string) {
	file, err := os.Open(fmt.Sprintf("testdata/%s.png", name))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	img, err := png.Decode(file)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.mode.String(), func(t *testing.T) {
			t.Run("Common", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)
				testdata1 := random.Bytes(256 + random.Int(256))
				testdata2 := random.Bytes(512 + random.Int(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// write data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// read data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv := 64 + random.Int(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				result := append(buf1, buf2...)

				expected := append(testdata1, testdata2...)

				fmt.Println(testdata1Len)
				fmt.Println(testdata2Len)

				fmt.Println(len(buf1))
				fmt.Println(len(buf2))

				convert.DumpBytes(expected)
				convert.DumpBytes(result)

				return

				require.Equal(t, expected, result)

				// compare image
				require.Equal(t, img, encrypter.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("Common Full", func(t *testing.T) {

			})

			t.Run("Reset", func(t *testing.T) {

			})

			t.Run("SetOffset", func(t *testing.T) {

			})

			t.Run("SetOffset Full", func(t *testing.T) {

			})

			t.Run("SetOffset Invalid", func(t *testing.T) {

			})
		})
	}
}
