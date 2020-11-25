package plugin

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"

	_ "project/internal/anko/goroot"
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

	err = mod.Start()
	require.NoError(t, err)

	err = mod.Call("")
	fmt.Println(err)

	err = mod.Call("Scan")
	fmt.Println(err)

	err = mod.Call("Scan", "1.1.1.1")
	fmt.Println(err)

	time.Sleep(time.Second)

	fmt.Println(mod.Name())
}
