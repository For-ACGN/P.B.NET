package convert

import (
	"strings"
)

// AbsInt64 is used to calculate the absolute value of the parameter.
func AbsInt64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
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
