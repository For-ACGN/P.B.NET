package convert

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// BEInt16ToBytes is used to convert int16 to bytes with big endian.
func BEInt16ToBytes(Int16 int16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(Int16))
	return b
}

// BEInt32ToBytes is used to convert int32 to bytes with big endian.
func BEInt32ToBytes(Int32 int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(Int32))
	return b
}

// BEInt64ToBytes is used to convert int64 to bytes with big endian.
func BEInt64ToBytes(Int64 int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(Int64))
	return b
}

// BEUint16ToBytes is used to convert uint16 to bytes with big endian.
func BEUint16ToBytes(Uint16 uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, Uint16)
	return b
}

// BEUint32ToBytes is used to convert uint32 to bytes with big endian.
func BEUint32ToBytes(Uint32 uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, Uint32)
	return b
}

// BEUint64ToBytes is used to convert uint64 to bytes with big endian.
func BEUint64ToBytes(Uint64 uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, Uint64)
	return b
}

// BEFloat32ToBytes is used to convert float32 to bytes with big endian.
func BEFloat32ToBytes(Float32 float32) []byte {
	b := make([]byte, 4)
	n := *(*uint32)(unsafe.Pointer(&Float32)) // #nosec
	binary.BigEndian.PutUint32(b, n)
	return b
}

// BEFloat64ToBytes is used to convert float64 to bytes with big endian.
func BEFloat64ToBytes(Float64 float64) []byte {
	b := make([]byte, 8)
	n := *(*uint64)(unsafe.Pointer(&Float64)) // #nosec
	binary.BigEndian.PutUint64(b, n)
	return b
}

// BEBytesToInt16 is used to convert bytes to int16 with big endian.
func BEBytesToInt16(Bytes []byte) int16 {
	return int16(binary.BigEndian.Uint16(Bytes))
}

// BEBytesToInt32 is used to convert bytes to int32 with big endian.
func BEBytesToInt32(Bytes []byte) int32 {
	return int32(binary.BigEndian.Uint32(Bytes))
}

// BEBytesToInt64 is used to convert bytes to int64 with big endian.
func BEBytesToInt64(Bytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(Bytes))
}

// BEBytesToUint16 is used to convert bytes to uint16 with big endian.
func BEBytesToUint16(Bytes []byte) uint16 {
	return binary.BigEndian.Uint16(Bytes)
}

// BEBytesToUint32 is used to convert bytes to uint32 with big endian.
func BEBytesToUint32(Bytes []byte) uint32 {
	return binary.BigEndian.Uint32(Bytes)
}

// BEBytesToUint64 is used to convert bytes to uint64 with big endian.
func BEBytesToUint64(Bytes []byte) uint64 {
	return binary.BigEndian.Uint64(Bytes)
}

// BEBytesToFloat32 is used to convert bytes to float32 with big endian.
func BEBytesToFloat32(Bytes []byte) float32 {
	b := binary.BigEndian.Uint32(Bytes)
	return *(*float32)(unsafe.Pointer(&b)) // #nosec
}

// BEBytesToFloat64 is used to convert bytes to float64 with big endian.
func BEBytesToFloat64(Bytes []byte) float64 {
	b := binary.BigEndian.Uint64(Bytes)
	return *(*float64)(unsafe.Pointer(&b)) // #nosec
}

// LEInt16ToBytes is used to convert int16 to bytes with little endian.
func LEInt16ToBytes(Int16 int16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(Int16))
	return b
}

// LEInt32ToBytes is used to convert int32 to bytes with little endian.
func LEInt32ToBytes(Int32 int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(Int32))
	return b
}

// LEInt64ToBytes is used to convert int64 to bytes with little endian.
func LEInt64ToBytes(Int64 int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(Int64))
	return b
}

// LEUint16ToBytes is used to convert uint16 to bytes with little endian.
func LEUint16ToBytes(Uint16 uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, Uint16)
	return b
}

// LEUint32ToBytes is used to convert uint32 to bytes with little endian.
func LEUint32ToBytes(Uint32 uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, Uint32)
	return b
}

// LEUint64ToBytes is used to convert uint64 to bytes with little endian.
func LEUint64ToBytes(Uint64 uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, Uint64)
	return b
}

// LEFloat32ToBytes is used to convert float32 to bytes with little endian.
func LEFloat32ToBytes(Float32 float32) []byte {
	b := make([]byte, 4)
	n := *(*uint32)(unsafe.Pointer(&Float32)) // #nosec
	binary.LittleEndian.PutUint32(b, n)
	return b
}

// LEFloat64ToBytes is used to convert float64 to bytes with little endian.
func LEFloat64ToBytes(Float64 float64) []byte {
	b := make([]byte, 8)
	n := *(*uint64)(unsafe.Pointer(&Float64)) // #nosec
	binary.LittleEndian.PutUint64(b, n)
	return b
}

// LEBytesToInt16 is used to convert bytes to int16 with little endian.
func LEBytesToInt16(Bytes []byte) int16 {
	return int16(binary.LittleEndian.Uint16(Bytes))
}

// LEBytesToInt32 is used to convert bytes to int32 with little endian.
func LEBytesToInt32(Bytes []byte) int32 {
	return int32(binary.LittleEndian.Uint32(Bytes))
}

// LEBytesToInt64 is used to convert bytes to int64 with little endian.
func LEBytesToInt64(Bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(Bytes))
}

