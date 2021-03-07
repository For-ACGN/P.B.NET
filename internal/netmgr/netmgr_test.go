package netmgr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/guid"
	"project/internal/testsuite"
)

func TestManager_TrackListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	listener := testsuite.NewMockListener()
	tListener := manager.TrackListener(listener)

	listeners := manager.Listeners()
	require.Len(t, listeners, 1)
	require.Equal(t, tListener, listeners[tListener.GUID()])

	err := tListener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener)

	listeners = manager.Listeners()
	require.Empty(t, listeners)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_TrackConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)

	conns := manager.Conns()
	require.Len(t, conns, 1)
	require.Equal(t, tConn, conns[tConn.GUID()])

	err := tConn.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn)

	conns = manager.Conns()
	require.Empty(t, conns)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetListenerMaxConnsByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		listener := testsuite.NewMockListener()
		tListener := manager.TrackListener(listener)
		tListener.SetMaxConns(1000)
		g := tListener.GUID()

		maxConns, err := manager.GetListenerMaxConnsByGUID(&g)
		require.NoError(t, err)
		require.Equal(t, uint64(1000), maxConns)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		maxConns, err := manager.GetListenerMaxConnsByGUID(&g)
		require.Error(t, err)
		require.Zero(t, maxConns)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_SetListenerMaxConnsByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		listener := testsuite.NewMockListener()
		tListener := manager.TrackListener(listener)
		g := tListener.GUID()

		err := manager.SetListenerMaxConnsByGUID(&g, 1000)
		require.NoError(t, err)

		maxConns := tListener.GetMaxConns()
		require.Equal(t, uint64(1000), maxConns)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.SetListenerMaxConnsByGUID(&g, 1000)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetListenerEstConnsNumByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		listener := testsuite.NewMockListener()
		tListener := manager.TrackListener(listener)
		g := tListener.GUID()

		conn, err := tListener.Accept()
		require.NoError(t, err)

		num, err := manager.GetListenerEstConnsNumByGUID(&g)
		require.NoError(t, err)
		require.Equal(t, uint64(1), num)

		num = tListener.GetEstConnsNum()
		require.Equal(t, uint64(1), num)

		err = conn.Close()
		require.NoError(t, err)

		num, err = manager.GetListenerEstConnsNumByGUID(&g)
		require.NoError(t, err)
		require.Zero(t, num)

		num = tListener.GetEstConnsNum()
		require.Zero(t, num)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		num, err := manager.GetListenerEstConnsNumByGUID(&g)
		require.Error(t, err)
		require.Zero(t, num)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		tConn.SetLimitRate(1000, 2000)
		g := tConn.GUID()

		read, write, err := manager.GetConnLimitRateByGUID(&g)
		require.NoError(t, err)
		require.Equal(t, uint64(1000), read)
		require.Equal(t, uint64(2000), write)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		read, write, err := manager.GetConnLimitRateByGUID(&g)
		require.Error(t, err)
		require.Zero(t, read)
		require.Zero(t, write)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_SetConnLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		g := tConn.GUID()

		err := manager.SetConnLimitRateByGUID(&g, 1000, 2000)
		require.NoError(t, err)

		read, write := tConn.GetLimitRate()
		require.Equal(t, uint64(1000), read)
		require.Equal(t, uint64(2000), write)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.SetConnLimitRateByGUID(&g, 1000, 2000)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnReadLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		tConn.SetReadLimitRate(1000)
		g := tConn.GUID()

		read, err := manager.GetConnReadLimitRateByGUID(&g)
		require.NoError(t, err)
		require.Equal(t, uint64(1000), read)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		read, err := manager.GetConnReadLimitRateByGUID(&g)
		require.Error(t, err)
		require.Zero(t, read)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_SetConnReadLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		g := tConn.GUID()

		err := manager.SetConnReadLimitRateByGUID(&g, 1000)
		require.NoError(t, err)

		read := tConn.GetReadLimitRate()
		require.Equal(t, uint64(1000), read)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.SetConnReadLimitRateByGUID(&g, 1000)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnWriteLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		tConn.SetWriteLimitRate(1000)
		g := tConn.GUID()

		write, err := manager.GetConnWriteLimitRateByGUID(&g)
		require.NoError(t, err)
		require.Equal(t, uint64(1000), write)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		write, err := manager.GetConnWriteLimitRateByGUID(&g)
		require.Error(t, err)
		require.Zero(t, write)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_SetConnWriteLimitRateByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		g := tConn.GUID()

		err := manager.SetConnWriteLimitRateByGUID(&g, 1000)
		require.NoError(t, err)

		write := tConn.GetWriteLimitRate()
		require.Equal(t, uint64(1000), write)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.SetConnWriteLimitRateByGUID(&g, 1000)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetListenerStatusByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		listener := testsuite.NewMockListener()
		tListener := manager.TrackListener(listener)
		status := tListener.Status()
		g := tListener.GUID()

		s, err := manager.GetListenerStatusByGUID(&g)
		require.NoError(t, err)

		require.Equal(t, status, s)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		status, err := manager.GetListenerStatusByGUID(&g)
		require.Error(t, err)
		require.Nil(t, status)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnStatusByGUID(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)
		status := tConn.Status()
		g := tConn.GUID()

		s, err := manager.GetConnStatusByGUID(&g)
		require.NoError(t, err)

		require.Equal(t, status, s)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		status, err := manager.GetConnStatusByGUID(&g)
		require.Error(t, err)
		require.Nil(t, status)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_CloseListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		listener := testsuite.NewMockListener()
		tListener := manager.TrackListener(listener)

		listeners := manager.Listeners()
		require.Len(t, listeners, 1)

		g := tListener.GUID()
		err := manager.CloseListener(&g)
		require.NoError(t, err)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)

		listeners = manager.Listeners()
		require.Empty(t, listeners)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.CloseListener(&g)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_CloseConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("common", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		tConn := manager.TrackConn(conn)

		conns := manager.Conns()
		require.Len(t, conns, 1)

		g := tConn.GUID()
		err := manager.CloseConn(&g)
		require.NoError(t, err)

		err = tConn.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn)

		conns = manager.Conns()
		require.Empty(t, conns)
	})

	t.Run("not exist", func(t *testing.T) {
		g := guid.GUID{}

		err := manager.CloseConn(&g)
		require.Error(t, err)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetListenersNum(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	num := manager.GetListenersNum()
	require.Zero(t, num)

	listener := testsuite.NewMockListener()
	tListener := manager.TrackListener(listener)

	num = manager.GetListenersNum()
	require.Equal(t, 1, num)

	err := tListener.Close()
	require.NoError(t, err)

	num = manager.GetListenersNum()
	require.Zero(t, num)

	testsuite.IsDestroyed(t, tListener)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnsNum(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	num := manager.GetConnsNum()
	require.Zero(t, num)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)

	num = manager.GetConnsNum()
	require.Equal(t, 1, num)

	err := tConn.Close()
	require.NoError(t, err)

	num = manager.GetConnsNum()
	require.Zero(t, num)

	testsuite.IsDestroyed(t, tConn)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetAllListenersStatus(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	allStatus := manager.GetAllListenersStatus()
	require.Empty(t, allStatus)

	listener1 := testsuite.NewMockListener()
	tListener1 := manager.TrackListener(listener1)
	status1 := tListener1.Status()

	listener2 := testsuite.NewMockListener()
	tListener2 := manager.TrackListener(listener2)
	status2 := tListener2.Status()

	allStatus = manager.GetAllListenersStatus()
	require.Len(t, allStatus, 2)
	require.Equal(t, allStatus[tListener1.GUID()], status1)
	require.Equal(t, allStatus[tListener2.GUID()], status2)

	err := tListener1.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener1)

	err = tListener2.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener2)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetAllConnsStatus(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	allStatus := manager.GetAllConnsStatus()
	require.Empty(t, allStatus)

	conn1 := testsuite.NewMockConn()
	tConn1 := manager.TrackConn(conn1)
	status1 := tConn1.Status()

	conn2 := testsuite.NewMockConn()
	tConn2 := manager.TrackConn(conn2)
	status2 := tConn2.Status()

	allStatus = manager.GetAllConnsStatus()
	require.Len(t, allStatus, 2)
	require.Equal(t, allStatus[tConn1.GUID()], status1)
	require.Equal(t, allStatus[tConn2.GUID()], status2)

	err := tConn1.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn1)

	err = tConn2.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn2)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_CloseAllListeners(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		manager := New(nil)

		listener1 := testsuite.NewMockListener()
		tListener1 := manager.TrackListener(listener1)

		listener2 := testsuite.NewMockListener()
		tListener2 := manager.TrackListener(listener2)

		num := manager.GetListenersNum()
		require.Equal(t, 2, num)

		err := manager.CloseAllListeners()
		require.NoError(t, err)

		num = manager.GetListenersNum()
		require.Zero(t, num)

		err = tListener1.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener1)

		err = tListener2.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener2)

		err = manager.Close()
		require.NoError(t, err)
	})

	t.Run("close with error", func(t *testing.T) {
		manager := New(nil)

		listener1 := testsuite.NewMockListenerWithCloseError()
		tListener1 := manager.TrackListener(listener1)

		listener2 := testsuite.NewMockListenerWithCloseError()
		tListener2 := manager.TrackListener(listener2)

		num := manager.GetListenersNum()
		require.Equal(t, 2, num)

		err := manager.CloseAllListeners()
		testsuite.IsMockListenerCloseError(t, err)

		num = manager.GetListenersNum()
		require.Equal(t, 2, num)

		err = tListener1.Close()
		testsuite.IsMockListenerCloseError(t, err)

		err = tListener2.Close()
		testsuite.IsMockListenerCloseError(t, err)

		err = manager.Close()
		testsuite.IsMockListenerCloseError(t, err)

		testsuite.IsDestroyed(t, manager)
	})
}

