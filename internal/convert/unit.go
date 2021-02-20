package convert

import (
	"math/big"
	"strconv"
)

// storage unit
const (
	Byte uint64 = 1
	KiB         = Byte * 1024
	MiB         = KiB * 1024
	GiB         = MiB * 1024
	TiB         = GiB * 1024
	PiB         = TiB * 1024
	EiB         = PiB * 1024
)

// StorageUnit is used to convert byte to larger unit.
// Output unit are Byte, KiB, MiB, GiB...
func StorageUnit(n uint64) string {
	if n < KiB {
		return strconv.FormatUint(n, 10) + " Byte"
	}
	unit, div := selectStorageUnitAndDiv(n)
	bf := new(big.Float).SetUint64(n)
	bf.Quo(bf, new(big.Float).SetUint64(div))
	value := TruncBigFloat(bf, 2)
	return value + " " + unit
}

// StorageUnitBigInt is used to convert byte to larger unit.
// Output unit are Byte, KiB, MiB, GiB...
func StorageUnitBigInt(n *big.Int) string {
	ui64 := n.Uint64()
	if ui64 < KiB {
		return strconv.FormatUint(ui64, 10) + " Byte"
	}
	unit, div := selectStorageUnitAndDiv(ui64)
	bf := new(big.Float).SetInt(n)
	bf.Quo(bf, new(big.Float).SetUint64(div))
	value := TruncBigFloat(bf, 2)
	return value + " " + unit
}

func selectStorageUnitAndDiv(n uint64) (string, uint64) {
	var (
		unit string
		div  uint64
	)
	switch {
	case n < MiB:
		unit = "KiB"
		div = KiB
	case n < GiB:
		unit = "MiB"
		div = MiB
	case n < TiB:
		unit = "GiB"
		div = GiB
	case n < PiB:
		unit = "TiB"
		div = TiB
	case n < EiB:
		unit = "PiB"
		div = PiB
	default:
		unit = "EiB"
		div = EiB
	}
	return unit, div
}
