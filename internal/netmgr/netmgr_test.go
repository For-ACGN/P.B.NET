package netmgr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestManager_TrackListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	listener := testsuite.NewMockListener()
	tListener := netmgr.TrackListener(listener)

	listeners := netmgr.Listeners()
	require.Len(t, listeners, 1)
	require.Equal(t, tListener, listeners[tListener.GUID()])

	err := tListener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener)

	listeners = netmgr.Listeners()
	require.Empty(t, listeners)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_TrackConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	conn := testsuite.NewMockConn()
	tConn := netmgr.TrackConn(conn)

	conns := netmgr.Conns()
	require.Len(t, conns, 1)
	require.Equal(t, tConn, conns[tConn.GUID()])

	err := tConn.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tConn)

	conns = netmgr.Conns()
	require.Empty(t, conns)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_KillListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	listener := testsuite.NewMockListener()
	tListener := netmgr.TrackListener(listener)

	listeners := netmgr.Listeners()
	require.Len(t, listeners, 1)

	guid := tListener.GUID()
	err := netmgr.KillListener(&guid)
	require.NoError(t, err)

	listeners = netmgr.Listeners()
	require.Empty(t, listeners)

	err = netmgr.KillListener(&guid)
	require.Error(t, err)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_KillConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	conn := testsuite.NewMockConn()
	tConn := netmgr.TrackConn(conn)

	conns := netmgr.Conns()
	require.Len(t, conns, 1)

	guid := tConn.GUID()
	err := netmgr.KillConn(&guid)
	require.NoError(t, err)

	conns = netmgr.Conns()
	require.Empty(t, conns)

	err = netmgr.KillConn(&guid)
	require.Error(t, err)

	err = netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}
