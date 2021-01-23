package lsb

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/hmac"
	"project/internal/patch/monkey"
	"project/internal/security"
	"project/internal/testsuite"
)

func testGeneratePNG(width, height int) *image.NRGBA64 {
	rect := image.Rect(0, 0, width, height)
	return image.NewNRGBA64(rect)
}

func testGeneratePNGReader(t *testing.T, width, height int) io.Reader {
	img := testGeneratePNG(width, height)
	buf := bytes.NewBuffer(make([]byte, 0, width*height/4))
	err := png.Encode(buf, img)
	require.NoError(t, err)
	return buf
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
	img := testGeneratePNGReader(t, 160, 90)
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
	require.Equal(t, errMockError, err)
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
	require.Equal(t, errMockError, err)
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
		require.Equal(t, errMockError, err)
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
		img := testsuite.NewMockConnWithReadError()

		_, err := NewPNGDecrypter(img, nil)
		require.Error(t, err)
	})

	t.Run("too small image", func(t *testing.T) {
		img := testGeneratePNGReader(t, 1, 2)

		_, err := NewPNGDecrypter(img, nil)
		require.Equal(t, ErrImgTooSmall, err)
	})

	t.Run("failed to reset", func(t *testing.T) {
		img := testGeneratePNGReader(t, 160, 90)
		invalidKey := make([]byte, 8)

		_, err := NewPNGDecrypter(img, invalidKey)
		require.Error(t, err)
	})
}

func TestPNGDecrypter_Read(t *testing.T) {
	img := testGeneratePNGReader(t, 160, 90)
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

func TestPNGDecrypter_validate(t *testing.T) {
	img := testGeneratePNGReader(t, 160, 90)
	key := make([]byte, aes.Key256Bit)
	decrypter, err := NewPNGDecrypter(img, key)
	require.NoError(t, err)

	t.Run("failed to read hmac signature", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(r io.Reader, b []byte) (int, error) {
			if len(b) == sha256.Size {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return io.ReadFull(r, b)
		}
		pg = monkey.Patch(io.ReadFull, patch)
		defer pg.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("failed to read iv", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(r io.Reader, b []byte) (int, error) {
			if len(b) == aes.IVSize {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return io.ReadFull(r, b)
		}
		pg = monkey.Patch(io.ReadFull, patch)
		defer pg.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("failed to decrypt size", func(t *testing.T) {
		var ctr *aes.CTR
		patch := func(interface{}, []byte, []byte) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.PatchInstanceMethod(ctr, "DecryptWithIV", patch)
		defer pg.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("invalid cipher data size", func(t *testing.T) {
		patch := func([]byte) int64 {
			return 0
		}
		pg := monkey.Patch(convert.BEBytesToInt64, patch)
		defer pg.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		require.Error(t, err)
	})

	t.Run("failed to compare signature", func(t *testing.T) {
		patch1 := func([]byte) int64 {
			return 128
		}
		pg1 := monkey.Patch(convert.BEBytesToInt64, patch1)
		defer pg1.Unpatch()

		patch2 := func(io.Writer, io.Reader, int64) (int64, error) {
			return 0, monkey.Error
		}
		pg2 := monkey.Patch(io.CopyN, patch2)
		defer pg2.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("invalid hmac signature", func(t *testing.T) {
		patch := func([]byte) int64 {
			return 128
		}
		pg := monkey.Patch(convert.BEBytesToInt64, patch)
		defer pg.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		require.Error(t, err)
	})

	t.Run("failed to reset offset", func(t *testing.T) {
		patch1 := func([]byte) int64 {
			return 128
		}
		pg1 := monkey.Patch(convert.BEBytesToInt64, patch1)
		defer pg1.Unpatch()

		patch2 := func([]byte, []byte) bool {
			return true
		}
		pg2 := monkey.Patch(hmac.Equal, patch2)
		defer pg2.Unpatch()

		decrypter.offset = -1024
		defer func() { decrypter.offset = 0 }()

		_, err = decrypter.Read(make([]byte, 16))
		require.Error(t, err)
	})

	t.Run("failed to set stream", func(t *testing.T) {
		patch1 := func([]byte) int64 {
			return 128
		}
		pg1 := monkey.Patch(convert.BEBytesToInt64, patch1)
		defer pg1.Unpatch()

		patch2 := func([]byte, []byte) bool {
			return true
		}
		pg2 := monkey.Patch(hmac.Equal, patch2)
		defer pg2.Unpatch()

		var ctr *aes.CTR
		patch3 := func(interface{}, []byte) error {
			return monkey.Error
		}
		pg3 := monkey.PatchInstanceMethod(ctr, "SetStream", patch3)
		defer pg3.Unpatch()

		_, err = decrypter.Read(make([]byte, 16))
		monkey.IsMonkeyError(t, err)
	})

	testsuite.IsDestroyed(t, decrypter)
}