func TestManager_CloseAllConns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("common", func(t *testing.T) {
		manager := New(nil)

		conn1 := testsuite.NewMockConn()
		tConn1 := manager.TrackConn(conn1)

		conn2 := testsuite.NewMockConn()
		tConn2 := manager.TrackConn(conn2)

		num := manager.GetConnsNum()
		require.Equal(t, 2, num)

		err := manager.CloseAllConns()
		require.NoError(t, err)

		num = manager.GetConnsNum()
		require.Zero(t, num)

		err = tConn1.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn1)

		err = tConn2.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tConn2)

		err = manager.Close()
		require.NoError(t, err)
	})

	t.Run("close with error", func(t *testing.T) {
		manager := New(nil)

		conn1 := testsuite.NewMockConnWithCloseError()
		tConn1 := manager.TrackConn(conn1)

		conn2 := testsuite.NewMockConnWithCloseError()
		tConn2 := manager.TrackConn(conn2)

		num := manager.GetConnsNum()
		require.Equal(t, 2, num)

		err := manager.CloseAllConns()
		testsuite.IsMockConnCloseError(t, err)

		num = manager.GetConnsNum()
		require.Equal(t, 2, num)

		err = tConn1.Close()
		testsuite.IsMockConnCloseError(t, err)

		err = tConn2.Close()
		testsuite.IsMockConnCloseError(t, err)

		err = manager.Close()
		testsuite.IsMockConnCloseError(t, err)

		testsuite.IsDestroyed(t, manager)
	})
}

