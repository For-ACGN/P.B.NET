package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestChannelListener(t *testing.T) {
	listener := NewChannelListener()

	mc := testsuite.NewMockConn()
	listener.SendConn(mc)

	conn, err := listener.Accept()
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	t.Log(listener.Addr().Network())
	t.Log(listener.Addr())

	err = listener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, mc)
	testsuite.IsDestroyed(t, listener)
}

func TestChannelListener_Accept(t *testing.T) {
	listener := NewChannelListener()

	listener.SendConn(nil)
	conn, err := listener.Accept()
	require.Error(t, err)
	require.Nil(t, conn)

	err = listener.Close()
	require.NoError(t, err)

	for i := 0; i < 128; i++ {
		listener.SendConn(nil)
	}
	for i := 0; i < 256; i++ {
		conn, err = listener.Accept()
		require.Error(t, err)
		require.Nil(t, conn)
	}

	testsuite.IsDestroyed(t, listener)
}

func TestChannelListener_Testsuite(t *testing.T) {
	listener := NewChannelListener()

	testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
		server, client := net.Pipe()
		listener.SendConn(server)
		return client, nil
	}, true)
}
