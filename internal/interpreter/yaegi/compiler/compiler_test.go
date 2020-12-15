package compiler

import (
	"fmt"
	"go/build"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	code, err := Compile(&build.Default, "testdata/pkg")
	require.NoError(t, err)
	fmt.Println(code)
}
