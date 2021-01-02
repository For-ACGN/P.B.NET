package convert

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

// copy from internal/testsuite/testsuite.go
func testDeferForPanic(t testing.TB) {
	r := recover()
	require.NotNil(t, r)
	t.Logf("\npanic in %s:\n%s\n", t.Name(), r)
}

func TestBENumberToBytes(t *testing.T) {
	if !bytes.Equal(BEInt16ToBytes(int16(0x0102)), []byte{1, 2}) {
		t.Fatal("BEInt16ToBytes() with invalid number")
	}
	if !bytes.Equal(BEInt32ToBytes(int32(0x01020304)), []byte{1, 2, 3, 4}) {
		t.Fatal("BEInt32ToBytes() with invalid number")
	}
	if !bytes.Equal(BEInt64ToBytes(int64(0x0102030405060708)), []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatal("BEInt64ToBytes() with invalid number")
	}
	if !bytes.Equal(BEUint16ToBytes(uint16(0x0102)), []byte{1, 2}) {
		t.Fatal("BEUint16ToBytes() with invalid number")
	}
	if !bytes.Equal(BEUint32ToBytes(uint32(0x01020304)), []byte{1, 2, 3, 4}) {
		t.Fatal("BEUint32ToBytes() with invalid number")
	}
	if !bytes.Equal(BEUint64ToBytes(uint64(0x0102030405060708)), []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatal("BEUint64ToBytes() with invalid number")
	}
	if !bytes.Equal(BEFloat32ToBytes(float32(123.123)), []byte{66, 246, 62, 250}) {
		t.Fatal("BEFloat32ToBytes() with invalid number")
	}
	if !bytes.Equal(BEFloat64ToBytes(123.123), []byte{64, 94, 199, 223, 59, 100, 90, 29}) {
		t.Fatal("BEFloat64ToBytes() with invalid number")
	}
}

func TestBEBytesToNumber(t *testing.T) {
	if BEBytesToInt16([]byte{1, 2}) != 0x0102 {
		t.Fatal("BEBytesToInt16() with invalid bytes")
	}
	if BEBytesToInt32([]byte{1, 2, 3, 4}) != 0x01020304 {
		t.Fatal("BEBytesToInt32() with invalid bytes")
	}
	if BEBytesToInt64([]byte{1, 2, 3, 4, 5, 6, 7, 8}) != 0x0102030405060708 {
		t.Fatal("BEBytesToInt64() with invalid bytes")
	}
	if BEBytesToUint16([]byte{1, 2}) != 0x0102 {
		t.Fatal("BEBytesToUint16() with invalid bytes")
	}
	if BEBytesToUint32([]byte{1, 2, 3, 4}) != 0x01020304 {
		t.Fatal("BEBytesToUint32() with invalid bytes")
	}
	if BEBytesToUint64([]byte{1, 2, 3, 4, 5, 6, 7, 8}) != 0x0102030405060708 {
		t.Fatal("BEBytesToUint64() with invalid bytes")
	}
	if BEBytesToFloat32([]byte{66, 246, 62, 250}) != 123.123 {
		t.Fatal("BEBytesToFloat32() with invalid bytes")
	}
	if BEBytesToFloat64([]byte{64, 94, 199, 223, 59, 100, 90, 29}) != 123.123 {
		t.Fatal("BEBytesToFloat64() with invalid bytes")
	}

	// negative number
	n := int64(-0x12345678)
	if BEBytesToInt64(BEInt64ToBytes(n)) != n {
		t.Fatal("negative number")
	}
}

func TestBEBytesToNumberWithInvalidBytes(t *testing.T) {
	t.Run("BEBytesToInt16", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToInt16([]byte{1})
	})

	t.Run("BEBytesToInt32", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToInt32([]byte{1})
	})

	t.Run("BEBytesToInt64", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToInt64([]byte{1})
	})

	t.Run("BEBytesToUint16", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToUint16([]byte{1})
	})

	t.Run("BEBytesToUint32", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToUint32([]byte{1})
	})

	t.Run("BEBytesToUint64", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToUint64([]byte{1})
	})

	t.Run("BEBytesToFloat32", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToFloat32([]byte{1})
	})

	t.Run("BEBytesToFloat64", func(t *testing.T) {
		defer testDeferForPanic(t)
		BEBytesToFloat64([]byte{1})
	})
}

