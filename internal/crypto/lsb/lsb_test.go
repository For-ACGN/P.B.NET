package lsb

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
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

func TestMode_String(t *testing.T) {
	for _, test := range tests {
		fmt.Println(test.mode)
	}
	fmt.Println(Mock)
	fmt.Println(Mode(1234578))
}

// test png image is 160*90
const testImageFullSize = 160 * 90

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
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
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

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)

				expected := convert.MergeBytes(testdata1, testdata2)
				actual := convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

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
				rv := 128 + random.Intn(128)
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

				rv = 32 + random.Intn(32)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)

				actual := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata, actual)

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

			t.Run("SetOffset", func(t *testing.T) {
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)
				offset := 512 + random.Int63n(128)

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

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata1, data)

				err = reader.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data = convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata2, data)

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
				rv := 128 + random.Intn(128)
				offset := 256 + random.Int63n(128)
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

				rv = 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata1, data)

				err = reader.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data = convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata2, data)

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

			t.Run("SetOffset with invalid value", func(t *testing.T) {
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

			t.Run("Reset", func(t *testing.T) {
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
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
				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata2Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)

				reader.Reset()

				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)

				actual := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata2, actual)

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
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// encrypt data
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

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData := convert.MergeBytes(buf1, buf2)

				expected := convert.MergeBytes(testdata1, testdata2)
				require.Equal(t, expected, plainData)

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("Common Full", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				testdata := random.Bytes(int(encrypter.Cap()))
				rv := 128 + random.Intn(128)
				testdata1 := testdata[:rv]
				testdata2 := testdata[rv:]
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				// write zero
				n, err = encrypter.Write(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// already full
				n, err = encrypter.Write(testdata1)
				require.Equal(t, ErrNoEnoughCapacity, err)
				require.Equal(t, 0, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv = 32 + random.Intn(32)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, testdata2Len+rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData := convert.MergeBytes(buf1, buf2)

				require.Equal(t, testdata, plainData)

				// read zero
				n, err = decrypter.Read(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// read EOF
				buf := make([]byte, testImageFullSize+1)
				n, err = decrypter.Read(buf)
				require.Equal(t, io.EOF, err)
				require.Equal(t, 0, n)

				// read remaining
				err = decrypter.Reset(nil)
				require.NoError(t, err)

				n, err = decrypter.Read(buf)
				require.NoError(t, err)
				require.Equal(t, testdata, buf[:n])

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("SetOffset", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)
				offset := 1024 + random.Int63n(128)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				err = encrypter.SetOffset(0)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				err = encrypter.SetOffset(offset)
				require.NoError(t, err)

				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				err = decrypter.SetOffset(0)
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata1, plainData)

				err = decrypter.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData = convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata2, plainData)

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("SetOffset Full", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				testdata := random.Bytes(int(encrypter.Cap()))
				rv := 128 + random.Intn(128)
				offset := 512 + random.Int63n(128)
				testdata1 := testdata[:rv]
				testdata2 := testdata[offset:]
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				err = encrypter.SetOffset(offset)
				require.NoError(t, err)

				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				// write zero
				n, err = encrypter.Write(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// already full
				n, err = encrypter.Write(testdata1)
				require.Equal(t, ErrNoEnoughCapacity, err)
				require.Equal(t, 0, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData := convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata1, plainData)

				err = decrypter.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				plainData = convert.MergeBytes(buf1, buf2)
				require.Equal(t, testdata2, plainData)

				// read zero
				n, err = decrypter.Read(nil)
				require.NoError(t, err)
				require.Equal(t, 0, n)

				// read EOF
				buf := make([]byte, testImageFullSize+1)
				n, err = decrypter.Read(buf)
				require.Equal(t, io.EOF, err)
				require.Equal(t, 0, n)

				// read remaining
				err = decrypter.SetOffset(offset)
				require.NoError(t, err)

				n, err = decrypter.Read(buf)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)
				require.Equal(t, testdata2, buf[:n])

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("SetOffset with invalid value", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)
				testdata := random.Bytes(128)

				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				err = encrypter.SetOffset(-1)
				require.Equal(t, ErrInvalidOffset, err)

				err = encrypter.SetOffset(math.MaxInt64)
				require.Equal(t, ErrInvalidOffset, err)

				n, err := encrypter.Write(testdata)
				require.NoError(t, err)
				require.Equal(t, 128, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				err = decrypter.SetOffset(-1)
				require.Equal(t, ErrInvalidOffset, err)

				err = decrypter.SetOffset(math.MaxInt64)
				require.Equal(t, ErrInvalidOffset, err)

				plainData, err := io.ReadAll(decrypter)
				require.NoError(t, err)

				require.Equal(t, testdata, plainData)

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("Reset without key", func(t *testing.T) {
				key := random.Bytes(aes.Key256Bit)
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len*2-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected := convert.MergeBytes(testdata1, testdata1)
				actual := convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				// reset encrypter
				err = encrypter.Reset(nil)
				require.NoError(t, err)

				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)
				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output.Reset()
				err = encrypter.Encode(output)
				require.NoError(t, err)

				decrypter, err = test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len*2-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected = convert.MergeBytes(testdata2, testdata2)
				actual = convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				// reset decrypter
				err = decrypter.Reset(nil)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len*2-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected = convert.MergeBytes(testdata2, testdata2)
				actual = convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("Reset with key", func(t *testing.T) {
				key1 := random.Bytes(aes.Key256Bit)
				key2 := random.Bytes(aes.Key256Bit)
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key1)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)
				n, err = encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key1)
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len*2-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected := convert.MergeBytes(testdata1, testdata1)
				actual := convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				require.Equal(t, key1, encrypter.Key())
				require.Equal(t, key1, decrypter.Key())

				// reset encrypter
				err = encrypter.Reset(key2)
				require.NoError(t, err)

				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)
				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output.Reset()
				err = encrypter.Encode(output)
				require.NoError(t, err)

				decrypter, err = test.newDecrypter(output.Bytes(), key2)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len*2-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected = convert.MergeBytes(testdata2, testdata2)
				actual = convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				// reset decrypter
				err = decrypter.Reset(key2)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len*2-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)

				expected = convert.MergeBytes(testdata2, testdata2)
				actual = convert.MergeBytes(buf1, buf2)
				require.Equal(t, expected, actual)

				// compare key
				require.Equal(t, key2, encrypter.Key())
				require.Equal(t, key2, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})

			t.Run("Reset with invalid key", func(t *testing.T) {
				key1 := random.Bytes(aes.Key256Bit)
				key2 := random.Bytes(aes.Key256Bit)
				invalidKey := random.Bytes(4)
				testdata := random.Bytes(128)

				encrypter, err := test.newEncrypter(img, key1)
				require.NoError(t, err)

				err = encrypter.Reset(invalidKey)
				require.Error(t, err)

				err = encrypter.Reset(key2)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata)
				require.NoError(t, err)
				require.Equal(t, 128, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				decrypter, err := test.newDecrypter(output.Bytes(), key1)
				require.NoError(t, err)

				err = decrypter.Reset(invalidKey)
				require.Error(t, err)

				err = decrypter.Reset(key2)
				require.NoError(t, err)

				plainData, err := io.ReadAll(decrypter)
				require.NoError(t, err)

				require.Equal(t, testdata, plainData)

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			})
		})
	}
}

