package lcx

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/netutil"

	"project/internal/logger"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func testGenerateListener(t *testing.T) *Listener {
	iNetwork := "tcp"
	iAddress := "127.0.0.1:0"
	opts := Options{LocalAddress: "127.0.0.1:0"}
	listener, err := NewListener("test", iNetwork, iAddress, logger.Test, &opts)
	require.NoError(t, err)
	return listener
}

func TestListener(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener := testGenerateListener(t)

	t.Log(listener.Name())
	t.Log(listener.Info())
	t.Log(listener.Status())

	require.Zero(t, listener.testIncomeAddress())
	require.Zero(t, listener.testLocalAddress())

	err := listener.Start()
	require.NoError(t, err)

	// mock slaver and started copy
	var targetAddress string
	switch {
	case testsuite.IPv4Enabled:
		targetAddress = "127.0.0.1:" + testsuite.HTTPServerPort
	case testsuite.IPv6Enabled:
		targetAddress = "[::1]:" + testsuite.HTTPServerPort
	}

	// slaver connect the listener
	iConn, err := net.Dial("tcp", listener.testIncomeAddress())
	require.NoError(t, err)
	defer func() { _ = iConn.Close() }() // must close
	// slaver connect the target
	tConn, err := net.Dial("tcp", targetAddress)
	require.NoError(t, err)
	defer func() { _ = tConn.Close() }() // must close
	// start copy
	go func() {
		_, _ = io.Copy(tConn, iConn)
	}()
	go func() {
		_, _ = io.Copy(iConn, tConn)
	}()
	// user dial local listener
	lConn, err := net.Dial("tcp", listener.testLocalAddress())
	require.NoError(t, err)
	testsuite.ProxyConn(t, lConn)

	time.Sleep(2 * time.Second)

	t.Log(listener.Name())
	t.Log(listener.Info())
	t.Log(listener.Status())

	err = listener.Restart()
	require.NoError(t, err)

	listener.Stop()

	testsuite.IsDestroyed(t, listener)
}

