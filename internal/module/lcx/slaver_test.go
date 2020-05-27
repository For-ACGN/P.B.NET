package lcx

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/patch/monkey"
	"project/internal/random"
	"project/internal/testsuite"
)

func testGenerateListenerAndSlaver(t *testing.T) (*Listener, *Slaver) {
	listener := testGenerateListener(t)
	err := listener.Start()
	require.NoError(t, err)

	lNetwork := "tcp"
	lAddress := listener.testIncomeAddress()
	dstNetwork := "tcp"
	var dstAddress string
	switch {
	case testsuite.IPv4Enabled:
		dstAddress = "127.0.0.1:" + testsuite.HTTPServerPort
	case testsuite.IPv6Enabled:
		dstAddress = "[::1]:" + testsuite.HTTPServerPort
	}
	opts := Options{LocalAddress: "127.0.0.1:0"}
	slaver, err := NewSlaver("test", lNetwork, lAddress,
		dstNetwork, dstAddress, logger.Test, &opts)
	require.NoError(t, err)
	return listener, slaver
}

func TestSlaver(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, slaver := testGenerateListenerAndSlaver(t)

	t.Log(slaver.Name())
	t.Log(slaver.Info())
	t.Log(slaver.Status())

	err := slaver.Start()
	require.NoError(t, err)

	// user dial local listener
	for i := 0; i < 3; i++ {
		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)

		testsuite.ProxyConn(t, lConn)
	}

	time.Sleep(2 * time.Second)

	t.Log(slaver.Name())
	t.Log(slaver.Info())
	t.Log(slaver.Status())

	err = slaver.Restart()
	require.NoError(t, err)

	slaver.Stop()
	listener.Stop()

	testsuite.IsDestroyed(t, slaver)
	testsuite.IsDestroyed(t, listener)
}

func TestNewSlaver(t *testing.T) {
	const (
		tag        = "test"
		lNetwork   = "tcp"
		lAddress   = "127.0.0.1:80"
		dstNetwork = "tcp"
		dstAddress = "127.0.0.1:3389"
	)

	t.Run("empty tag", func(t *testing.T) {
		_, err := NewSlaver("", lNetwork, lAddress,
			dstNetwork, dstAddress, nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid listener address", func(t *testing.T) {
		_, err := NewSlaver(tag, "foo", "foo",
			dstNetwork, dstAddress, nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid destination address", func(t *testing.T) {
		_, err := NewSlaver(tag, lNetwork, lAddress,
			"foo", "foo", nil, nil)
		require.Error(t, err)
	})

	t.Run("empty options", func(t *testing.T) {
		_, err := NewSlaver(tag, lNetwork, lAddress,
			dstNetwork, dstAddress, nil, nil)
		require.NoError(t, err)
	})
}

func TestSlaver_Start(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, slaver := testGenerateListenerAndSlaver(t)

	err := slaver.Start()
	require.NoError(t, err)
	err = slaver.Start()
	require.Error(t, err)

	slaver.Stop()
	listener.Stop()

	testsuite.IsDestroyed(t, slaver)
	testsuite.IsDestroyed(t, listener)
}

func TestSlaver_Stop(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		err := slaver.Start()
		require.NoError(t, err)

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		defer func() { _ = lConn.Close() }()

		slaver.Stop()
		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close with error", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.started = true
		slaver.ctx, slaver.cancel = context.WithCancel(context.Background())

		conn := &sConn{
			slaver: slaver,
			local:  testsuite.NewMockConnWithCloseError(),
		}
		slaver.trackConn(conn, true)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})
}

