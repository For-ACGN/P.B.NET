package netstat

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestNetstat(t *testing.T) {
	netstat, err := New(nil)
	require.NoError(t, err)

	t.Run("TCP Over IPv4", func(t *testing.T) {
		conns, err := netstat.GetTCP4Conns()
		require.NoError(t, err)
		fmt.Println("Local Address      Remote Address      State      PID      Process")
		for _, conn := range conns {
			fmt.Printf(
				"%s      %s      %s      %d      %s\n",
				conn.LocalAddr(), conn.RemoteAddr(),
				GetTCPConnState(conn.State), conn.PID, conn.Process,
			)
		}
		testsuite.IsDestroyed(t, &conns)
	})

	t.Run("TCP Over IPV6", func(t *testing.T) {
		conns, err := netstat.GetTCP6Conns()
		require.NoError(t, err)
		fmt.Println("Local Address      Remote Address      State      PID      Process")
		for _, conn := range conns {
			fmt.Printf(
				"%s      %s      %s      %d      %s\n",
				conn.LocalAddr(), conn.RemoteAddr(),
				GetTCPConnState(conn.State), conn.PID, conn.Process,
			)
		}
		testsuite.IsDestroyed(t, &conns)
	})

	t.Run("UDP Over IPv4", func(t *testing.T) {
		conns, err := netstat.GetUDP4Conns()
		require.NoError(t, err)
		fmt.Println("Local Address      PID      Process")
		for _, conn := range conns {
			fmt.Printf("%s      %d      %s\n", conn.Addr(), conn.PID, conn.Process)
		}
		testsuite.IsDestroyed(t, &conns)
	})

	t.Run("UDP Over IPV6", func(t *testing.T) {
		conns, err := netstat.GetUDP6Conns()
		require.NoError(t, err)
		fmt.Println("Local Address      PID      Process")
		for _, conn := range conns {
			fmt.Printf("%s      %d      %s\n", conn.Addr(), conn.PID, conn.Process)
		}
		testsuite.IsDestroyed(t, &conns)
	})

	err = netstat.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netstat)
}

func TestTCP4Conn_ID(t *testing.T) {
	conn := TCP4Conn{
		LocalIP:    net.IP{0x01, 0x02, 0x03, 0x04},
		LocalPort:  0x1127,
		RemoteIP:   net.IP{0x05, 0x06, 0x07, 0x08},
		RemotePort: 0x1657,
	}
	id := string([]byte{
		0x01, 0x02, 0x03, 0x04, 0x11, 0x27,
		0x05, 0x06, 0x07, 0x08, 0x16, 0x57,
	})
	require.Equal(t, id, conn.ID())
}

func TestTCP6Conn_ID(t *testing.T) {
	conn := TCP6Conn{
		LocalIP: net.IP{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		},
		LocalScopeID: 0x12341127,
		LocalPort:    0x1657,
		RemoteIP: net.IP{
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		},
		RemoteScopeID: 0x12341657,
		RemotePort:    0x1127,
	}
	id := string([]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x12, 0x34, 0x11, 0x27, 0x16, 0x57,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		0x12, 0x34, 0x16, 0x57, 0x11, 0x27,
	})
	require.Equal(t, id, conn.ID())
}

func TestUDP4Conn_ID(t *testing.T) {
	conn := UDP4Conn{
		LocalIP:   net.IP{0x01, 0x02, 0x03, 0x04},
		LocalPort: 0x1127,
	}
	id := string([]byte{0x01, 0x02, 0x03, 0x04, 0x11, 0x27})
	require.Equal(t, id, conn.ID())
}

func TestUDP6Conn_ID(t *testing.T) {
	conn := UDP6Conn{
		LocalIP: net.IP{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		},
		LocalScopeID: 0x12341127,
		LocalPort:    0x1657,
	}
	id := string([]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x12, 0x34, 0x11, 0x27, 0x16, 0x57,
	})
	require.Equal(t, id, conn.ID())
}