func TestLENumberToBytes(t *testing.T) {
	if !bytes.Equal(LEInt16ToBytes(int16(0x0102)), []byte{2, 1}) {
		t.Fatal("LEInt16ToBytes() with invalid number")
	}
	if !bytes.Equal(LEInt32ToBytes(int32(0x01020304)), []byte{4, 3, 2, 1}) {
		t.Fatal("LEInt32ToBytes() with invalid number")
	}
	if !bytes.Equal(LEInt64ToBytes(int64(0x0102030405060708)), []byte{8, 7, 6, 5, 4, 3, 2, 1}) {
		t.Fatal("LEInt64ToBytes() with invalid number")
	}
	if !bytes.Equal(LEUint16ToBytes(uint16(0x0102)), []byte{2, 1}) {
		t.Fatal("LEUint16ToBytes() with invalid number")
	}
	if !bytes.Equal(LEUint32ToBytes(uint32(0x01020304)), []byte{4, 3, 2, 1}) {
		t.Fatal("LEUint32ToBytes() with invalid number")
	}
	if !bytes.Equal(LEUint64ToBytes(uint64(0x0102030405060708)), []byte{8, 7, 6, 5, 4, 3, 2, 1}) {
		t.Fatal("LEUint64ToBytes() with invalid number")
	}
	if !bytes.Equal(LEFloat32ToBytes(float32(123.123)), []byte{250, 62, 246, 66}) {
		t.Fatal("LEFloat32ToBytes() with invalid number")
	}
	if !bytes.Equal(LEFloat64ToBytes(123.123), []byte{29, 90, 100, 59, 223, 199, 94, 64}) {
		t.Fatal("LEFloat64ToBytes() with invalid number")
	}
}

func TestLEBytesToNumber(t *testing.T) {
	if LEBytesToInt16([]byte{2, 1}) != 0x0102 {
		t.Fatal("LEBytesToInt16() with invalid bytes")
	}
	if LEBytesToInt32([]byte{4, 3, 2, 1}) != 0x01020304 {
		t.Fatal("LEBytesToInt32() with invalid bytes")
	}
	if LEBytesToInt64([]byte{8, 7, 6, 5, 4, 3, 2, 1}) != 0x0102030405060708 {
		t.Fatal("LEBytesToInt64() with invalid bytes")
	}
	if LEBytesToUint16([]byte{2, 1}) != 0x0102 {
		t.Fatal("LEBytesToUint16() with invalid bytes")
	}
	if LEBytesToUint32([]byte{4, 3, 2, 1}) != 0x01020304 {
		t.Fatal("LEBytesToUint32() with invalid bytes")
	}
	if LEBytesToUint64([]byte{8, 7, 6, 5, 4, 3, 2, 1}) != 0x0102030405060708 {
		t.Fatal("LEBytesToUint64() with invalid bytes")
	}
	if LEBytesToFloat32([]byte{250, 62, 246, 66}) != 123.123 {
		t.Fatal("LEBytesToFloat32() with invalid bytes")
	}
	if LEBytesToFloat64([]byte{29, 90, 100, 59, 223, 199, 94, 64}) != 123.123 {
		t.Fatal("LEBytesToFloat64() with invalid bytes")
	}

	// negative number
	n := int64(-0x12345678)
	if LEBytesToInt64(LEInt64ToBytes(n)) != n {
		t.Fatal("negative number")
	}
}

