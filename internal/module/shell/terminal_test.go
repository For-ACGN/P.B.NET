package shell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandLineToArgv(t *testing.T) {
	exe1 := "test"
	exe2 := `"test test"`
	testdata := []struct {
		cmd  string
		args []string
	}{
		{"net", []string{"net"}},
		{`net -a -b`, []string{"net", "-a", "-b"}},
		{`net -a -b "a"`, []string{"net", "-a", "-b", "a"}},
		{`"net net"`, []string{"net net"}},
		{`"net\net"`, []string{`net\net`}},
		{`"net\net net"`, []string{`net\net net`}},
		{`net -a \"net`, []string{"net", "-a", `"net`}},
		{`net -a ""`, []string{"net", "-a", ""}},
		{`""net""  -a  -b`, []string{"net", "-a", "-b"}},
		{`"""net""" -a`, []string{`"net"`, "-a"}},
	}
	for i := 0; i < len(testdata); i++ {
		args := CommandLineToArgv(exe1 + " " + testdata[i].cmd)
		require.Equal(t, append([]string{"test"}, testdata[i].args...), args)
	}
	for i := 0; i < len(testdata); i++ {
		args := CommandLineToArgv(exe2 + " " + testdata[i].cmd)
		require.Equal(t, append([]string{"test test"}, testdata[i].args...), args)
	}
}
