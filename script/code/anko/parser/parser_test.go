package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/script/internal/exec"
)

func TestGenerateParser(t *testing.T) {
	// change current directory to anko/parser
	err := os.Chdir("../../../../external/anko/parser")
	require.NoError(t, err)
	// generate parser.go
	output, code, err := exec.Run("goyacc", "-o", "parser.go", "parser.go.y")
	require.NoError(t, err)
	fmt.Println(output)
	require.Equal(t, 0, code)
	// remove output file
	err = os.Remove("y.output")
	require.NoError(t, err)
	// format generated code
	output, code, err = exec.Run("gofmt", "-s", "-w", "parser.go")
	require.NoError(t, err)
	require.Equal(t, 0, code, output)
	// add directive comment about development tools.
	parser, err := os.ReadFile("parser.go")
	require.NoError(t, err)
	// replace code
	str := string(parser)
	for _, item := range [...]*struct {
		target      string
		replacement string
	}{
		{
			"func yyErrorMessage",
			"//gocyclo:ignore\nfunc yyErrorMessage",
		},
		{
			"func (yyrcvr *yyParserImpl) Parse",
			"//gocyclo:ignore\nfunc (yyrcvr *yyParserImpl) Parse",
		},
	} {
		str = strings.ReplaceAll(str, item.target, item.replacement)
	}
	// save code
	err = os.WriteFile("parser.go", []byte(str), 0600)
	require.NoError(t, err)
}