func TestManager_GetListenerMaxConns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	maxConns := manager.GetListenerMaxConns()
	require.Zero(t, maxConns)

	manager.SetListenerMaxConns(1000)

	maxConns = manager.GetListenerMaxConns()
	require.Equal(t, uint64(1000), maxConns)

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	read, write := manager.GetConnLimitRate()
	require.Zero(t, read)
	require.Zero(t, write)

	manager.SetConnLimitRate(1000, 2000)

	read, write = manager.GetConnLimitRate()
	require.Equal(t, uint64(1000), read)
	require.Equal(t, uint64(2000), write)

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnReadLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	read := manager.GetConnReadLimitRate()
	require.Zero(t, read)

	manager.SetConnReadLimitRate(1000)

	read = manager.GetConnReadLimitRate()
	require.Equal(t, uint64(1000), read)

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_GetConnWriteLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	write := manager.GetConnWriteLimitRate()
	require.Zero(t, write)

	manager.SetConnWriteLimitRate(1000)

	write = manager.GetConnWriteLimitRate()
	require.Equal(t, uint64(1000), write)

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_Close(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	err := manager.Close()
	require.NoError(t, err)

	listener := testsuite.NewMockListener()
	tListener := manager.TrackListener(listener)
	c, err := tListener.Accept()
	testsuite.IsMockListenerClosedError(t, err)
	require.Nil(t, c)

	testsuite.IsDestroyed(t, tListener)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)
	n, err := tConn.Write(nil)
	testsuite.IsMockConnClosedError(t, err)
	require.Zero(t, n)

	testsuite.IsDestroyed(t, tConn)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_TrackListener_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			testsuite.RunMultiTimes(100, func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				listeners := manager.Listeners()
				require.Equal(t, tListener, listeners[g])

				err := manager.CloseListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			})

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var manager *Manager

			init := func() {
				manager = New(nil)
			}
			track := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				listeners := manager.Listeners()
				require.Equal(t, tListener, listeners[g])

				err := manager.CloseListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			}
			cleanup := func() {
				listeners := manager.Listeners()
				require.Empty(t, listeners)

				err := manager.Close()
				require.NoError(t, err)

				listeners = manager.Listeners()
				require.Empty(t, listeners)
			}
			testsuite.RunParallelTest(100, init, cleanup, track, track)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			track := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				manager.Listeners()

				_ = manager.CloseListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				track, track, track, track,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, nil, nil, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var manager *Manager

			init := func() {
				manager = New(nil)
			}
			track := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				manager.Listeners()

				_ = manager.CloseListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				track, track, track, track,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, init, nil, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})
}