func TestNewListener(t *testing.T) {
	const (
		tag      = "test"
		iNetwork = "tcp"
		iAddress = "127.0.0.1:80"
	)

	t.Run("empty tag", func(t *testing.T) {
		_, err := NewListener("", "", "", nil, nil)
		require.Error(t, err)
	})

	t.Run("empty income address", func(t *testing.T) {
		_, err := NewListener(tag, "", "", nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid income address", func(t *testing.T) {
		_, err := NewListener(tag, "foo", "foo", nil, nil)
		require.Error(t, err)
	})

	t.Run("empty options", func(t *testing.T) {
		_, err := NewListener(tag, iNetwork, iAddress, logger.Test, nil)
		require.NoError(t, err)
	})

	t.Run("invalid local address", func(t *testing.T) {
		opts := Options{
			LocalNetwork: "foo",
			LocalAddress: "foo",
		}
		_, err := NewListener(tag, iNetwork, iAddress, logger.Test, &opts)
		require.Error(t, err)
	})
}

func TestListener_Start(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("started twice", func(t *testing.T) {
		listener := testGenerateListener(t)

		err := listener.Start()
		require.NoError(t, err)
		err = listener.Start()
		require.Error(t, err)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("failed to listen income", func(t *testing.T) {
		iNetwork := "tcp"
		iAddress := "0.0.0.1:0"
		tranner, err := NewListener("test", iNetwork, iAddress, logger.Test, nil)
		require.NoError(t, err)

		err = tranner.Start()
		require.Error(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("failed to listen local", func(t *testing.T) {
		iNetwork := "tcp"
		iAddress := "127.0.0.1:0"
		opts := Options{LocalAddress: "0.0.0.1:0"}
		tranner, err := NewListener("test", iNetwork, iAddress, logger.Test, &opts)
		require.NoError(t, err)

		err = tranner.Start()
		require.Error(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})
}

func TestListener_Stop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		listener := testGenerateListener(t)

		err := listener.Start()
		require.NoError(t, err)

		iConn, err := net.Dial("tcp", listener.testIncomeAddress())
		require.NoError(t, err)
		defer func() { _ = iConn.Close() }()

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		defer func() { _ = lConn.Close() }()

		// wait serve
		time.Sleep(time.Second)

		t.Log(listener.Status())

		listener.Stop()
		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close with error", func(t *testing.T) {
		listener := testGenerateListener(t)

		listener.iListener = testsuite.NewMockListenerWithCloseError()
		listener.lListener = testsuite.NewMockListenerWithCloseError()

		conn := &lConn{
			ctx:    listener,
			remote: testsuite.NewMockConnWithCloseError(),
			local:  testsuite.NewMockConnWithCloseError(),
		}
		listener.trackConn(conn, true)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})
}

func TestListener_Call(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

}

func TestListener_List(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

}

func TestListener_Kill(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

}

func TestListener_serve(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("accept income", func(t *testing.T) {
		listener := testGenerateListener(t)

		err := listener.Start()
		require.NoError(t, err)

		iConn, err := net.Dial("tcp", listener.testIncomeAddress())
		require.NoError(t, err)
		defer func() { _ = iConn.Close() }()

		// wait accept
		time.Sleep(time.Second)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("accept panic", func(t *testing.T) {
		listener := testGenerateListener(t)

		patch := func(net.Listener, int) net.Listener {
			return testsuite.NewMockListenerWithAcceptPanic()
		}
		pg := monkey.Patch(netutil.LimitListener, patch)
		defer pg.Unpatch()

		err := listener.Start()
		require.NoError(t, err)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close listener error", func(t *testing.T) {
		listener := testGenerateListener(t)

		patch := func(net.Listener, int) net.Listener {
			return testsuite.NewMockListenerWithCloseError()
		}
		pg := monkey.Patch(netutil.LimitListener, patch)
		defer pg.Unpatch()

		err := listener.Start()
		require.NoError(t, err)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})
}

func TestListener_accept(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener := testGenerateListener(t)

	patch := func(net.Listener, int) net.Listener {
		return testsuite.NewMockListenerWithAcceptError()
	}
	pg := monkey.Patch(netutil.LimitListener, patch)
	defer pg.Unpatch()

	err := listener.Start()
	require.NoError(t, err)

	listener.Stop()

	testsuite.IsDestroyed(t, listener)
}

func TestListener_trackConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener := testGenerateListener(t)

	t.Run("failed to add conn", func(t *testing.T) {
		ok := listener.trackConn(nil, true)
		require.False(t, ok)
	})

	t.Run("add and delete", func(t *testing.T) {
		err := listener.Start()
		require.NoError(t, err)

		ok := listener.trackConn(nil, true)
		require.True(t, ok)

		ok = listener.trackConn(nil, false)
		require.True(t, ok)
	})

	listener.Stop()

	testsuite.IsDestroyed(t, listener)
}

func TestLConn_Serve(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to track conn", func(t *testing.T) {
		listener := testGenerateListener(t)

		remote := testsuite.NewMockConn()
		local := testsuite.NewMockConn()
		conn := listener.newConn(remote, local)
		conn.Serve()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("close connection error", func(t *testing.T) {
		listener := testGenerateListener(t)

		remote := testsuite.NewMockConnWithCloseError()
		local := testsuite.NewMockConnWithCloseError()
		conn := listener.newConn(remote, local)
		conn.Serve()

		testsuite.IsDestroyed(t, listener)
	})

	t.Run("panic from copy", func(t *testing.T) {
		listener := testGenerateListener(t)

		patch := func(io.Writer, io.Reader) (int64, error) {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(io.Copy, patch)
		defer pg.Unpatch()

		err := listener.Start()
		require.NoError(t, err)

		iConn, err := net.Dial("tcp", listener.testIncomeAddress())
		require.NoError(t, err)
		defer func() { _ = iConn.Close() }()

		lConn, err := net.Dial("tcp", listener.testLocalAddress())
		require.NoError(t, err)
		defer func() { _ = lConn.Close() }()

		// wait serve
		time.Sleep(time.Second)

		listener.Stop()

		testsuite.IsDestroyed(t, listener)
	})
}

func TestLConn_Close(t *testing.T) {
	conn := lConn{
		remote: testsuite.NewMockConnWithCloseError(),
		local:  testsuite.NewMockConn(),
	}
	err := conn.Close()
	require.Error(t, err)
}

func TestListener_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	listener := testGenerateListener(t)

	start := func() {
		_ = listener.Start()
	}
	stop := func() {
		listener.Stop()
	}
	restart := func() {
		_ = listener.Restart()
	}
	info := func() {
		_ = listener.Info()
	}
	status := func() {
		_ = listener.Status()
	}
	track := func() {
		conn := &lConn{
			ctx:    listener,
			remote: testsuite.NewMockConn(),
			local:  testsuite.NewMockConn(),
		}
		listener.trackConn(conn, true)
	}
	testsuite.RunParallel(100, nil, nil,
		start, stop, restart, info, status, track)

	listener.Stop()

	testsuite.IsDestroyed(t, listener)
}
