package netmgr

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestListener(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		err := listener.Close()
		require.Error(t, err)
	}()
	address := listener.Addr().String()

	tListener := manager.TrackListener(listener)

	guid := tListener.GUID()
	require.False(t, guid.IsZero())
	require.NotZero(t, tListener.Status().Listened)

	testsuite.ListenerAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	}, true)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestListener_Accept(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		err := listener.Close()
		require.Error(t, err)
	}()
	address := listener.Addr().String()

	tListener := manager.TrackListener(listener)

	num := tListener.GetEstConnsNum()
	require.Zero(t, num)
	status := tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)
	require.Zero(t, status.LastAccept)

	server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
		return net.Dial("tcp", address)
	})

	num = tListener.GetEstConnsNum()
	require.Equal(t, uint64(1), num)
	status = tListener.Status()
	require.Equal(t, uint64(1), status.EstConns)
	require.NotZero(t, status.LastAccept)

	testsuite.ConnSC(t, server, client, true)

	num = tListener.GetEstConnsNum()
	require.Zero(t, num)
	status = tListener.Status()
	require.Equal(t, uint64(0), status.EstConns)
	require.NotZero(t, status.LastAccept)

	err = tListener.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, tListener)

	err = manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestListener_AcceptEx(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("set max conns", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		address := listener.Addr().String()

		tListener := manager.TrackListener(listener)

		maxConns := tListener.GetMaxConns()
		require.Zero(t, maxConns)

		tListener.SetMaxConns(1)
		maxConns = tListener.GetMaxConns()
		require.Equal(t, uint64(1), maxConns)

		testsuite.ListenerAndDial(t, tListener, func() (net.Conn, error) {
			return net.Dial("tcp", address)
		}, true)
	})

	t.Run("reach max conns", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		address := listener.Addr().String()

		tListener := manager.TrackListener(listener)
		tListener.SetMaxConns(1)

		server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
			return net.Dial("tcp", address)
		})

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			now := time.Now()

			server, client := testsuite.AcceptAndDial(t, tListener, func() (net.Conn, error) {
				return net.Dial("tcp", address)
			})
			err := server.Close()
			require.NoError(t, err)
			err = client.Close()
			require.NoError(t, err)

			require.True(t, time.Since(now) > 3*time.Second)
		}()

		time.Sleep(4 * time.Second)

		err = server.Close()
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)

		wg.Wait()

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("listener is closed", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		tListener := manager.TrackListener(listener)

		err = tListener.Close()
		require.NoError(t, err)

		conn, err := tListener.Accept()
		require.EqualError(t, err, "listener is closed")
		require.Nil(t, conn)

		testsuite.IsDestroyed(t, tListener)
	})

	t.Run("failed to accept", func(t *testing.T) {
		listener := testsuite.NewMockListenerWithAcceptError()
		tListener := manager.TrackListener(listener)

		conn, err := tListener.Accept()
		testsuite.IsMockListenerAcceptError(t, err)
		require.Nil(t, conn)

		err = tListener.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, tListener)
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}

func TestListener_Close(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("close when full conns", func(t *testing.T) {
		manager := New(nil)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() {
			err := listener.Close()
			require.Error(t, err)
		}()
		address := listener.Addr().String()

		tListener := manager.TrackListener(listener)
		tListener.SetMaxConns(1)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn1, err := tListener.Accept()
			require.NoError(t, err)

			conn2, err := tListener.Accept()
			require.EqualError(t, err, "listener is closed")
			require.Nil(t, conn2)

			err = conn1.Close()
			require.NoError(t, err)
		}()

		conn, err := net.Dial("tcp", address)
		require.NoError(t, err)

		// wait to run the second Accept
		time.Sleep(3 * time.Second)

		err = tListener.Close()
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)

		wg.Wait()

		testsuite.IsDestroyed(t, tListener)

		err = manager.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, manager)
	})

	t.Run("failed to close inner listener", func(t *testing.T) {
		manager := New(nil)

		listener := testsuite.NewMockListenerWithCloseError()
		tListener := manager.TrackListener(listener)

		err := tListener.Close()
		testsuite.IsMockListenerCloseError(t, err)

		err = manager.Close()
		testsuite.IsMockListenerCloseError(t, err)

		testsuite.IsDestroyed(t, tListener)
		testsuite.IsDestroyed(t, manager)
	})
}

