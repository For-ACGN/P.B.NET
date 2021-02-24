package netmgr

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		err := listener.Close()
		require.Error(t, err)
	}()
	address := listener.Addr().String()

	tListener := netmgr.TrackListener(listener)

	listeners := netmgr.Listeners()
	require.Len(t, listeners, 1)

	testsuite.ListenerAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	}, true)

	require.Empty(t, netmgr.Listeners())

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestListener_Accept(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		err := listener.Close()
		require.Error(t, err)
	}()
	address := listener.Addr().String()

	tListener := netmgr.TrackListener(listener)

	listeners := netmgr.Listeners()
	require.Len(t, listeners, 1)

	num := tListener.GetEstConnsNum()
	require.Zero(t, num)
	status := tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)

	server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	})

	num = tListener.GetEstConnsNum()
	require.Equal(t, uint64(1), num)
	status = tListener.Status()
	require.Equal(t, uint64(1), status.EstConns)

	testsuite.ConnSC(t, server, client, true)

	num = tListener.GetEstConnsNum()
	require.Zero(t, num)
	status = tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)

	err = tListener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener)

	require.Empty(t, netmgr.Listeners())

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}
