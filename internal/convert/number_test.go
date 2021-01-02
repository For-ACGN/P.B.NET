package convert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbsInt64(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  int64
		output int64
	}{
		{-1, 1},
		{0, 0},
		{1, 1},
		{-10, 10},
		{10, 10},
	} {
		require.Equal(t, testdata.output, AbsInt64(testdata.input))
	}
}

func TestSplitNumber(t *testing.T) {
	for _, testdata := range [...]*struct {
		input  string
		output string
	}{
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
		{"12345", "12,345"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"12345678", "12,345,678"},
		{"123456789", "123,456,789"},
		{"123456789.1", "123,456,789.1"},
		{"123456789.12", "123,456,789.12"},
		{"123456789.123", "123,456,789.123"},
		{"123456789.1234", "123,456,789.1234"},
		{"0.123", "0.123"},
		{"0.1234", "0.1234"},
		{".1234", ".1234"},
		{".12", ".12"},
		{"0.123456", "0.123456"},
		{"123456.789", "123,456.789"},
	} {
		require.Equal(t, testdata.output, SplitNumber(testdata.input))
	}
}
