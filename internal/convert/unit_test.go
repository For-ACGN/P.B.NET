package convert

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestStorageUnit(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			input  uint64
			output string
		}{
			{1023 * Byte, "1023 Byte"},
			{1024 * Byte, "1 KiB"},
			{1536 * Byte, "1.5 KiB"},
			{MiB, "1 MiB"},
			{1536 * KiB, "1.5 MiB"},
			{GiB, "1 GiB"},
			{1536 * MiB, "1.5 GiB"},
			{TiB, "1 TiB"},
			{1536 * GiB, "1.5 TiB"},
			{PiB, "1 PiB"},
			{1536 * TiB, "1.5 PiB"},
			{EiB, "1 EiB"},
			{1536 * PiB, "1.5 EiB"},
			{1264, "1.234 KiB"},  // 1264/1024 = 1.234375
			{1153539, "1.1 MiB"}, // 1.1001 MiB
		} {
			if runtime.GOOS == "windows" {
				testdata.output = strings.ReplaceAll(testdata.output, "iB", "B")
			}
			require.Equal(t, testdata.output, StorageUnit(testdata.input))
		}
	})

	t.Run("internal error", func(t *testing.T) {
		patch := func(string, int) (float64, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(strconv.ParseFloat, patch)
		defer pg.Unpatch()

		defer testDeferForPanic(t)
		StorageUnit(1024)
	})
}
