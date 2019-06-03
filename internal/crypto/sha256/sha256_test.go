package sha256

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SHA256(t *testing.T) {
	data := []byte("123456")
	sha256_string := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"
	sha256_bytes := []byte{0x8d, 0x96, 0x9e, 0xef, 0x6e, 0xca, 0xd3, 0xc2, 0x9a, 0x3a,
		0x62, 0x92, 0x80, 0xe6, 0x86, 0xcf, 0xc, 0x3f, 0x5d, 0x5a, 0x86, 0xaf, 0xf3, 0xca,
		0x12, 0x2, 0xc, 0x92, 0x3a, 0xdc, 0x6c, 0x92}
	require.Equal(t, String(data), sha256_string, "invalid sha256 string")
	require.Equal(t, Bytes(data), sha256_bytes, "invalid sha256 bytes")
}