func TestManager_TrackConn_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			testsuite.RunMultiTimes(100, func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.CloseConn(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tConn)
			})

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var manager *Manager

			init := func() {
				manager = New(nil)
			}
			track := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.CloseConn(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tConn)
			}
			cleanup := func() {
				conns := manager.Conns()
				require.Empty(t, conns)

				err := manager.Close()
				require.NoError(t, err)

				conns = manager.Conns()
				require.Empty(t, conns)
			}
			testsuite.RunParallelTest(100, init, cleanup, track, track)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			track := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.CloseConn(&g)

				testsuite.IsDestroyed(t, tConn)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				track, track, track, track,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, nil, nil, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var manager *Manager

			init := func() {
				manager = New(nil)
			}
			track := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.CloseConn(&g)

				testsuite.IsDestroyed(t, tConn)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				track, track, track, track,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, init, nil, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})
}

func TestManager_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			var (
				listener1GUID guid.GUID
				listener2GUID guid.GUID
				listener3GUID guid.GUID
				listener4GUID guid.GUID
				conn1GUID     guid.GUID
				conn2GUID     guid.GUID
				conn3GUID     guid.GUID
				conn4GUID     guid.GUID
			)

			init := func() {
				listener1 := manager.TrackListener(testsuite.NewMockListener())
				listener2 := manager.TrackListener(testsuite.NewMockListener())
				listener3 := manager.TrackListener(testsuite.NewMockListener())
				listener4 := manager.TrackListener(testsuite.NewMockListener())
				conn1 := manager.TrackConn(testsuite.NewMockConn())
				conn2 := manager.TrackConn(testsuite.NewMockConn())
				conn3 := manager.TrackConn(testsuite.NewMockConn())
				conn4 := manager.TrackConn(testsuite.NewMockConn())

				listener1GUID = listener1.GUID()
				listener2GUID = listener2.GUID()
				listener3GUID = listener3.GUID()
				listener4GUID = listener4.GUID()
				conn1GUID = conn1.GUID()
				conn2GUID = conn2.GUID()
				conn3GUID = conn3.GUID()
				conn4GUID = conn4.GUID()
			}
			trackListener := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				listeners := manager.Listeners()
				require.Equal(t, tListener, listeners[g])

				err := manager.CloseListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.CloseConn(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tConn)
			}
			getListener1 := func() {
				listener, err := manager.GetListener(&listener3GUID)
				require.NoError(t, err)
				require.Equal(t, listener3GUID, listener.GUID())
			}
			getListener2 := func() {
				listener, err := manager.GetListener(&listener4GUID)
				require.NoError(t, err)
				require.Equal(t, listener4GUID, listener.GUID())
			}
			getConn1 := func() {
				conn, err := manager.GetConn(&conn3GUID)
				require.NoError(t, err)
				require.Equal(t, conn3GUID, conn.GUID())
			}
			getConn2 := func() {
				conn, err := manager.GetConn(&conn4GUID)
				require.NoError(t, err)
				require.Equal(t, conn4GUID, conn.GUID())
			}
			getListenerMaxConnsByGUID1 := func() {
				_, err := manager.GetListenerMaxConnsByGUID(&listener3GUID)
				require.NoError(t, err)
			}
			getListenerMaxConnsByGUID2 := func() {
				_, err := manager.GetListenerMaxConnsByGUID(&listener4GUID)
				require.NoError(t, err)
			}
			setListenerMaxConnsByGUID1 := func() {
				err := manager.SetListenerMaxConnsByGUID(&listener3GUID, 1000)
				require.NoError(t, err)
			}
			setListenerMaxConnsByGUID2 := func() {
				err := manager.SetListenerMaxConnsByGUID(&listener4GUID, 1000)
				require.NoError(t, err)
			}
			getListenerEstConnsNumByGUID1 := func() {
				num, err := manager.GetListenerEstConnsNumByGUID(&listener3GUID)
				require.NoError(t, err)
				require.Zero(t, num)
			}
			getListenerEstConnsNumByGUID2 := func() {
				num, err := manager.GetListenerEstConnsNumByGUID(&listener4GUID)
				require.NoError(t, err)
				require.Zero(t, num)
			}
			getConnLimitRateByGUID1 := func() {
				_, _, err := manager.GetConnLimitRateByGUID(&conn3GUID)
				require.NoError(t, err)
			}
			getConnLimitRateByGUID2 := func() {
				_, _, err := manager.GetConnLimitRateByGUID(&conn4GUID)
				require.NoError(t, err)
			}
			setConnLimitRateByGUID1 := func() {
				err := manager.SetConnLimitRateByGUID(&conn3GUID, 1000, 2000)
				require.NoError(t, err)
			}
			setConnLimitRateByGUID2 := func() {
				err := manager.SetConnLimitRateByGUID(&conn4GUID, 1000, 2000)
				require.NoError(t, err)
			}
			getConnReadLimitRateByGUID1 := func() {
				_, err := manager.GetConnReadLimitRateByGUID(&conn3GUID)
				require.NoError(t, err)
			}
			getConnReadLimitRateByGUID2 := func() {
				_, err := manager.GetConnReadLimitRateByGUID(&conn4GUID)
				require.NoError(t, err)
			}
			setConnReadLimitRateByGUID1 := func() {
				err := manager.SetConnReadLimitRateByGUID(&conn3GUID, 1000)
				require.NoError(t, err)
			}
			setConnReadLimitRateByGUID2 := func() {
				err := manager.SetConnReadLimitRateByGUID(&conn4GUID, 1000)
				require.NoError(t, err)
			}
			getConnWriteLimitRateByGUID1 := func() {
				_, err := manager.GetConnWriteLimitRateByGUID(&conn3GUID)
				require.NoError(t, err)
			}
			getConnWriteLimitRateByGUID2 := func() {
				_, err := manager.GetConnWriteLimitRateByGUID(&conn4GUID)
				require.NoError(t, err)
			}
			setConnWriteLimitRateByGUID1 := func() {
				err := manager.SetConnWriteLimitRateByGUID(&conn3GUID, 1000)
				require.NoError(t, err)
			}
			setConnWriteLimitRateByGUID2 := func() {
				err := manager.SetConnWriteLimitRateByGUID(&conn4GUID, 1000)
				require.NoError(t, err)
			}
			getListenerStatusByGUID1 := func() {
				status, err := manager.GetListenerStatusByGUID(&listener3GUID)
				require.NoError(t, err)
				require.NotNil(t, status)
			}
			getListenerStatusByGUID2 := func() {
				status, err := manager.GetListenerStatusByGUID(&listener4GUID)
				require.NoError(t, err)
				require.NotNil(t, status)
			}
			getConnStatusByGUID1 := func() {
				status, err := manager.GetConnStatusByGUID(&conn3GUID)
				require.NoError(t, err)
				require.NotNil(t, status)
			}
			getConnStatusByGUID2 := func() {
				status, err := manager.GetConnStatusByGUID(&conn4GUID)
				require.NoError(t, err)
				require.NotNil(t, status)
			}
			closeListener1 := func() {
				err := manager.CloseListener(&listener1GUID)
				require.NoError(t, err)
			}
			closeListener2 := func() {
				err := manager.CloseListener(&listener2GUID)
				require.NoError(t, err)
			}
			closeConn1 := func() {
				err := manager.CloseConn(&conn1GUID)
				require.NoError(t, err)
			}
			closeConn2 := func() {
				err := manager.CloseConn(&conn2GUID)
				require.NoError(t, err)
			}
			getListenersNum := func() {
				num := manager.GetListenersNum()
				require.NotZero(t, num)
			}
			getConnsNum := func() {
				num := manager.GetConnsNum()
				require.NotZero(t, num)
			}
			listeners := func() {
				listeners := manager.Listeners()
				require.NotEmpty(t, listeners)
			}
			conns := func() {
				conns := manager.Conns()
				require.NotEmpty(t, conns)
			}
			getAllListenersStatus := func() {
				status := manager.GetAllListenersStatus()
				require.NotEmpty(t, status)
			}
			getAllConnsStatus := func() {
				status := manager.GetAllConnsStatus()
				require.NotEmpty(t, status)
			}
			getListenerMaxConns := func() {
				manager.GetListenerMaxConns()
			}
			setListenerMaxConns := func() {
				manager.SetListenerMaxConns(1000)
			}
			getConnLimitRate := func() {
				manager.GetConnLimitRate()
			}
			setConnLimitRate := func() {
				manager.SetConnLimitRate(1000, 2000)
			}
			getConnReadLimitRate := func() {
				manager.GetConnReadLimitRate()
			}
			setConnReadLimitRate := func() {
				manager.SetConnReadLimitRate(1000)
			}
			getConnWriteLimitRate := func() {
				manager.GetConnWriteLimitRate()
			}
			setConnWriteLimitRate := func() {
				manager.SetConnWriteLimitRate(2000)
			}
			cleanup := func() {
				ls := manager.Listeners()
				require.NotEmpty(t, ls)
				cs := manager.Conns()
				require.NotEmpty(t, cs)

				err := manager.CloseListener(&listener3GUID)
				require.NoError(t, err)
				err = manager.CloseListener(&listener4GUID)
				require.NoError(t, err)
				err = manager.CloseConn(&conn3GUID)
				require.NoError(t, err)
				err = manager.CloseConn(&conn4GUID)
				require.NoError(t, err)

				ls = manager.Listeners()
				require.Empty(t, ls)
				cs = manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackListener, trackConn, trackConn,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConnsByGUID1, getListenerMaxConnsByGUID2,
				setListenerMaxConnsByGUID1, setListenerMaxConnsByGUID2,
				getListenerEstConnsNumByGUID1, getListenerEstConnsNumByGUID2,
				getListenerEstConnsNumByGUID2, getListenerEstConnsNumByGUID2,
				getConnLimitRateByGUID1, getConnLimitRateByGUID2,
				setConnLimitRateByGUID1, setConnLimitRateByGUID2,
				getConnReadLimitRateByGUID1, getConnReadLimitRateByGUID2,
				setConnReadLimitRateByGUID1, setConnReadLimitRateByGUID2,
				getConnWriteLimitRateByGUID1, getConnWriteLimitRateByGUID2,
				setConnWriteLimitRateByGUID1, setConnWriteLimitRateByGUID2,
				getListenerStatusByGUID1, getListenerStatusByGUID2,
				getConnStatusByGUID1, getConnStatusByGUID2,
				closeListener1, closeListener2, closeConn1, closeConn2,
				getListenersNum, getListenersNum, getConnsNum, getConnsNum,
				listeners, listeners, listeners, conns, conns, conns,
				getAllListenersStatus, getAllListenersStatus,
				getAllConnsStatus, getAllConnsStatus,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
			}
			testsuite.RunParallelTest(100, init, cleanup, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var (
				manager *Manager

				listener1GUID guid.GUID
				listener2GUID guid.GUID
				listener3GUID guid.GUID
				listener4GUID guid.GUID
				conn1GUID     guid.GUID
				conn2GUID     guid.GUID
				conn3GUID     guid.GUID
				conn4GUID     guid.GUID
			)

			init := func() {
				manager = New(nil)

				listener1 := manager.TrackListener(testsuite.NewMockListener())
				listener2 := manager.TrackListener(testsuite.NewMockListener())
				listener3 := manager.TrackListener(testsuite.NewMockListener())
				listener4 := manager.TrackListener(testsuite.NewMockListener())
				conn1 := manager.TrackConn(testsuite.NewMockConn())
				conn2 := manager.TrackConn(testsuite.NewMockConn())
				conn3 := manager.TrackConn(testsuite.NewMockConn())
				conn4 := manager.TrackConn(testsuite.NewMockConn())

				listener1GUID = listener1.GUID()
				listener2GUID = listener2.GUID()
				listener3GUID = listener3.GUID()
				listener4GUID = listener4.GUID()
				conn1GUID = conn1.GUID()
				conn2GUID = conn2.GUID()
				conn3GUID = conn3.GUID()
				conn4GUID = conn4.GUID()
			}
			trackListener := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				listeners := manager.Listeners()
				require.Equal(t, tListener, listeners[g])

				err := manager.CloseListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.CloseConn(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tConn)
			}
			getListener1 := func() {
				listener, err := manager.GetListener(&listener3GUID)
				require.NoError(t, err)
				require.Equal(t, listener3GUID, listener.GUID())
			}
			getListener2 := func() {
				listener, err := manager.GetListener(&listener4GUID)
				require.NoError(t, err)
				require.Equal(t, listener4GUID, listener.GUID())
			}
			getConn1 := func() {
				conn, err := manager.GetConn(&conn3GUID)
				require.NoError(t, err)
				require.Equal(t, conn3GUID, conn.GUID())
			}
			getConn2 := func() {
				conn, err := manager.GetConn(&conn4GUID)
				require.NoError(t, err)
				require.Equal(t, conn4GUID, conn.GUID())
			}
			closeListener1 := func() {
				err := manager.CloseListener(&listener1GUID)
				require.NoError(t, err)
			}
			closeListener2 := func() {
				err := manager.CloseListener(&listener2GUID)
				require.NoError(t, err)
			}
			closeConn1 := func() {
				err := manager.CloseConn(&conn1GUID)
				require.NoError(t, err)
			}
			closeConn2 := func() {
				err := manager.CloseConn(&conn2GUID)
				require.NoError(t, err)
			}
			listeners := func() {
				listeners := manager.Listeners()
				require.NotEmpty(t, listeners)
			}
			conns := func() {
				conns := manager.Conns()
				require.NotEmpty(t, conns)
			}
			getListenerMaxConns := func() {
				manager.GetListenerMaxConns()
			}
			setListenerMaxConns := func() {
				manager.SetListenerMaxConns(1000)
			}
			getConnLimitRate := func() {
				manager.GetConnLimitRate()
			}
			setConnLimitRate := func() {
				manager.SetConnLimitRate(1000, 2000)
			}
			getConnReadLimitRate := func() {
				manager.GetConnReadLimitRate()
			}
			setConnReadLimitRate := func() {
				manager.SetConnReadLimitRate(1000)
			}
			getConnWriteLimitRate := func() {
				manager.GetConnWriteLimitRate()
			}
			setConnWriteLimitRate := func() {
				manager.SetConnWriteLimitRate(2000)
			}
			cleanup := func() {
				ls := manager.Listeners()
				require.NotEmpty(t, ls)
				cs := manager.Conns()
				require.NotEmpty(t, cs)

				err := manager.Close()
				require.NoError(t, err)

				ls = manager.Listeners()
				require.Empty(t, ls)
				cs = manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackListener, trackConn, trackConn,
				listeners, listeners, conns, conns,
				closeListener1, closeListener2, closeConn1, closeConn2,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
			}
			testsuite.RunParallelTest(100, init, cleanup, fns...)

			testsuite.IsDestroyed(t, manager)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			manager := New(nil)

			var (
				listener1GUID guid.GUID
				listener2GUID guid.GUID
				listener3GUID guid.GUID
				listener4GUID guid.GUID
				conn1GUID     guid.GUID
				conn2GUID     guid.GUID
				conn3GUID     guid.GUID
				conn4GUID     guid.GUID
			)

			init := func() {
				listener1 := manager.TrackListener(testsuite.NewMockListener())
				listener2 := manager.TrackListener(testsuite.NewMockListener())
				listener3 := manager.TrackListener(testsuite.NewMockListener())
				listener4 := manager.TrackListener(testsuite.NewMockListener())
				conn1 := manager.TrackConn(testsuite.NewMockConn())
				conn2 := manager.TrackConn(testsuite.NewMockConn())
				conn3 := manager.TrackConn(testsuite.NewMockConn())
				conn4 := manager.TrackConn(testsuite.NewMockConn())

				listener1GUID = listener1.GUID()
				listener2GUID = listener2.GUID()
				listener3GUID = listener3.GUID()
				listener4GUID = listener4.GUID()
				conn1GUID = conn1.GUID()
				conn2GUID = conn2.GUID()
				conn3GUID = conn3.GUID()
				conn4GUID = conn4.GUID()
			}
			trackListener := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				manager.Listeners()

				_ = manager.CloseListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.CloseConn(&g)

				testsuite.IsDestroyed(t, tConn)
			}
			getListener1 := func() {
				_, _ = manager.GetListener(&listener3GUID)
			}
			getListener2 := func() {
				_, _ = manager.GetListener(&listener4GUID)
			}
			getConn1 := func() {
				_, _ = manager.GetConn(&conn3GUID)
			}
			getConn2 := func() {
				_, _ = manager.GetConn(&conn4GUID)
			}
			closeListener1 := func() {
				_ = manager.CloseListener(&listener1GUID)
			}
			closeListener2 := func() {
				_ = manager.CloseListener(&listener2GUID)
			}
			closeConn1 := func() {
				_ = manager.CloseConn(&conn1GUID)
			}
			closeConn2 := func() {
				_ = manager.CloseConn(&conn2GUID)
			}
			listeners := func() {
				manager.Listeners()
			}
			conns := func() {
				manager.Conns()
			}
			getListenerMaxConns := func() {
				manager.GetListenerMaxConns()
			}
			setListenerMaxConns := func() {
				manager.SetListenerMaxConns(1000)
			}
			getConnLimitRate := func() {
				manager.GetConnLimitRate()
			}
			setConnLimitRate := func() {
				manager.SetConnLimitRate(1000, 2000)
			}
			getConnReadLimitRate := func() {
				manager.GetConnReadLimitRate()
			}
			setConnReadLimitRate := func() {
				manager.SetConnReadLimitRate(1000)
			}
			getConnWriteLimitRate := func() {
				manager.GetConnWriteLimitRate()
			}
			setConnWriteLimitRate := func() {
				manager.SetConnWriteLimitRate(2000)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			cleanup := func() {
				ls := manager.Listeners()
				require.Empty(t, ls)
				cs := manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackListener, trackConn, trackConn,
				listeners, listeners, conns, conns,
				closeListener1, closeListener2, closeConn1, closeConn2,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, init, cleanup, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {
			var (
				manager *Manager

				listener1GUID guid.GUID
				listener2GUID guid.GUID
				listener3GUID guid.GUID
				listener4GUID guid.GUID
				conn1GUID     guid.GUID
				conn2GUID     guid.GUID
				conn3GUID     guid.GUID
				conn4GUID     guid.GUID
			)

			init := func() {
				manager = New(nil)

				listener1 := manager.TrackListener(testsuite.NewMockListener())
				listener2 := manager.TrackListener(testsuite.NewMockListener())
				listener3 := manager.TrackListener(testsuite.NewMockListener())
				listener4 := manager.TrackListener(testsuite.NewMockListener())
				conn1 := manager.TrackConn(testsuite.NewMockConn())
				conn2 := manager.TrackConn(testsuite.NewMockConn())
				conn3 := manager.TrackConn(testsuite.NewMockConn())
				conn4 := manager.TrackConn(testsuite.NewMockConn())

				listener1GUID = listener1.GUID()
				listener2GUID = listener2.GUID()
				listener3GUID = listener3.GUID()
				listener4GUID = listener4.GUID()
				conn1GUID = conn1.GUID()
				conn2GUID = conn2.GUID()
				conn3GUID = conn3.GUID()
				conn4GUID = conn4.GUID()
			}
			trackListener := func() {
				listener := testsuite.NewMockListener()
				tListener := manager.TrackListener(listener)

				g := tListener.GUID()

				manager.Listeners()

				_ = manager.CloseListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.CloseConn(&g)

				testsuite.IsDestroyed(t, tConn)
			}
			getListener1 := func() {
				_, _ = manager.GetListener(&listener3GUID)
			}
			getListener2 := func() {
				_, _ = manager.GetListener(&listener4GUID)
			}
			getConn1 := func() {
				_, _ = manager.GetConn(&conn3GUID)
			}
			getConn2 := func() {
				_, _ = manager.GetConn(&conn4GUID)
			}
			closeListener1 := func() {
				_ = manager.CloseListener(&listener1GUID)
			}
			closeListener2 := func() {
				_ = manager.CloseListener(&listener2GUID)
			}
			closeConn1 := func() {
				_ = manager.CloseConn(&conn1GUID)
			}
			closeConn2 := func() {
				_ = manager.CloseConn(&conn2GUID)
			}
			listeners := func() {
				manager.Listeners()
			}
			conns := func() {
				manager.Conns()
			}
			getListenerMaxConns := func() {
				manager.GetListenerMaxConns()
			}
			setListenerMaxConns := func() {
				manager.SetListenerMaxConns(1000)
			}
			getConnLimitRate := func() {
				manager.GetConnLimitRate()
			}
			setConnLimitRate := func() {
				manager.SetConnLimitRate(1000, 2000)
			}
			getConnReadLimitRate := func() {
				manager.GetConnReadLimitRate()
			}
			setConnReadLimitRate := func() {
				manager.SetConnReadLimitRate(1000)
			}
			getConnWriteLimitRate := func() {
				manager.GetConnWriteLimitRate()
			}
			setConnWriteLimitRate := func() {
				manager.SetConnWriteLimitRate(2000)
			}
			close1 := func() {
				err := manager.Close()
				require.NoError(t, err)
			}
			cleanup := func() {
				ls := manager.Listeners()
				require.Empty(t, ls)
				cs := manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackListener, trackConn, trackConn,
				listeners, listeners, conns, conns,
				closeListener1, closeListener2, closeConn1, closeConn2,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
				close1, close1, close1, close1,
			}
			testsuite.RunParallelTest(100, init, cleanup, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})
}