func TestListener_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	manager := New(nil)

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			listener := testsuite.NewMockListener()
			tListener := manager.TrackListener(listener)

			accept := func() {
				conn, err := tListener.Accept()
				require.NoError(t, err)
				err = conn.Close()
				require.NoError(t, err)
			}
			getGUID := func() {
				tListener.GUID()
			}
			getMaxConns := func() {
				tListener.GetMaxConns()
			}
			setMaxConns := func() {
				tListener.SetMaxConns(1000)
			}
			getEstConnsNum := func() {
				tListener.GetEstConnsNum()
			}
			status := func() {
				tListener.Status()
			}
			fns := []func(){
				accept, getGUID,
				getMaxConns, setMaxConns,
				getEstConnsNum, status,
			}
			testsuite.RunParallel(100, nil, nil, fns...)

			err := tListener.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tListener)
		})

		t.Run("whole", func(t *testing.T) {
			var tListener *Listener

			init := func() {
				listener := testsuite.NewMockListener()
				tListener = manager.TrackListener(listener)
			}
			accept := func() {
				conn, err := tListener.Accept()
				require.NoError(t, err)
				err = conn.Close()
				require.NoError(t, err)
			}
			getGUID := func() {
				tListener.GUID()
			}
			getMaxConns := func() {
				tListener.GetMaxConns()
			}
			setMaxConns := func() {
				tListener.SetMaxConns(1000)
			}
			getEstConnsNum := func() {
				tListener.GetEstConnsNum()
			}
			status := func() {
				tListener.Status()
			}
			cleanup := func() {
				err := tListener.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				accept, getGUID,
				getMaxConns, setMaxConns,
				getEstConnsNum, status,
			}
			testsuite.RunParallel(100, init, cleanup, fns...)

			err := tListener.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tListener)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			listener := testsuite.NewMockListener()
			tListener := manager.TrackListener(listener)

			accept := func() {
				conn, err := tListener.Accept()
				if err != nil {
					return
				}
				err = conn.Close()
				require.NoError(t, err)
			}
			getGUID := func() {
				tListener.GUID()
			}
			getMaxConns := func() {
				tListener.GetMaxConns()
			}
			setMaxConns := func() {
				tListener.SetMaxConns(1000)
			}
			getEstConnsNum := func() {
				tListener.GetEstConnsNum()
			}
			status := func() {
				tListener.Status()
			}
			close1 := func() {
				err := tListener.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				accept, getGUID,
				getMaxConns, setMaxConns,
				getEstConnsNum, status,
				close1,
			}
			testsuite.RunParallel(100, nil, nil, fns...)

			err := tListener.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tListener)
		})

		t.Run("whole", func(t *testing.T) {
			var tListener *Listener

			init := func() {
				listener := testsuite.NewMockListener()
				tListener = manager.TrackListener(listener)
			}
			accept := func() {
				conn, err := tListener.Accept()
				if err != nil {
					return
				}
				err = conn.Close()
				require.NoError(t, err)
			}
			getGUID := func() {
				tListener.GUID()
			}
			getMaxConns := func() {
				tListener.GetMaxConns()
			}
			setMaxConns := func() {
				tListener.SetMaxConns(1000)
			}
			getEstConnsNum := func() {
				tListener.GetEstConnsNum()
			}
			status := func() {
				tListener.Status()
			}
			close1 := func() {
				err := tListener.Close()
				require.NoError(t, err)
			}
			fns := []func(){
				accept, getGUID,
				getMaxConns, setMaxConns,
				getEstConnsNum, status,
				close1,
			}
			testsuite.RunParallel(100, init, nil, fns...)

			err := tListener.Close()
			require.NoError(t, err)

			testsuite.IsDestroyed(t, tListener)
		})
	})

	err := manager.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, manager)
}
