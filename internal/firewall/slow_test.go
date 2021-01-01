package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestSlowListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	rawListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener := NewSlowListener(rawListener, 1, 2)
	addr := listener.Addr().String()

	testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
		return net.Dial("tcp", addr)
	}, true)
}

func TestSlowListener_Accept(t *testing.T) {
	rawListener := testsuite.NewMockListenerWithAcceptError()
	listener := NewSlowListener(rawListener, 1, 2)

	conn, err := listener.Accept()
	require.Error(t, err)
	require.Nil(t, conn)

	err = listener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, listener)
}
