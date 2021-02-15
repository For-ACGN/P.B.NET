package security

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPasswordStrength(t *testing.T) {
	for _, testdata := range [...]*struct {
		password string
		err      string
	}{
		{"Admin123**142", ""},
		{"193abcAAA**15", ""},

		{"aB1*", errPasswordLen.Error()},

		{"akcAHC/*-/!@#", ruleErrors[containNumber].Error()},
		{"aec123/*-/!@#", ruleErrors[containUpper].Error()},
		{"183APN/*-/!@#", ruleErrors[containLower].Error()},
		{"ach48923AGH48", ruleErrors[containSpecial].Error()},

		{"a1A2C/-*abcde", "find continuous content: \"abcde\""},
		{"a1A2C/-*12345", "find continuous content: \"12345\""},
		{"a1A2C/-*ABCDE", "find continuous content: \"ABCDE\""},
	} {
		err := CheckPasswordStrength([]byte(testdata.password))
		if testdata.err == "" {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, testdata.err)
		}
	}
}
