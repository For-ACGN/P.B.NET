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

func TestManager_KillListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	listener := testsuite.NewMockListener()
	tListener := manager.TrackListener(listener)

	listeners := manager.Listeners()
	require.Len(t, listeners, 1)

	g := tListener.GUID()
	err := manager.KillListener(&g)
	require.NoError(t, err)

	listeners = manager.Listeners()
	require.Empty(t, listeners)

	err = manager.KillListener(&g)
	require.Error(t, err)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestManager_KillConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	conn := testsuite.NewMockConn()
	tConn := manager.TrackConn(conn)

	conns := manager.Conns()
	require.Len(t, conns, 1)

	g := tConn.GUID()
	err := manager.KillConn(&g)
	require.NoError(t, err)

	conns = manager.Conns()
	require.Empty(t, conns)

	err = manager.KillConn(&g)
	require.Error(t, err)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
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

				err := manager.KillListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.KillConn(&g)
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
			killListener1 := func() {
				err := manager.KillListener(&listener1GUID)
				require.NoError(t, err)
			}
			killListener2 := func() {
				err := manager.KillListener(&listener2GUID)
				require.NoError(t, err)
			}
			killConn1 := func() {
				err := manager.KillConn(&conn1GUID)
				require.NoError(t, err)
			}
			killConn2 := func() {
				err := manager.KillConn(&conn2GUID)
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

				err := manager.KillListener(&listener3GUID)
				require.NoError(t, err)
				err = manager.KillListener(&listener4GUID)
				require.NoError(t, err)
				err = manager.KillConn(&conn3GUID)
				require.NoError(t, err)
				err = manager.KillConn(&conn4GUID)
				require.NoError(t, err)

				ls = manager.Listeners()
				require.Empty(t, ls)
				cs = manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackListener, trackConn, trackConn,
				listeners, listeners, conns, conns,
				killListener1, killListener2, killConn1, killConn2,
				getListener1, getListener2, getConn1, getConn2,
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

				err := manager.KillListener(&g)
				require.NoError(t, err)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				conns := manager.Conns()
				require.Equal(t, tConn, conns[g])

				err := manager.KillConn(&g)
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
			killListener1 := func() {
				err := manager.KillListener(&listener1GUID)
				require.NoError(t, err)
			}
			killListener2 := func() {
				err := manager.KillListener(&listener2GUID)
				require.NoError(t, err)
			}
			killConn1 := func() {
				err := manager.KillConn(&conn1GUID)
				require.NoError(t, err)
			}
			killConn2 := func() {
				err := manager.KillConn(&conn2GUID)
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
				killListener1, killListener2, killConn1, killConn2,
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

				_ = manager.KillListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.KillConn(&g)

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
			killListener1 := func() {
				_ = manager.KillListener(&listener1GUID)
			}
			killListener2 := func() {
				_ = manager.KillListener(&listener2GUID)
			}
			killConn1 := func() {
				_ = manager.KillConn(&conn1GUID)
			}
			killConn2 := func() {
				_ = manager.KillConn(&conn2GUID)
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
				killListener1, killListener2, killConn1, killConn2,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
				close1, close1,
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

				_ = manager.KillListener(&g)

				testsuite.IsDestroyed(t, tListener)
			}
			trackConn := func() {
				conn := testsuite.NewMockConn()
				tConn := manager.TrackConn(conn)

				g := tConn.GUID()

				manager.Conns()

				_ = manager.KillConn(&g)

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
			killListener1 := func() {
				_ = manager.KillListener(&listener1GUID)
			}
			killListener2 := func() {
				_ = manager.KillListener(&listener2GUID)
			}
			killConn1 := func() {
				_ = manager.KillConn(&conn1GUID)
			}
			killConn2 := func() {
				_ = manager.KillConn(&conn2GUID)
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
				killListener1, killListener2, killConn1, killConn2,
				getListener1, getListener2, getConn1, getConn2,
				getListenerMaxConns, setListenerMaxConns,
				getConnLimitRate, setConnLimitRate,
				getConnReadLimitRate, setConnReadLimitRate,
				getConnWriteLimitRate, setConnWriteLimitRate,
				close1, close1,
			}
			testsuite.RunParallelTest(100, init, cleanup, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})
	})
}