// LEBytesToUint16 is used to convert bytes to uint16 with little endian.
func LEBytesToUint16(Bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(Bytes)
}

// LEBytesToUint32 is used to convert bytes to uint32 with little endian.
func LEBytesToUint32(Bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(Bytes)
}

// LEBytesToUint64 is used to convert bytes to uint64 with little endian.
func LEBytesToUint64(Bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(Bytes)
}

// LEBytesToFloat32 is used to convert bytes to float32 with little endian.
func LEBytesToFloat32(Bytes []byte) float32 {
	b := binary.LittleEndian.Uint32(Bytes)
	return *(*float32)(unsafe.Pointer(&b)) // #nosec
}

// LEBytesToFloat64 is used to convert bytes to float64 with little endian.
func LEBytesToFloat64(Bytes []byte) float64 {
	b := binary.LittleEndian.Uint64(Bytes)
	return *(*float64)(unsafe.Pointer(&b)) // #nosec
}

// AbsInt64 is used to calculate the absolute value of the parameter.
func AbsInt64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// storage unit
const (
	Byte uint64 = 1
	KiB         = Byte * 1024
	MiB         = KiB * 1024
	GiB         = MiB * 1024
	TiB         = GiB * 1024
	PiB         = TiB * 1024
	EiB         = PiB * 1024
)

// ByteUnit is used to convert byte to larger unit.
// output unit are KB KiB, MB MiB ...
func ByteUnit(n uint64) string {
	if n < KiB {
		return strconv.Itoa(int(n)) + " Byte"
	}
	var (
		unit string
		div  uint64
	)
	switch {
	case n < MiB:
		unit = "KiB"
		div = KiB
	case n < GiB:
		unit = "MiB"
		div = MiB
	case n < TiB:
		unit = "GiB"
		div = GiB
	case n < PiB:
		unit = "TiB"
		div = TiB
	case n < EiB:
		unit = "PiB"
		div = PiB
	default:
		unit = "EiB"
		div = EiB
	}
	// a history of habit
	if runtime.GOOS == "windows" {
		unit = strings.ReplaceAll(unit, "iB", "B")
	}
	bf := new(big.Float).SetUint64(n)
	bf.Quo(bf, new(big.Float).SetUint64(div))
	// 1.99999999 -> 1.999
	text := bf.Text('G', 64)
	offset := strings.Index(text, ".")
	if offset != -1 && len(text[offset+1:]) > 3 {
		text = text[:offset+1+3]
	}
	// delete zero: 1.100 -> 1.1
	result, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(fmt.Sprintf("convert: internal error: %s", err))
	}
	value := strconv.FormatFloat(result, 'f', -1, 64)
	return value + " " + unit
}

// SplitNumber is used to convert "123456.789" to "123,456.789".
func SplitNumber(str string) string {
	length := len(str)
	if length < 4 {
		return str
	}
	all := strings.SplitN(str, ".", 2)
	allLen := len(all)
	integer := len(all[0])
	if integer < 4 {
		return str
	}
	count := (integer - 1) / 3  // 1234 -> 1,[234]
	offset := integer - 3*count // 1234 > [1],234
	builder := strings.Builder{}
	// write first number
	if offset != 0 {
		builder.WriteString(str[:offset])
	}
	for i := 0; i < count; i++ {
		builder.WriteString(",")
		builder.WriteString(str[offset+i*3 : offset+i*3+3])
	}
	// write float
	if allLen == 2 {
		builder.WriteString(".")
		builder.WriteString(all[1])
	}
	return builder.String()
}

// OutputBytes is used to print byte slice, each line is 8 bytes.
func OutputBytes(b []byte) string {
	return OutputBytesWithSize(b, 8)
}

// OutputBytesWithSize is used to print byte slice.
//
// Output:
// ------one line------
// []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
// -------common-------
// []byte{
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00,
// }
// ------full line-----
// []byte{
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// }
func OutputBytesWithSize(b []byte, line int) string {
	const (
		begin = "[]byte{"
		end   = "}"
	)
	// special: empty data
	l := len(b)
	if l == 0 {
		return begin + end
	}
	// invalid line size
	if line < 1 {
		line = 8
	}
	// create builder
	builder := new(strings.Builder)
	builder.Grow(len(begin+end) + len("0x00, ")*l)
	// write begin string
	builder.WriteString("[]byte{")
	buf := make([]byte, 2)
	// special: one line
	if l <= line {
		for i := 0; i < l; i++ {
			hex.Encode(buf, []byte{b[i]})
			builder.WriteString("0x")
			builder.Write(bytes.ToUpper(buf))
			if i != l-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString("}")
		return builder.String()
	}
	// write begin string
	var counter int // need new line
	builder.WriteString("\n")
	for i := 0; i < l; i++ {
		if counter == 0 {
			builder.WriteString("\t")
		}
		hex.Encode(buf, []byte{b[i]})
		builder.WriteString("0x")
		builder.Write(bytes.ToUpper(buf))
		counter++
		if counter == line {
			builder.WriteString(",\n")
			counter = 0
		} else {
			builder.WriteString(", ")
		}
	}
	// write end string
	if counter != 0 { // delete last space
		return builder.String()[:builder.Len()-1] + "\n}"
	}
	builder.WriteString("}")
	return builder.String()
}
