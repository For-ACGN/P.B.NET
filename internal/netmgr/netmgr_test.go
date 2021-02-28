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
				conn1GUID     guid.GUID
				conn2GUID     guid.GUID
			)

			init := func() {
				listener1 := manager.TrackListener(testsuite.NewMockListener())
				listener2 := manager.TrackListener(testsuite.NewMockListener())
				conn1 := manager.TrackConn(testsuite.NewMockConn())
				conn2 := manager.TrackConn(testsuite.NewMockConn())

				listener1GUID = listener1.GUID()
				listener2GUID = listener2.GUID()
				conn1GUID = conn1.GUID()
				conn2GUID = conn2.GUID()
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
			listeners := func() {
				manager.Listeners()
			}
			conns := func() {
				manager.Conns()
			}
			killListener := func() {
				err := manager.KillListener(&listener1GUID)
				require.NoError(t, err)
				err = manager.KillListener(&listener2GUID)
				require.NoError(t, err)
			}
			killConn := func() {
				err := manager.KillConn(&conn1GUID)
				require.NoError(t, err)
				err = manager.KillConn(&conn2GUID)
				require.NoError(t, err)
			}
			cleanup := func() {
				ls := manager.Listeners()
				require.Empty(t, ls)

				cs := manager.Conns()
				require.Empty(t, cs)
			}
			fns := []func(){
				trackListener, trackConn, listeners, conns,
				killListener, killConn,
			}
			testsuite.RunParallel(100, init, cleanup, fns...)

			err := manager.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, manager)
		})

		t.Run("whole", func(t *testing.T) {

		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {

		})

		t.Run("whole", func(t *testing.T) {

		})
	})
}
