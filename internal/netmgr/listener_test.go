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

	guid := tListener.GUID()
	require.False(t, guid.IsZero())
	require.NotZero(t, tListener.Status().Listened)

	testsuite.ListenerAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	}, true)

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

	num := tListener.GetEstConnsNum()
	require.Zero(t, num)
	status := tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)
	require.Zero(t, status.LastAccept)

	server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	})

	num = tListener.GetEstConnsNum()
	require.Equal(t, uint64(1), num)
	status = tListener.Status()
	require.Equal(t, uint64(1), status.EstConns)
	require.NotZero(t, status.LastAccept)

	testsuite.ConnSC(t, server, client, true)

	num = tListener.GetEstConnsNum()
	require.Zero(t, num)
	status = tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)
	require.NotZero(t, status.LastAccept)

	err = tListener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestListener_AcceptEx(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	t.Run("set max conn", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		address := listener.Addr().String()

		tListener := netmgr.TrackListener(listener)

		maxConns := tListener.GetMaxConns()
		require.Zero(t, maxConns)

		tListener.SetMaxConns(1)
		maxConns = tListener.GetMaxConns()
		require.Equal(t, uint64(1), maxConns)

		testsuite.ListenerAndDial(t, tListener, func() (net.Conn, error) {
			return net.Dial("tcp", address)
		}, true)
	})

	t.Run("full", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		address := listener.Addr().String()

		tListener := netmgr.TrackListener(listener)
		tListener.SetMaxConns(1)

		server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
			return net.Dial("tcp", address)
		})

		server.Close()
		client.Close()

	})

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}
