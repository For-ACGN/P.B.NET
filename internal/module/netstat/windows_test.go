// +build windows

package netstat

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/module/windows/api"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestNetstat_GetTCP4Conns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netstat, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		conns, err := netstat.GetTCP4Conns()
		require.NoError(t, err)
		require.NotEmpty(t, conns)
	})

	t.Run("fail", func(t *testing.T) {
		patch := func(uint32) ([]*api.TCP4Conn, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(api.GetTCP4Conns, patch)
		defer pg.Unpatch()

		conns, err := netstat.GetTCP4Conns()
		require.Error(t, err)
		require.Empty(t, conns)
	})

	err = netstat.Close()
	require.NoError(t, err)
}

func TestNetstat_GetTCP6Conns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netstat, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		conns, err := netstat.GetTCP6Conns()
		require.NoError(t, err)
		require.NotEmpty(t, conns)
	})

	t.Run("fail", func(t *testing.T) {
		patch := func(uint32) ([]*api.TCP6Conn, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(api.GetTCP6Conns, patch)
		defer pg.Unpatch()

		conns, err := netstat.GetTCP6Conns()
		require.Error(t, err)
		require.Empty(t, conns)
	})

	err = netstat.Close()
	require.NoError(t, err)
}

func TestNetstat_GetUDP4Conns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netstat, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		conns, err := netstat.GetUDP4Conns()
		require.NoError(t, err)
		require.NotEmpty(t, conns)
	})

	t.Run("fail", func(t *testing.T) {
		patch := func(uint32) ([]*api.UDP4Conn, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(api.GetUDP4Conns, patch)
		defer pg.Unpatch()

		conns, err := netstat.GetUDP4Conns()
		require.Error(t, err)
		require.Empty(t, conns)
	})

	err = netstat.Close()
	require.NoError(t, err)
}

func TestNetstat_GetUDP6Conns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netstat, err := New(nil)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		conns, err := netstat.GetUDP6Conns()
		require.NoError(t, err)
		require.NotEmpty(t, conns)
	})

	t.Run("fail", func(t *testing.T) {
		patch := func(uint32) ([]*api.UDP6Conn, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(api.GetUDP6Conns, patch)
		defer pg.Unpatch()

		conns, err := netstat.GetUDP6Conns()
		require.Error(t, err)
		require.Empty(t, conns)
	})

	err = netstat.Close()
	require.NoError(t, err)
}
