package plugin

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

type testExternal struct{}

func (testExternal) SendMessage(msg string) {
	fmt.Println(msg)
}

func TestAnko(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	script, err := os.ReadFile("testdata/example.ank")
	require.NoError(t, err)

	mod, err := NewAnko(new(testExternal), os.Stdout, string(script))
	require.NoError(t, err)

	fmt.Println(mod.Name())
}
