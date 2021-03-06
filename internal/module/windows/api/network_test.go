// +build windows

package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func testPrintTCP4Conns(t *testing.T, conns []*TCP4Conn) {
	fmt.Println("Local Address      Remote Address      State      PID      Process")
	for _, conn := range conns {
		fmt.Printf("%s:%d      %s:%d      %s      %d      %s\n",
			conn.LocalAddr, conn.LocalPort,
			conn.RemoteAddr, conn.RemotePort,
			GetTCPConnState(conn.State), conn.PID, conn.Process,
		)
	}
	testsuite.IsDestroyed(t, &conns)
}

func testPrintTCP6Conns(t *testing.T, conns []*TCP6Conn) {
	fmt.Println("Local Address      Remote Address      State      PID      Process")
	for _, conn := range conns {
		fmt.Printf("[%s%%%d]:%d      [%s%%%d]:%d      %s      %d      %s\n",
			conn.LocalAddr, conn.LocalScopeID, conn.LocalPort,
			conn.RemoteAddr, conn.RemoteScopeID, conn.RemotePort,
			GetTCPConnState(conn.State), conn.PID, conn.Process,
		)
	}
	testsuite.IsDestroyed(t, &conns)
}

func TestGetTCP4Conns(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		t.Run("listeners", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableBasicListener)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("connections", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableBasicConnections)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("all", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableBasicAll)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})
	})

	t.Run("owner pid", func(t *testing.T) {
		t.Run("listeners", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerPIDListener)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("connections", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerPIDConnections)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("all", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerPIDAll)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})
	})

	t.Run("owner module", func(t *testing.T) {
		t.Run("listeners", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerModuleListener)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("connections", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerModuleConnections)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})

		t.Run("all", func(t *testing.T) {
			conns, err := GetTCP4Conns(TCPTableOwnerModuleAll)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP4Conns(t, conns)
		})
	})
}

func TestGetTCP6Conns(t *testing.T) {
	t.Run("owner pid", func(t *testing.T) {
		t.Run("listeners", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerPIDListener)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})

		t.Run("connections", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerPIDConnections)
			require.NoError(t, err)

			if !testsuite.IPv6Enabled {
				return
			}
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})

		t.Run("all", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerPIDAll)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})
	})

	t.Run("owner module", func(t *testing.T) {
		t.Run("listeners", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerModuleListener)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})

		t.Run("connections", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerModuleConnections)
			require.NoError(t, err)

			if !testsuite.IPv6Enabled {
				return
			}
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})

		t.Run("all", func(t *testing.T) {
			conns, err := GetTCP6Conns(TCPTableOwnerModuleAll)
			require.NoError(t, err)
			require.NotEmpty(t, conns)

			testPrintTCP6Conns(t, conns)
		})
	})
}

func testPrintUDP4Conns(t *testing.T, conns []*UDP4Conn) {
	fmt.Println("Local Address      PID      Process")
	for _, conn := range conns {
		fmt.Printf("%s:%d      %d      %s\n",
			conn.LocalAddr, conn.LocalPort,
			conn.PID, conn.Process,
		)
	}
	testsuite.IsDestroyed(t, &conns)
}

func testPrintUDP6Conns(t *testing.T, conns []*UDP6Conn) {
	fmt.Println("Local Address      PID      Process")
	for _, conn := range conns {
		fmt.Printf("[%s%%%d]:%d      %d      %s\n",
			conn.LocalAddr, conn.LocalScopeID, conn.LocalPort,
			conn.PID, conn.Process,
		)
	}
	testsuite.IsDestroyed(t, &conns)
}

func TestGetUDP4Conns(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		conns, err := GetUDP4Conns(UDPTableBasic)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP4Conns(t, conns)
	})

	t.Run("owner pid", func(t *testing.T) {
		conns, err := GetUDP4Conns(UDPTableOwnerPID)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP4Conns(t, conns)
	})

	t.Run("owner module", func(t *testing.T) {
		conns, err := GetUDP4Conns(UDPTableOwnerModule)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP4Conns(t, conns)
	})
}

func TestGetUDP6Conns(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		conns, err := GetUDP6Conns(UDPTableBasic)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP6Conns(t, conns)
	})

	t.Run("owner pid", func(t *testing.T) {
		conns, err := GetUDP6Conns(UDPTableOwnerPID)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP6Conns(t, conns)
	})

	t.Run("owner module", func(t *testing.T) {
		conns, err := GetUDP6Conns(UDPTableOwnerModule)
		require.NoError(t, err)
		require.NotEmpty(t, conns)

		testPrintUDP6Conns(t, conns)
	})
}
