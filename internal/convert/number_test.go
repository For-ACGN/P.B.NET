package convert

import (
	"math/big"
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

func TestTruncFloat(t *testing.T) {
	for _, testdata := range [...]*struct {
		f32    float32
		f64    float64
		output string
	}{
		{1.2345, 1.2345, "1.234"},
		{1.234, 1.234, "1.234"},
		{129.234, 129.234, "129.234"},
		{129.1974, 129.1974, "129.197"},
		{100.1975, 100.1975, "100.197"},
		{128.1975, 128.1975, "128.197"},
		{1.9998, 1.9998, "1.999"},
		{1.9994, 1.9994, "1.999"},
		{1.0001, 1.0001, "1"},
		{1, 1, "1"},
		{0.2345, 0.2345, "0.234"},
		{0.234, 0.234, "0.234"},
		{0.1975, 0.1975, "0.197"},
		{0.9998, 0.9998, "0.999"},
		{0.9994, 0.9994, "0.999"},
		{0.0001, 0.0001, "0"},
		{0, 0, "0"},
	} {
		require.Equal(t, testdata.output, TruncFloat32(testdata.f32, 3))
		require.Equal(t, testdata.output, TruncFloat64(testdata.f64, 3))
		bf := big.NewFloat(testdata.f64)
		require.Equal(t, testdata.output, TruncBigFloat(bf, 3))
	}

	// truncate all decimal
	for _, testdata := range [...]*struct {
		f32    float32
		f64    float64
		output string
	}{
		{1.2345, 1.2345, "1"},
		{1.234, 1.234, "1"},
		{129.234, 129.234, "129"},
		{129.1974, 129.1974, "129"},
		{100.1975, 100.1975, "100"},
		{128.1975, 128.1975, "128"},
		{1.9998, 1.9998, "1"},
		{1.9994, 1.9994, "1"},
		{1.0001, 1.0001, "1"},
		{1, 1, "1"},
		{0.2345, 0.2345, "0"},
		{0.234, 0.234, "0"},
		{0.1975, 0.1975, "0"},
		{0.9998, 0.9998, "0"},
		{0.9994, 0.9994, "0"},
		{0.0001, 0.0001, "0"},
		{0, 0, "0"},
	} {
		require.Equal(t, testdata.output, TruncFloat32(testdata.f32, 0))
		require.Equal(t, testdata.output, TruncFloat64(testdata.f64, 0))
		bf := big.NewFloat(testdata.f64)
		require.Equal(t, testdata.output, TruncBigFloat(bf, 0))
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
