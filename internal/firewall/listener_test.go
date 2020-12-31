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
		addr := listener.Addr().String()
		listener.AddAllowedHost("127.0.0.1")

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)

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
		addr := listener.Addr().String()
		listener.AddBlockedHost("127.0.0.2")

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)

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
