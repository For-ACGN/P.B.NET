package convert

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	n, err := DumpString(testdata)
	require.NoError(t, err)
	require.Equal(t, defaultStringsLineLen+16+2, n)
}

func TestSdumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	output := SdumpString(testdata)

	fmt.Println(output)
}

func TestFdumpString(t *testing.T) {
	testdata := strings.Repeat("a", defaultStringsLineLen+16)

	n, err := FdumpString(os.Stdout, testdata+"\n")
	require.NoError(t, err)
	require.Equal(t, defaultStringsLineLen+16+2, n)
}

func TestFdumpStringWithPL(t *testing.T) {
	t.Run("common", func(t *testing.T) {

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
}
