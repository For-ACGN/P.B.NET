package convert

import (
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
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
// output unit are KB KiB, MB MiB ...
func StorageUnit(n uint64) string {
	if n < KiB {
		return strconv.Itoa(int(n)) + " Byte"
	}
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
	// history of habit
	if runtime.GOOS == "windows" {
		unit = strings.ReplaceAll(unit, "iB", "B")
	}
	bf := new(big.Float).SetUint64(n)
	bf.Quo(bf, new(big.Float).SetUint64(div))
	// 1.99999999 -> 1.999
	text := bf.Text('G', 64)
	offset := strings.Index(text, ".")
	if offset != -1 && len(text[offset+1:]) > 3 {
		text = text[:offset+1+3]
	}
	// delete zero: 1.100 -> 1.1
	result, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(fmt.Sprintf("convert: internal error: %s", err))
	}
	value := strconv.FormatFloat(result, 'f', -1, 64)
	return value + " " + unit
}
