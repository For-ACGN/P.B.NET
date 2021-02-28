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

func TestManager_GetListenerMaxConns(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	maxConns := netmgr.GetListenerMaxConns()
	require.Zero(t, maxConns)

	netmgr.SetListenerMaxConns(1000)

	maxConns = netmgr.GetListenerMaxConns()
	require.Equal(t, uint64(1000), maxConns)

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_GetConnLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	read, write := netmgr.GetConnLimitRate()
	require.Zero(t, read)
	require.Zero(t, write)

	netmgr.SetConnLimitRate(1000, 2000)

	read, write = netmgr.GetConnLimitRate()
	require.Equal(t, uint64(1000), read)
	require.Equal(t, uint64(2000), write)

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_GetConnReadLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	read := netmgr.GetConnReadLimitRate()
	require.Zero(t, read)

	netmgr.SetConnReadLimitRate(1000)

	read = netmgr.GetConnReadLimitRate()
	require.Equal(t, uint64(1000), read)

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}

func TestManager_GetConnWriteLimitRate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	netmgr := New(nil)

	write := netmgr.GetConnWriteLimitRate()
	require.Zero(t, write)

	netmgr.SetConnWriteLimitRate(1000)

	write = netmgr.GetConnWriteLimitRate()
	require.Equal(t, uint64(1000), write)

	err := netmgr.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, netmgr)
}