func testGenerateImage() image.Image {
	width := 256 + random.Intn(128)
	height := 128 + random.Intn(64)
	rect := image.Rect(0, 0, width, height)
	img := image.NewNRGBA64(rect)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c := color.NRGBA64{
				R: uint16(random.Intn(65536)),
				G: uint16(random.Intn(65536)),
				B: uint16(random.Intn(65536)),
				A: uint16(random.Intn(65536)),
			}
			img.SetNRGBA64(x, y, c)
		}
	}
	return img
}

func TestWriterAndReader_Fuzz(t *testing.T) {
	for _, test := range tests {
		t.Run(test.mode.String(), func(t *testing.T) {
			for i := 0; i < 10; i++ {
				img := testGenerateImage()
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)
				offset := 1024 + random.Int63n(512)

				// writer
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

				// reader
				reader, err := test.newReader(output.Bytes())
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data := convert.MergeBytes(buf1, buf2)

				require.Equal(t, testdata1, data)

				err = reader.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(reader, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(reader, buf2)
				require.NoError(t, err)
				data = convert.MergeBytes(buf1, buf2)

				require.Equal(t, testdata2, data)

				// compare image
				require.Equal(t, img, writer.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, reader.Image())

				// compare mode
				require.Equal(t, writer.Mode(), reader.Mode())

				testsuite.IsDestroyed(t, writer)
				testsuite.IsDestroyed(t, reader)
			}
		})
	}
}

