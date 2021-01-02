package convert

import (
	"encoding/binary"
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