func TestLEBytesToNumberWithInvalidBytes(t *testing.T) {
	t.Run("LEBytesToInt16", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToInt16([]byte{1})
	})

	t.Run("LEBytesToInt32", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToInt32([]byte{1})
	})

	t.Run("LEBytesToInt64", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToInt64([]byte{1})
	})

	t.Run("LEBytesToUint16", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToUint16([]byte{1})
	})

	t.Run("LEBytesToUint32", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToUint32([]byte{1})
	})

	t.Run("LEBytesToUint64", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToUint64([]byte{1})
	})

	t.Run("LEBytesToFloat32", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToFloat32([]byte{1})
	})

	t.Run("LEBytesToFloat64", func(t *testing.T) {
		defer testDeferForPanic(t)
		LEBytesToFloat64([]byte{1})
	})
}

func TestAbsInt64(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  int64
		output int64
	}{
		{-1, 1},
		{0, 0},
		{1, 1},
		{-10, 10},
		{10, 10},
	} {
		require.Equal(t, testdata.output, AbsInt64(testdata.input))
	}
}

func TestByteUnit(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			input  uint64
			output string
		}{
			{1023 * Byte, "1023 Byte"},
			{1024 * Byte, "1 KiB"},
			{1536 * Byte, "1.5 KiB"},
			{MiB, "1 MiB"},
			{1536 * KiB, "1.5 MiB"},
			{GiB, "1 GiB"},
			{1536 * MiB, "1.5 GiB"},
			{TiB, "1 TiB"},
			{1536 * GiB, "1.5 TiB"},
			{PiB, "1 PiB"},
			{1536 * TiB, "1.5 PiB"},
			{EiB, "1 EiB"},
			{1536 * PiB, "1.5 EiB"},
			{1264, "1.234 KiB"},  // 1264/1024 = 1.234375
			{1153539, "1.1 MiB"}, // 1.1001 MiB
		} {
			if runtime.GOOS == "windows" {
				testdata.output = strings.ReplaceAll(testdata.output, "iB", "B")
			}
			require.Equal(t, testdata.output, ByteUnit(testdata.input))
		}
	})

	t.Run("internal error", func(t *testing.T) {
		patch := func(string, int) (float64, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(strconv.ParseFloat, patch)
		defer pg.Unpatch()

		defer testDeferForPanic(t)
		ByteUnit(1024)
	})
}

func TestSplitNumber(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  string
		output string
	}{
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
		{"12345", "12,345"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"12345678", "12,345,678"},
		{"123456789", "123,456,789"},
		{"123456789.1", "123,456,789.1"},
		{"123456789.12", "123,456,789.12"},
		{"123456789.123", "123,456,789.123"},
		{"123456789.1234", "123,456,789.1234"},
		{"0.123", "0.123"},
		{"0.1234", "0.1234"},
		{".1234", ".1234"},
		{".12", ".12"},
		{"0.123456", "0.123456"},
		{"123456.789", "123,456.789"},
	} {
		require.Equal(t, testdata.output, SplitNumber(testdata.input))
	}
}

func TestOutputBytes(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  []byte
		output string
	}{
		{[]byte{}, "[]byte{}"},
		{[]byte{1}, `[]byte{0x01}`},
		{[]byte{255, 254}, `[]byte{0xFF, 0xFE}`},
		{[]byte{0, 0, 0, 0, 0, 0, 255, 254},
			`[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE}`},
		{[]byte{0, 0, 0, 0, 0, 0, 255, 254, 1}, `[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01,
}`},
		{[]byte{
			0, 0, 0, 0, 0, 0, 255, 254,
			1, 2, 2, 2, 2, 2, 2, 2,
		}, `[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
}`},
		{[]byte{
			0, 0, 0, 0, 0, 0, 255, 254,
			1, 2, 2, 2, 2, 2, 2, 2,
			4, 5,
		}, `[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	0x04, 0x05,
}`},
	} {
		require.Equal(t, testdata.output, OutputBytes(testdata.input))
	}
}

func TestOutputBytesWithSize(t *testing.T) {
	const expected = `
[]byte{0x00, 0x00, 0x00}`
	b := []byte{0, 0, 0}
	str := OutputBytesWithSize(b, 0)
	require.Equal(t, expected[1:], str)
}