func TestSlaver_serve(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("full", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.opts.MaxConns = 1 // force change

		sleeper := new(random.Sleeper)
		patch := func(interface{}, uint, uint) <-chan time.Time {
			return time.After(500 * time.Millisecond)
		}
		pg := monkey.PatchInstanceMethod(sleeper, "Sleep", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		defer func() { _ = lConn.Close() }()
		_, err = lConn.Write(make([]byte, 1))
		require.NoError(t, err)

		// wait call full()
		time.Sleep(2 * time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("failed to connect listener", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.lAddress = "0.0.0.0:1"
		slaver.opts.MaxConns = 1 // force change

		sleeper := new(random.Sleeper)
		patch1 := func(interface{}, uint, uint) <-chan time.Time {
			return time.After(500 * time.Millisecond)
		}
		pg1 := monkey.PatchInstanceMethod(sleeper, "Sleep", patch1)
		defer pg1.Unpatch()

		dialer := new(net.Dialer)
		patch2 := func(interface{}, context.Context, string, string) (net.Conn, error) {
			return nil, monkey.Error
		}
		pg2 := monkey.PatchInstanceMethod(dialer, "DialContext", patch2)
		defer pg2.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(2 * time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("panic", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		patch := func(context.Context, time.Duration) (context.Context, context.CancelFunc) {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(context.WithTimeout, patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})
}

func TestSlaver_trackConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, slaver := testGenerateListenerAndSlaver(t)

	t.Run("failed to add conn", func(t *testing.T) {
		ok := slaver.trackConn(nil, true)
		require.False(t, ok)
	})

	t.Run("add and delete", func(t *testing.T) {
		err := slaver.Start()
		require.NoError(t, err)

		ok := slaver.trackConn(nil, true)
		require.True(t, ok)

		ok = slaver.trackConn(nil, false)
		require.True(t, ok)
	})

	slaver.Stop()
	listener.Stop()

	testsuite.IsDestroyed(t, slaver)
	testsuite.IsDestroyed(t, listener)
}

func TestSConn_Serve(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to track conn", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.ctx = context.Background()

		c := testsuite.NewMockConn()
		conn := slaver.newConn(c)
		conn.Serve()

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close local connection error", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.ctx = context.Background()

		c := testsuite.NewMockConnWithCloseError()
		conn := slaver.newConn(c)
		conn.Serve()

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("failed to connect target", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)
		slaver.dstAddress = "0.0.0.0:1"

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("local failed to write", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		dialer := new(net.Dialer)
		patch := func(interface{}, context.Context, string, string) (net.Conn, error) {
			return testsuite.NewMockConnWithWriteError(), nil
		}
		pg := monkey.PatchInstanceMethod(dialer, "DialContext", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(10 * time.Millisecond)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("done block local to remote", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		conn := new(sConn)
		patch := func(c *sConn) {
			done := make(chan struct{}, 2)
			// block
			done <- struct{}{}
			done <- struct{}{}
			c.slaver.wg.Add(1)
			go c.serve(done)

			time.Sleep(time.Second)
			go slaver.Stop()
			go listener.Stop()

			<-c.slaver.ctx.Done()
		}
		pg := monkey.PatchInstanceMethod(conn, "Serve", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		_, _ = lConn.Write(make([]byte, 1))

		// wait serve
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		// because of monkey
		// testsuite.IsDestroyed(t, slaver)
		// testsuite.IsDestroyed(t, listener)
	})

	t.Run("done block remote to local", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		conn := new(sConn)
		patch := func(c *sConn) {
			done := make(chan struct{}, 2)
			// block
			done <- struct{}{}
			c.slaver.wg.Add(1)
			go c.serve(done)

			time.Sleep(time.Second)
			go slaver.Stop()
			go listener.Stop()

			<-c.slaver.ctx.Done()
		}
		pg := monkey.PatchInstanceMethod(conn, "Serve", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		testsuite.SendHTTPRequest(t, lConn)

		// wait serve
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		// because of monkey
		// testsuite.IsDestroyed(t, slaver)
		// testsuite.IsDestroyed(t, listener)
	})

	t.Run("done block in defer", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		conn := new(sConn)
		patch1 := func(c *sConn) {
			done := make(chan struct{})
			c.slaver.wg.Add(1)
			go c.serve(done)

			<-c.slaver.ctx.Done()
		}
		pg1 := monkey.PatchInstanceMethod(conn, "Serve", patch1)
		defer pg1.Unpatch()

		dialer := new(net.Dialer)
		patch2 := func(interface{}, context.Context, string, string) (net.Conn, error) {
			return testsuite.NewMockConn(), nil
		}
		pg2 := monkey.PatchInstanceMethod(dialer, "DialContext", patch2)
		defer pg2.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		// because of monkey
		// testsuite.IsDestroyed(t, slaver)
		// testsuite.IsDestroyed(t, listener)

	})

	t.Run("panic from copy", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		conn := new(net.TCPConn)
		patch := func(interface{}, time.Time) error {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(conn, "SetReadDeadline", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close remote connection error", func(t *testing.T) {
		listener, slaver := testGenerateListenerAndSlaver(t)

		dialer := new(net.Dialer)
		patch := func(interface{}, context.Context, string, string) (net.Conn, error) {
			return testsuite.NewMockConnWithCloseError(), nil
		}
		pg := monkey.PatchInstanceMethod(dialer, "DialContext", patch)
		defer pg.Unpatch()

		err := slaver.Start()
		require.NoError(t, err)

		// wait serve()
		time.Sleep(time.Second)

		slaver.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, slaver)
		testsuite.IsDestroyed(t, listener)
	})
}

func TestSlaver_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener, slaver := testGenerateListenerAndSlaver(t)

	f1 := func() {
		_ = slaver.Start()
	}
	f2 := func() {
		slaver.Stop()
	}
	f3 := func() {
		_ = slaver.Restart()
	}
	f4 := func() {
		_ = slaver.Info()
	}
	f5 := func() {
		_ = slaver.Status()
	}
	f6 := func() {
		conn := &sConn{
			slaver: slaver,
			local:  testsuite.NewMockConn(),
		}
		slaver.trackConn(conn, true)
	}
	testsuite.RunParallel(100, f1, f2, f3, f4, f5, f6)

	listener.Stop()
	slaver.Stop()

	testsuite.IsDestroyed(t, listener)
	testsuite.IsDestroyed(t, slaver)
}
