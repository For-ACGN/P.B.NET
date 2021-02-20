package convert

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageUnit(t *testing.T) {
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
		{1264, "1.23 KiB"},   // 1264/1024 = 1.234375
		{1153539, "1.1 MiB"}, // 1.1001 MiB
	} {
		require.Equal(t, testdata.output, StorageUnit(testdata.input))

		bi := new(big.Int).SetUint64(testdata.input)
		clone := bi.Bytes()
		require.Equal(t, testdata.output, StorageUnitBigInt(bi))
		require.Equal(t, clone, bi.Bytes())
	}
}
