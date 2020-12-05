package kcp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

var (
	testPassword = []byte("test password")
	testSalt     = []byte("test salt")
)

func TestListenAndDial(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, err := Listen("localhost:0", testPassword, testSalt)
	require.NoError(t, err)
	address := listener.Addr().String()

	testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
		return Dial(address, testPassword, testSalt)
	}, true)
}
