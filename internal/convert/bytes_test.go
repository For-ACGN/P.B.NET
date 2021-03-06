package convert

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestDumpBytes(t *testing.T) {
	testdata := []byte{0, 0, 0}

	DumpBytes(testdata)
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
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			str := string(b)
			if str == "[]byte{" {
				return 0, monkey.Error
			}
			return builder.WriteString(str)
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytes(nil)
		require.Empty(t, output)
	})

	t.Run("failed to write begin new line", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			if bytes.Equal(b, newLine) {
				return 0, monkey.Error
			}
			return builder.WriteString(string(b))
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytes(make([]byte, defaultBytesLineLen+1))
		require.Equal(t, "[]byte{", output)
	})

	t.Run("failed to write body", func(t *testing.T) {
		patch := func(io.Writer, []byte, string, int) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(FdumpBytesWithPL, patch)
		defer pg.Unpatch()

		output := SdumpBytes(make([]byte, defaultBytesLineLen+1))
		require.Equal(t, "[]byte{\n", output)
	})

	t.Run("failed to write end new line", func(t *testing.T) {
		var (
			builder *strings.Builder
			n       int
		)
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			if bytes.Equal(b, newLine) {
				if n > 1 {
					return 0, monkey.Error
				}
				n++
			}
			return builder.WriteString(string(b))
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytes(make([]byte, defaultBytesLineLen+1))
		expected := "[]byte{\n\t0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,\n\t0x00,"
		require.Equal(t, expected, output)
	})
}

func TestDumpBytesWithPL(t *testing.T) {
	testdata := []byte{1, 2, 3, 4, 5}

	DumpBytesWithPL(testdata, "\t", 4)
}

func TestSdumpBytesWithPL(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  []byte
		output string
	}{
		{[]byte{}, ""},
		{[]byte{1}, "\t0x01,"},
		{[]byte{1, 2, 3, 4}, "\t0x01, 0x02, 0x03, 0x04,"},
		{[]byte{1, 2, 3, 4, 5}, "\t0x01, 0x02, 0x03, 0x04,\n\t0x05,"},
	} {
		output := SdumpBytesWithPL(testdata.input, "\t", 4)
		require.Equal(t, testdata.output, output)
	}
}

func TestFdumpBytesWithPL(t *testing.T) {
	t.Run("invalid line length", func(t *testing.T) {
		output := SdumpBytesWithPL([]byte{1, 2}, "", -1)
		expected := "0x01, 0x02,"
		require.Equal(t, expected, output)
	})

	t.Run("failed to write prefix", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			str := string(b)
			if str == "\t" {
				return 0, monkey.Error
			}
			return builder.WriteString(str)
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytesWithPL([]byte{1, 2}, "\t", 2)
		require.Zero(t, output)
	})

	t.Run("failed to write hexed byte", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			str := string(b)
			if str == "0x02, " {
				return 0, monkey.Error
			}
			return builder.WriteString(str)
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytesWithPL([]byte{1, 2, 3}, "\t", 4)
		require.Equal(t, "\t0x01, ", output)
	})

	t.Run("failed to write new line", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			if bytes.Equal(b, newLine) {
				return 0, monkey.Error
			}
			return builder.WriteString(string(b))
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpBytesWithPL([]byte{1, 2, 3}, "\t", 2)
		require.Equal(t, "\t0x01, 0x02,", output)
	})
}

func TestMergeBytes(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		a := []byte{1, 2}
		b := []byte{3, 4, 5}
		c := []byte{1, 2, 3, 4, 5}

		require.Equal(t, c, MergeBytes(a, b))
	})

	t.Run("empty", func(t *testing.T) {
		b := MergeBytes()
		require.Zero(t, b)
	})
}
