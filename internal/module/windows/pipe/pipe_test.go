// +build windows

package pipe

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
)

const testPipePath = `\\.\pipe\test`

func TestPipe(t *testing.T) {
	fmt.Println(ErrListenerClosed)

	listener, err := Listen(testPipePath, nil)
	require.NoError(t, err)
	defer func() {
		err := listener.Close()
		require.NoError(t, err)
	}()
}

func TestListen(t *testing.T) {

}

func TestConfig(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/config.toml")
	require.NoError(t, err)

	// check unnecessary field
	cfg := new(Config)
	err = toml.Unmarshal(data, cfg)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, cfg)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{"sd", cfg.SecurityDescriptor},
		{true, cfg.MessageMode},
		{int32(1024), cfg.InputBufferSize},
		{int32(2048), cfg.OutputBufferSize},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}
