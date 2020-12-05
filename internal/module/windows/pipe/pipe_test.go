// +build windows

package pipe

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"syscall"
	"testing"

	"github.com/Microsoft/go-winio"
	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
)

// create leaks goroutine first in github.com/Microsoft/go-winio.
func init() {
	_, _ = winio.MakeOpenFile(0)
}

const (
	testPipePath = `\\.\pipe\test`
	testMessage  = "hello"
)

func TestPipe(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, err := Listen(testPipePath, nil)
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_, err = conn.Write([]byte(testMessage))
			require.NoError(t, err)
			err = conn.Close()
			require.NoError(t, err)
		}
	}()

	conn, err := Dial(testPipePath, nil)
	require.NoError(t, err)
	buf := make([]byte, len(testMessage))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	require.Equal(t, testMessage, string(buf))
	err = conn.Close()
	require.NoError(t, err)

	err = listener.Close()
	require.NoError(t, err)

	fmt.Println(ErrListenerClosed)
}

func TestListen(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	cfg := Config{MessageMode: true}
	listener, err := Listen(testPipePath, &cfg)
	require.NoError(t, err)

	err = listener.Close()
	require.NoError(t, err)
}

func TestDial(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, err := Listen(testPipePath, nil)
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_, err = conn.Write([]byte(testMessage))
			require.NoError(t, err)
			err = conn.Close()
			require.NoError(t, err)
		}
	}()

	// DialContext
	conn, err := DialContext(context.Background(), testPipePath)
	require.NoError(t, err)
	buf := make([]byte, len(testMessage))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	require.Equal(t, testMessage, string(buf))
	err = conn.Close()
	require.NoError(t, err)

	// DialAccess
	const access = syscall.GENERIC_READ | syscall.GENERIC_WRITE
	conn, err = DialAccess(context.Background(), testPipePath, access)
	require.NoError(t, err)
	buf = make([]byte, len(testMessage))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	require.Equal(t, testMessage, string(buf))
	err = conn.Close()
	require.NoError(t, err)

	err = listener.Close()
	require.NoError(t, err)
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
