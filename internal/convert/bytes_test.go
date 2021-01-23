package convert

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDumpBytes(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  []byte
		output string
	}{
		{[]byte{}, "[]byte{}"},
		{[]byte{1}, `[]byte{0x01,}`},
		{[]byte{255, 254}, `[]byte{0xFF, 0xFE,}`},
		{[]byte{0, 0, 0, 0, 0, 0, 255, 254},
			`[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,}`},
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
		require.Equal(t, testdata.output, SdumpBytes(testdata.input))
	}

	b := []byte{0, 0, 0}
	t.Run("with line length", func(t *testing.T) {
		// str := SdumpBytesWithLineLength(b, 0)
		// require.Equal(t, "[]byte{0x00, 0x00, 0x00}", str)
	})

	t.Run("FdumpBytes", func(t *testing.T) {
		_, err := FdumpBytes(os.Stdout, b)
		require.NoError(t, err)

		fmt.Println()
	})

	t.Run("DumpBytes", func(t *testing.T) {
		DumpBytes(b)
	})
}

func TestMergeBytes(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		a := []byte{1, 2}
		b := []byte{3, 4, 5}

		c := MergeBytes(a, b)
		require.Equal(t, []byte{1, 2, 3, 4, 5}, c)
	})

	t.Run("nil", func(t *testing.T) {
		b := MergeBytes()
		require.Zero(t, b)
	})
}
