package netmgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	conn := testsuite.NewMockConn()
	tConn := netmgr.TrackConn(conn)

	guid := tConn.GUID()
	require.False(t, guid.IsZero())
	require.NotZero(t, tConn.Status().Established)

	err := tConn.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestConn_Read(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	t.Run("limit", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		rate := tConn.GetReadLimitRate()
		require.Zero(t, rate)
		rate, _ = tConn.GetLimitRate()
		require.Zero(t, rate)
		status := tConn.Status()
		require.Equal(t, uint64(0), status.ReadLimitRate)
		require.Equal(t, uint64(0), status.Read)
		require.Zero(t, status.LastRead)

		tConn.SetReadLimitRate(16)

		time.Sleep(4 * time.Second)

		now := time.Now()

		n, err := tConn.Read(make([]byte, 64))
		require.NoError(t, err)
		require.Equal(t, 64, n)

		require.True(t, time.Since(now) > 2*time.Second)

		rate = tConn.GetReadLimitRate()
		require.Equal(t, uint64(16), rate)
		rate, _ = tConn.GetLimitRate()
		require.Equal(t, uint64(16), rate)
		status = tConn.Status()
		require.Equal(t, uint64(16), status.ReadLimitRate)
		require.Equal(t, uint64(48), status.Read)
		require.NotZero(t, status.LastRead)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("failed to wait", func(t *testing.T) {

	})

	t.Run("failed to read", func(t *testing.T) {

	})

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestConn_Write(t *testing.T) {

}
