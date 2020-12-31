package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("default mode", func(t *testing.T) {
		rawListener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := rawListener.Addr().String()
		opts := ListenerOptions{
			Mode:            ListenerModeDefault,
			MaxConnsPerHost: 1,
		}
		listener, err := NewListener(rawListener, &opts)
		require.NoError(t, err)

		go func() {
			conn1, err := listener.Accept()
			require.NoError(t, err)

			conn2, err := listener.Accept()
			require.Error(t, err)
			require.Nil(t, conn2)

			err = conn1.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, conn1)
		}()

		conn1, err := net.Dial("tcp", addr)
		require.NoError(t, err)
		conn2, err := net.Dial("tcp", addr)
		require.NoError(t, err)
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

	t.Run("allow mode", func(t *testing.T) {

	})

	t.Run("block mode", func(t *testing.T) {

	})
}
