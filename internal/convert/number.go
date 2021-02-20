package convert

import (
	"math/big"
	"strconv"
	"strings"
)

// AbsInt64 is used to calculate the absolute value of the parameter.
func AbsInt64(v int64) int64 {
	y := v >> 63
	return (v ^ y) - y
}

// TruncFloat32 is used to truncate float32 after decimal point.
func TruncFloat32(f32 float32, n int) string {
	text := strconv.FormatFloat(float64(f32), 'f', -1, 32)
	return truncFloat(text, n)
}

// TruncFloat64 is used to truncate float64 after decimal point.
func TruncFloat64(f64 float64, n int) string {
	text := strconv.FormatFloat(f64, 'f', -1, 64)
	return truncFloat(text, n)
}

// TruncBigFloat is used to truncate big.Float after decimal point.
func TruncBigFloat(bf *big.Float, n int) string {
	f64, _ := bf.Float64()
	text := strconv.FormatFloat(f64, 'f', -1, 64)
	return truncFloat(text, n)
}

func truncFloat(f string, n int) string {
	// has decimal point
	offset := strings.Index(f, ".")
	if offset == -1 {
		return f
	}
	// truncate all decimal point
	if n <= 0 {
		return f[:offset]
	}
	// truncate long: 1.9999999 -> 1.999
	if len(f[offset+1:]) > n {
		f = f[:offset+1+n]
	}
	// delete zero: 1.100 -> 1.1, 1.000 -> 1
	var end int
	for i := len(f) - 1; i > offset; i-- {
		if f[i] == '0' {
			end++
		} else {
			break
		}
	}
	// 1. -> 1
	f = f[:len(f)-end]
	if f[len(f)-1] == '.' {
		f = f[:len(f)-1]
	}
	return f
}

// SplitNumber is used to convert "123456.789" to "123,456.789".
func SplitNumber(str string) string {
	length := len(str)
	if length < 4 {
		return str
	}
	all := strings.SplitN(str, ".", 2)
	allLen := len(all)
	integer := len(all[0])
	if integer < 4 {
		return str
	}
	count := (integer - 1) / 3  // 1234 -> 1,[234]
	offset := integer - 3*count // 1234 > [1],234
	builder := strings.Builder{}
	// write first number
	if offset != 0 {
		builder.WriteString(str[:offset])
	}
	for i := 0; i < count; i++ {
		builder.WriteString(",")
		builder.WriteString(str[offset+i*3 : offset+i*3+3])
	}
	// write float
	if allLen == 2 {
		builder.WriteString(".")
		builder.WriteString(all[1])
	}
	return builder.String()
}
