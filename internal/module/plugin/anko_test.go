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

	script, err := os.ReadFile("testdata/test.ank")
	require.NoError(t, err)

	mod, err := NewAnko(new(testExternal), os.Stdout, string(script))
	require.NoError(t, err)

	err = mod.Start()
	require.NoError(t, err)

	go func() {
		ret, err := mod.Call("")
		fmt.Println(ret, err)
	}()

	go func() {
		ret, err := mod.Call("Scan", "1.1.1.2")
		fmt.Println(ret, err)
	}()

	go func() {
		ret, err := mod.Call("Scan", "1.1.1.1")
		fmt.Println(ret, err)
	}()

	time.Sleep(time.Second)

	fmt.Println(mod.Name())
}
