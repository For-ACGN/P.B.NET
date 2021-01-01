package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func testDial(t *testing.T, local, remote string) net.Conn {
	lAddr, err := net.ResolveTCPAddr("tcp", local)
	require.NoError(t, err)
	rAddr, err := net.ResolveTCPAddr("tcp", remote)
	require.NoError(t, err)
	conn, err := net.DialTCP("tcp", lAddr, rAddr)
	require.NoError(t, err)
	return conn
}

func TestListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("default mode-per host", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode:      FilterModeDefault,
			MaxConnsPerHost: 1,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		t.Log(listener.FilterMode())
		addr := listener.Addr().String()

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)
			conn2, err := listener.Accept()
			require.NoError(t, err)

			conn3, err := listener.Accept()
			require.Error(t, err)
			t.Log(err)
			require.Nil(t, conn3)

			conns := listener.GetConns()
			require.Len(t, conns, 2)

			err = conn1.Close()
			require.NoError(t, err)
			err = conn2.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, conn1)
			testsuite.IsDestroyed(t, conn2)
		}()

		conn1 := testDial(t, "127.0.0.1:0", addr)
		conn2 := testDial(t, "127.0.0.2:0", addr)
		conn3 := testDial(t, "127.0.0.1:0", addr)

		_, err = conn1.Read(make([]byte, 1024))
		require.Error(t, err)
		_, err = conn2.Read(make([]byte, 1024))
		require.Error(t, err)

		err = conn1.Close()
		require.NoError(t, err)
		err = conn2.Close()
		require.NoError(t, err)
		err = conn3.Close()
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("default mode-total", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode:    FilterModeDefault,
			MaxConnsTotal: 2,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		t.Log(listener.FilterMode())
		addr := listener.Addr().String()

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)
			conn2, err := listener.Accept()
			require.NoError(t, err)

			conn3, err := listener.Accept()
			require.Error(t, err)
			t.Log(err)
			require.Nil(t, conn3)

			conns := listener.GetConns()
			require.Len(t, conns, 2)

			err = conn1.Close()
			require.NoError(t, err)
			err = conn2.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, conn1)
			testsuite.IsDestroyed(t, conn2)
		}()

		conn1 := testDial(t, "127.0.0.1:0", addr)
		conn2 := testDial(t, "127.0.0.2:0", addr)
		conn3 := testDial(t, "127.0.0.3:0", addr)

		_, err = conn1.Read(make([]byte, 1024))
		require.Error(t, err)
		_, err = conn2.Read(make([]byte, 1024))
		require.Error(t, err)
		// conn3 is alive but listener can't accept it
		// _, err = conn3.Read(make([]byte, 1024))
		// require.Error(t, err)

		err = conn1.Close()
		require.NoError(t, err)
		err = conn2.Close()
		require.NoError(t, err)
		err = conn3.Close()
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("allow mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode: FilterModeAllow,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		t.Log(listener.FilterMode())
		addr := listener.Addr().String()
		listener.AddAllowedHost("127.0.0.1")

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)

			conns := listener.GetConns()
			require.Len(t, conns, 1)

			err = conn1.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, conn1)

			conn2, err := listener.Accept()
			require.Error(t, err)
			require.Nil(t, conn2)
		}()

		conn1 := testDial(t, "127.0.0.1:0", addr)
		conn2 := testDial(t, "127.0.0.2:0", addr)

		_, err = conn1.Read(make([]byte, 1024))
		require.Error(t, err)
		_, err = conn2.Read(make([]byte, 1024))
		require.Error(t, err)

		err = conn1.Close()
		require.NoError(t, err)
		err = conn2.Close()
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("block mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode: FilterModeBlock,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		t.Log(listener.FilterMode())
		addr := listener.Addr().String()
		listener.AddBlockedHost("127.0.0.2")

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)

			conns := listener.GetConns()
			require.Len(t, conns, 1)

			err = conn1.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, conn1)

			conn2, err := listener.Accept()
			require.Error(t, err)
			require.Nil(t, conn2)
		}()

		conn1 := testDial(t, "127.0.0.1:0", addr)
		conn2 := testDial(t, "127.0.0.2:0", addr)

		_, err = conn1.Read(make([]byte, 1024))
		require.Error(t, err)
		_, err = conn2.Read(make([]byte, 1024))
		require.Error(t, err)

		err = conn1.Close()
		require.NoError(t, err)
		err = conn2.Close()
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, listener)
	})
}

