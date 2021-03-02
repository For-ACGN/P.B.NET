package netmgr

import (
	"context"
	"math"
	"net"
	"sync"
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

	manager := New(nil)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)

	guid := tConn.GUID()
	require.False(t, guid.IsZero())
	require.NotZero(t, tConn.Status().Established)

	err := tConn.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestConn_Read(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("no limit", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

		n, err := tConn.Read(make([]byte, 64))
		testsuite.IsMockConnReadError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestConn_Write(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("no limit", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

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
		tConn := manager.TrackConn(conn)

		n, err := tConn.Write(make([]byte, 64))
		testsuite.IsMockConnWriteError(t, err)
		require.Equal(t, 0, n)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestConn_SetLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)

	const (
		readLimitRate  uint64 = 16
		writeLimitRate uint64 = 8
	)

	rlr, wlr := tConn.GetLimitRate()
	require.Zero(t, rlr)
	require.Zero(t, wlr)
	status := tConn.Status()
	require.Equal(t, uint64(0), status.ReadLimitRate)
	require.Equal(t, uint64(0), status.WriteLimitRate)
	require.Equal(t, uint64(0), status.Read)
	require.Equal(t, uint64(0), status.Written)
	require.Zero(t, status.LastRead)
	require.Zero(t, status.LastWrite)

	tConn.SetLimitRate(readLimitRate, writeLimitRate)

	time.Sleep(2 * time.Second)

	now := time.Now()

	n, err := tConn.Read(make([]byte, readLimitRate*3))
	require.NoError(t, err)
	require.Equal(t, int(readLimitRate*3), n)

	require.True(t, time.Since(now) > 2*time.Second)

	now = time.Now()

	n, err = tConn.Write(make([]byte, writeLimitRate*4))
	require.NoError(t, err)
	require.Equal(t, int(writeLimitRate*4), n)

	require.True(t, time.Since(now) > 3*time.Second)

	rlr, wlr = tConn.GetLimitRate()
	require.Equal(t, readLimitRate, rlr)
	require.Equal(t, writeLimitRate, wlr)
	status = tConn.Status()
	require.Equal(t, readLimitRate, status.ReadLimitRate)
	require.Equal(t, writeLimitRate, status.WriteLimitRate)
	require.Equal(t, readLimitRate*3, status.Read)
	require.Equal(t, writeLimitRate*4, status.Written)
	require.NotZero(t, status.LastRead)
	require.NotZero(t, status.LastWrite)

	err = tConn.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestConn_Close(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("close when read or write", func(t *testing.T) {
		manager := New(nil)

		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		tConn.SetLimitRate(16, 32)

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()

			n, err := tConn.Read(make([]byte, 1024))
			require.Equal(t, net.ErrClosed, err)
			require.Zero(t, n)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			n, err := tConn.Write(make([]byte, 1024))
			require.Equal(t, net.ErrClosed, err)
			require.Zero(t, n)
		}()

		err := tConn.Close()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, tConn)

		err = manager.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, manager)
	})

	t.Run("failed to close inner conn", func(t *testing.T) {
		manager := New(nil)

		conn := testsuite.NewMockConnWithCloseError()
		tConn := manager.TrackConn(conn)

		err := tConn.Close()
		testsuite.IsMockConnCloseError(t, err)

		err = manager.Close()
		testsuite.IsMockConnCloseError(t, err)

		testsuite.IsDestroyed(t, tConn)
		testsuite.IsDestroyed(t, manager)
	})
}

func TestConn_MaxLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("SetLimitRate", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		tConn.SetLimitRate(math.MaxUint64, math.MaxUint64)

		now := time.Now()

		n, err := tConn.Read(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		n, err = tConn.Write(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		require.True(t, time.Since(now) < time.Second)

		err = tConn.Close()
		require.NoError(t, err)
	})

	t.Run("SetReadLimitRate", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		tConn.SetReadLimitRate(math.MaxUint64)

		now := time.Now()

		n, err := tConn.Read(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		require.True(t, time.Since(now) < time.Second)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("SetWriteLimitRate", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		tConn.SetWriteLimitRate(math.MaxUint64)

		now := time.Now()

		n, err := tConn.Write(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		require.True(t, time.Since(now) < time.Second)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("default config", func(t *testing.T) {
		manager.SetConnLimitRate(math.MaxUint64, math.MaxUint64)

		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		now := time.Now()

		n, err := tConn.Read(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		n, err = tConn.Write(make([]byte, 1024))
		require.NoError(t, err)
		require.Equal(t, 1024, n)

		require.True(t, time.Since(now) < time.Second)

		err = tConn.Close()
		require.NoError(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestConn_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			conn := testsuite.NewMockConn()
			tConn := manager.TrackConn(conn)

			read := func() {
				n, err := tConn.Read(make([]byte, 64))
				require.NoError(t, err)
				require.Equal(t, 64, n)
			}
			write := func() {
				n, err := tConn.Write(make([]byte, 32))
				require.NoError(t, err)
				require.Equal(t, 32, n)
			}
			getGUID := func() {
				tConn.GUID()
			}
			getLimitRate := func() {
				tConn.GetLimitRate()
			}
			setLimitRate := func() {
				tConn.SetLimitRate(1024, 2048)
			}
			getReadLimitRate := func() {
				tConn.GetReadLimitRate()
			}
			setReadLimitRate := func() {
				tConn.SetReadLimitRate(1024)
			}
			getWriteLimitRate := func() {
				tConn.GetWriteLimitRate()
			}
			setWriteLimitRate := func() {
				tConn.SetWriteLimitRate(2048)
			}
			status := func() {
				tConn.Status()
			}
			fns := []func(){
				read, write, write,
				getGUID, getLimitRate, setLimitRate,
				getReadLimitRate, setReadLimitRate,
				getWriteLimitRate, setWriteLimitRate,
				status, status,
			}
			testsuite.RunParallel(100, nil, nil, fns...)

			err := tConn.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tConn)
		})

		t.Run("whole", func(t *testing.T) {
			var tConn *Conn

			init := func() {
				conn := testsuite.NewMockConn()
				tConn = manager.TrackConn(conn)
			}
			read := func() {
				n, err := tConn.Read(make([]byte, 64))
				require.NoError(t, err)
				require.Equal(t, 64, n)
			}
			write := func() {
				n, err := tConn.Write(make([]byte, 32))
				require.NoError(t, err)
				require.Equal(t, 32, n)
			}
			getGUID := func() {
				tConn.GUID()
			}
			getLimitRate := func() {
				tConn.GetLimitRate()
			}
			setLimitRate := func() {
				tConn.SetLimitRate(1024, 2048)
			}
			getReadLimitRate := func() {
				tConn.GetReadLimitRate()
			}
			setReadLimitRate := func() {
				tConn.SetReadLimitRate(1024)
			}
			getWriteLimitRate := func() {
				tConn.GetWriteLimitRate()
			}
			setWriteLimitRate := func() {
				tConn.SetWriteLimitRate(2048)
			}
			status := func() {
				tConn.Status()
			}
			cleanup := func() {
				err := tConn.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				read, write, write,
				getGUID, getLimitRate, setLimitRate,
				getReadLimitRate, setReadLimitRate,
				getWriteLimitRate, setWriteLimitRate,
				status, status,
			}
			testsuite.RunParallel(100, init, cleanup, fns...)

			err := tConn.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tConn)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {

		})

		t.Run("whole", func(t *testing.T) {

		})
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}
