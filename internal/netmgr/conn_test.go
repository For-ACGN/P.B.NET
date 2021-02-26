package netmgr

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"project/internal/patch/monkey"
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

	t.Run("no limit", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		n, err := tConn.Read(make([]byte, 64))
		require.NoError(t, err)
		require.Equal(t, 64, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("limit", func(t *testing.T) {
		const limitRate uint64 = 16

		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		rlr := tConn.GetReadLimitRate()
		require.Zero(t, rlr)
		rlr, _ = tConn.GetLimitRate()
		require.Zero(t, rlr)
		status := tConn.Status()
		require.Equal(t, uint64(0), status.ReadLimitRate)
		require.Equal(t, uint64(0), status.Read)
		require.Zero(t, status.LastRead)

		tConn.SetReadLimitRate(limitRate)

		time.Sleep(2 * time.Second)

		now := time.Now()

		n, err := tConn.Read(make([]byte, limitRate*3))
		require.NoError(t, err)
		require.Equal(t, int(limitRate*3), n)

		require.True(t, time.Since(now) > 2*time.Second)

		rlr = tConn.GetReadLimitRate()
		require.Equal(t, limitRate, rlr)
		rlr, _ = tConn.GetLimitRate()
		require.Equal(t, limitRate, rlr)
		status = tConn.Status()
		require.Equal(t, limitRate, status.ReadLimitRate)
		require.Equal(t, limitRate*3, status.Read)
		require.NotZero(t, status.LastRead)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("conn closed", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		tConn.SetReadLimitRate(16)

		err := tConn.Close()
		require.NoError(t, err)

		n, err := tConn.Read(make([]byte, 64))
		require.Equal(t, net.ErrClosed, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("failed to wait", func(t *testing.T) {
		var limiter *rate.Limiter
		patch := func(interface{}, context.Context, int) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(limiter, "WaitN", patch)
		defer pg.Unpatch()

		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		tConn.SetReadLimitRate(16)

		n, err := tConn.Read(make([]byte, 64))
		monkey.IsMonkeyError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("failed to read", func(t *testing.T) {
		conn := testsuite.NewMockConnWithReadError()
		tConn := netmgr.TrackConn(conn)

		n, err := tConn.Read(make([]byte, 64))
		testsuite.IsMockConnReadError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestConn_Write(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	t.Run("no limit", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		n, err := tConn.Write(make([]byte, 64))
		require.NoError(t, err)
		require.Equal(t, 64, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("limit", func(t *testing.T) {
		const limitRate uint64 = 16

		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		wlr := tConn.GetWriteLimitRate()
		require.Zero(t, wlr)
		_, wlr = tConn.GetLimitRate()
		require.Zero(t, wlr)
		status := tConn.Status()
		require.Equal(t, uint64(0), status.WriteLimitRate)
		require.Equal(t, uint64(0), status.Written)
		require.Zero(t, status.LastWrite)

		tConn.SetWriteLimitRate(limitRate)

		time.Sleep(2 * time.Second)

		now := time.Now()

		n, err := tConn.Write(make([]byte, limitRate*3))
		require.NoError(t, err)
		require.Equal(t, int(limitRate*3), n)

		require.True(t, time.Since(now) > 2*time.Second)

		wlr = tConn.GetWriteLimitRate()
		require.Equal(t, limitRate, wlr)
		_, wlr = tConn.GetLimitRate()
		require.Equal(t, limitRate, wlr)
		status = tConn.Status()
		require.Equal(t, limitRate, status.WriteLimitRate)
		require.Equal(t, limitRate*3, status.Written)
		require.NotZero(t, status.LastWrite)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("conn closed", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		tConn.SetWriteLimitRate(16)

		err := tConn.Close()
		require.NoError(t, err)

		n, err := tConn.Write(make([]byte, 64))
		require.Equal(t, net.ErrClosed, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("failed to wait", func(t *testing.T) {
		var limiter *rate.Limiter
		patch := func(interface{}, context.Context, int) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(limiter, "WaitN", patch)
		defer pg.Unpatch()

		conn := testsuite.NewMockConn()
		tConn := netmgr.TrackConn(conn)

		tConn.SetWriteLimitRate(16)

		n, err := tConn.Write(make([]byte, 64))
		monkey.IsMonkeyError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("failed to write", func(t *testing.T) {
		conn := testsuite.NewMockConnWithWriteError()
		tConn := netmgr.TrackConn(conn)

		n, err := tConn.Write(make([]byte, 64))
		testsuite.IsMockConnWriteError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}
