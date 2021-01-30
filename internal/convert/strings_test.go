package convert

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestDumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	DumpString(testdata)
}

func TestSdumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	output := SdumpString(testdata)
	fmt.Println(output)

	expected := strings.Repeat("a", defaultStringsLineLen)
	expected += "\n"
	expected += strings.Repeat("a", 16)
	require.Equal(t, expected, output)
}

func TestFdumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	n, err := FdumpString(os.Stdout, testdata+"\n")
	require.NoError(t, err)
	require.Equal(t, defaultStringsLineLen+16+2, n)
}

func TestFdumpStringWithPL(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			input  string
			output string
		}{
			{"123", "  123"},
			{"1234", "  1234"},
			{"12345", "  1234\n  5"},
			{"12345678", "  1234\n  5678"},
			{"123456789", "  1234\n  5678\n  9"},
		} {
			output := SdumpStringWithPL(testdata.input, "  ", 4)
			require.Equal(t, testdata.output, output)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		n, err := FdumpStringWithPL(nil, "", "", defaultStringsLineLen)
		require.NoError(t, err)
		require.Zero(t, n)
	})

	t.Run("invalid line length", func(t *testing.T) {
		n, err := FdumpStringWithPL(os.Stdout, "a", "", -1)
		require.NoError(t, err)
		require.Equal(t, 1, n)

		// fix goland new line bug
		fmt.Println()
	})

	t.Run("failed to write prefix", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(interface{}, []byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpStringWithPL("12345", "  ", 4)
		require.Empty(t, output)
	})

	t.Run("failed to write buf", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			str := string(b)
			if str == "1234" {
				return 0, monkey.Error
			}
			return builder.WriteString(str)
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpStringWithPL("12345", "  ", 4)
		require.Equal(t, "  ", output)
	})

	t.Run("failed to write new line", func(t *testing.T) {
		var builder *strings.Builder
		patch := func(builder *strings.Builder, b []byte) (int, error) {
			str := string(b)
			if str == "\n" {
				return 0, monkey.Error
			}
			return builder.WriteString(str)
		}
		pg := monkey.PatchInstanceMethod(builder, "Write", patch)
		defer pg.Unpatch()

		output := SdumpStringWithPL("12345", "  ", 4)
		require.Equal(t, "  1234", output)
	})
}
