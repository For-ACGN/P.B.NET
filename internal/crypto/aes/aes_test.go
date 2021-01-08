package aes

import (
	"bytes"
)

var (
	test128BitKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16}
	test192BitKey = append(test128BitKey, []byte{17, 18, 19, 20, 21, 22, 23, 24}...)
	test256BitKey = bytes.Repeat(test128BitKey, 2)
)

func generateBytes() []byte {
	testdata := make([]byte, 63)
	for i := 0; i < 63; i++ {
		testdata[i] = byte(i)
	}
	return testdata
}
