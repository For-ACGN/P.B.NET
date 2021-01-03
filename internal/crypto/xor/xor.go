package xor

// XOR is used to xor data with key.
func XOR(data, key []byte) []byte {
	output := make([]byte, len(data))
	Buffer(output, data, key)
	return output
}

// Buffer is use to xor data with key to buffer.
func Buffer(buf, data, key []byte) {
	bufLen := len(buf)
	if bufLen == 0 {
		return
	}
	dataLen := len(data)
	if dataLen != bufLen {
		panic("buffer length is not equal with data")
	}
	keyLen := len(key)
	if keyLen == 0 {
		panic("empty key")
	}
	ki := 0
	for i := 0; i < len(data); i++ {
		buf[i] = data[i] ^ key[ki]
		ki++
		if ki >= keyLen {
			ki = 0
		}
	}
}