func TestEncrypterAndDecrypter_Fuzz(t *testing.T) {
	for _, test := range tests {
		t.Run(test.mode.String(), func(t *testing.T) {
			for i := 0; i < 10; i++ {
				img := testGenerateImage()
				key := random.Bytes(aes.Key256Bit)
				testdata1 := random.Bytes(256 + random.Intn(256))
				testdata2 := random.Bytes(512 + random.Intn(512))
				testdata1Len := len(testdata1)
				testdata2Len := len(testdata2)
				offset := 1024 + random.Int63n(512)

				// encrypt data
				encrypter, err := test.newEncrypter(img, key)
				require.NoError(t, err)

				n, err := encrypter.Write(testdata1)
				require.NoError(t, err)
				require.Equal(t, testdata1Len, n)

				err = encrypter.SetOffset(offset)
				require.NoError(t, err)

				n, err = encrypter.Write(testdata2)
				require.NoError(t, err)
				require.Equal(t, testdata2Len, n)

				output := bytes.NewBuffer(make([]byte, 0, 8192))
				err = encrypter.Encode(output)
				require.NoError(t, err)

				// decrypt data
				decrypter, err := test.newDecrypter(output.Bytes(), key)
				require.NoError(t, err)

				rv := 64 + random.Intn(64)
				buf1 := make([]byte, testdata1Len-rv)
				buf2 := make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				data := convert.MergeBytes(buf1, buf2)

				require.Equal(t, testdata1, data)

				err = decrypter.SetOffset(offset)
				require.NoError(t, err)

				rv = 64 + random.Intn(64)
				buf1 = make([]byte, testdata2Len-rv)
				buf2 = make([]byte, rv)
				_, err = io.ReadFull(decrypter, buf1)
				require.NoError(t, err)
				_, err = io.ReadFull(decrypter, buf2)
				require.NoError(t, err)
				data = convert.MergeBytes(buf1, buf2)

				require.Equal(t, testdata2, data)

				// compare key
				require.Equal(t, key, encrypter.Key())
				require.Equal(t, key, decrypter.Key())

				// compare image
				require.Equal(t, img, encrypter.Image())

				outputPNG, err := png.Decode(bytes.NewReader(output.Bytes()))
				require.NoError(t, err)
				require.Equal(t, outputPNG, decrypter.Image())

				// compare capacity
				require.Equal(t, encrypter.Cap(), decrypter.Cap())

				// compare mode
				require.Equal(t, encrypter.Mode(), decrypter.Mode())

				testsuite.IsDestroyed(t, encrypter)
				testsuite.IsDestroyed(t, decrypter)
			}
		})
	}
}

var errMockError = errors.New("mock error")

type mockWriter struct {
	setOffsetError bool
}

func (mockWriter) Write([]byte) (int, error) {
	return 0, errMockError
}

func (mockWriter) Encode(io.Writer) error {
	return errMockError
}

func (mw *mockWriter) SetOffset(int64) error {
	if mw.setOffsetError {
		return errMockError
	}
	return nil
}

func (mockWriter) Reset() {}

func (mockWriter) Image() image.Image {
	return nil
}

func (mockWriter) Cap() int64 {
	return 0
}

func (mockWriter) Mode() Mode {
	return Mock
}

type mockReader struct{}

func (mockReader) Read([]byte) (int, error) {
	return 0, errMockError
}

func (mockReader) SetOffset(int64) error {
	return errMockError
}

func (mockReader) Reset() {}

func (mockReader) Image() image.Image {
	return nil
}

func (mockReader) Cap() int64 {
	return 0
}

func (mockReader) Mode() Mode {
	return Mock
}
