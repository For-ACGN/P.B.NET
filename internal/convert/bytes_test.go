package convert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDumpBytes(t *testing.T) {
	testdata := []byte{0, 0, 0}

	n, err := DumpBytes(testdata)
	require.NoError(t, err)

	expected := len("[]byte{}") + 6*len(testdata)
	require.Equal(t, expected, n)
}

func TestSdumpBytes(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  []byte
		output string
	}{
		{
			[]byte{},
			"[]byte{}",
		},
		{
			[]byte{1},
			`[]byte{0x01,}`,
		},
		{
			[]byte{255, 254},
			`[]byte{0xFF, 0xFE,}`,
		},
		{
			[]byte{0, 0, 0, 0, 0, 0, 255, 254},
			`[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,}`,
		},
		{
			[]byte{0, 0, 0, 0, 0, 0, 255, 254, 1},
			`
[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01,
}`[1:],
		},
		{
			[]byte{
				0, 0, 0, 0, 0, 0, 255, 254,
				1, 2, 2, 2, 2, 2, 2, 2,
			},
			`
[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
}`[1:],
		},
		{
			[]byte{
				0, 0, 0, 0, 0, 0, 255, 254,
				1, 2, 2, 2, 2, 2, 2, 2,
				4, 5,
			},
			`
[]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFE,
	0x01, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	0x04, 0x05,
}`[1:],
		},
	} {
		output := SdumpBytes(testdata.input)
		require.Equal(t, testdata.output, output)
	}
}

func TestFdumpBytes(t *testing.T) {
	t.Run("failed to write begin", func(t *testing.T) {

	})

	t.Run("failed to write begin new line", func(t *testing.T) {

	})

	t.Run("failed to write body", func(t *testing.T) {

	})

	t.Run("failed to write end new line", func(t *testing.T) {

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