func TestListener_Testsuite(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("default mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		listener, err := NewListener(rawListener, nil)
		require.NoError(t, err)
		addr := listener.Addr().String()

		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			return net.Dial("tcp", addr)
		}, true)
	})

	t.Run("allow mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode: FilterModeAllow,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		addr := listener.Addr().String()
		listener.AddAllowedHost("127.0.0.1")

		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			testDial(t, "127.0.0.2:0", addr)
			testDial(t, "127.0.0.3:0", addr)
			return net.Dial("tcp", addr)
		}, true)
	})

	t.Run("block mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		opts := ListenerOptions{
			FilterMode: FilterModeBlock,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)
		addr := listener.Addr().String()
		listener.AddBlockedHost("127.0.0.2")
		listener.AddBlockedHost("127.0.0.3")

		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			testDial(t, "127.0.0.2:0", addr)
			testDial(t, "127.0.0.3:0", addr)
			return net.Dial("tcp", addr)
		}, true)
	})
}

func TestListener_SetMaxConns(t *testing.T) {
	rawListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener, err := NewListener(rawListener, nil)
	require.NoError(t, err)

	n := listener.GetMaxConnsPerHost()
	require.Equal(t, defaultMaxConnsPerHost, n)
	n = listener.GetMaxConnsTotal()
	require.Equal(t, defaultMaxConnsTotal, n)

	listener.SetMaxConnsPerHost(1000000)
	n = listener.GetMaxConnsPerHost()
	require.Equal(t, defaultMaxConnsTotal, n)

	listener.SetMaxConnsPerHost(0)
	n = listener.GetMaxConnsPerHost()
	require.Equal(t, defaultMaxConnsPerHost, n)

	listener.SetMaxConnsTotal(0)
	n = listener.GetMaxConnsTotal()
	require.Equal(t, defaultMaxConnsTotal, n)

	err = listener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, listener)
}

func TestListener_AllowHost(t *testing.T) {
	const host = "127.0.0.1"

	rawListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	opts := ListenerOptions{
		FilterMode: FilterModeAllow,
	}
	listener, err := NewListener(rawListener, &opts)
	require.NoError(t, err)

	listener.AddAllowedHost(host)
	listener.AddBlockedHost(host)

	list := listener.AllowList()
	require.Equal(t, []string{host}, list)
	list = listener.BlockList()
	require.Zero(t, list)

	allowed := listener.IsAllowedHost(host)
	require.True(t, allowed)
	blocked := listener.IsBlockedHost(host)
	require.False(t, blocked)

	listener.DeleteAllowedHost(host)
	listener.DeleteBlockedHost(host)

	list = listener.AllowList()
	require.Empty(t, list)

	err = listener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, listener)
}

func TestListener_BlockHost(t *testing.T) {
	const host = "127.0.0.1"

	rawListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	opts := ListenerOptions{
		FilterMode: FilterModeBlock,
	}
	listener, err := NewListener(rawListener, &opts)
	require.NoError(t, err)

	listener.AddAllowedHost(host)
	listener.AddBlockedHost(host)

	list := listener.AllowList()
	require.Zero(t, list)
	list = listener.BlockList()
	require.Equal(t, []string{host}, list)

	allowed := listener.IsAllowedHost(host)
	require.False(t, allowed)
	blocked := listener.IsBlockedHost(host)
	require.True(t, blocked)

	listener.DeleteAllowedHost(host)
	listener.DeleteBlockedHost(host)

	list = listener.BlockList()
	require.Empty(t, list)

	err = listener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, listener)
}
