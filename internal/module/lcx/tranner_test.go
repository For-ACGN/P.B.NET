package lcx

import (
	"context"
	"fmt"
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

func testGenerateTranner(t *testing.T) *Tranner {
	dstNetwork := "tcp"
	var dstAddress string
	switch {
	case testsuite.IPv4Enabled:
		dstAddress = "127.0.0.1:" + testsuite.HTTPServerPort
	case testsuite.IPv6Enabled:
		dstAddress = "[::1]:" + testsuite.HTTPServerPort
	}
	opts := Options{LocalAddress: "127.0.0.1:0"}
	tranner, err := NewTranner("test", dstNetwork, dstAddress, logger.Test, &opts)
	require.NoError(t, err)
	return tranner
}

func TestTranner(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	t.Log(tranner.Name())
	t.Log(tranner.Description())

	t.Log(tranner.Info())
	t.Log(tranner.Status())

	require.Zero(t, tranner.testAddress())

	err := tranner.Start()
	require.NoError(t, err)

	// test connect test http server
	address := tranner.testAddress()
	testsuite.RunMultiTimes(100, func() {
		lConn, err := net.Dial("tcp", address)
		require.NoError(t, err)

		testsuite.ProxyConn(t, lConn)
	})

	t.Log(tranner.Info())
	t.Log(tranner.Status())

	for _, method := range tranner.Methods() {
		fmt.Println(method)
		fmt.Println()
	}

	err = tranner.Restart()
	require.NoError(t, err)
	require.True(t, tranner.IsStarted())

	tranner.Stop()
	require.False(t, tranner.IsStarted())

	testsuite.IsDestroyed(t, tranner)
}

func TestNewTranner(t *testing.T) {
	const (
		tag        = "test"
		dstNetwork = "tcp"
		dstAddress = "127.0.0.1:80"
	)

	t.Run("empty tag", func(t *testing.T) {
		_, err := NewTranner("", "", "", nil, nil)
		require.Error(t, err)
	})

	t.Run("empty destination address", func(t *testing.T) {
		_, err := NewTranner(tag, "", "", nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid destination address", func(t *testing.T) {
		_, err := NewTranner(tag, "foo", "foo", nil, nil)
		require.Error(t, err)
	})

	t.Run("empty options", func(t *testing.T) {
		_, err := NewTranner(tag, dstNetwork, dstAddress, logger.Test, nil)
		require.NoError(t, err)
	})

	t.Run("invalid local address", func(t *testing.T) {
		opts := Options{
			LocalNetwork: "foo",
			LocalAddress: "foo",
		}
		_, err := NewTranner(tag, dstNetwork, dstAddress, logger.Test, &opts)
		require.Error(t, err)
	})
}

func TestTranner_Start(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("started twice", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		err := tranner.Start()
		require.NoError(t, err)
		err = tranner.Start()
		require.Error(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("failed to listen", func(t *testing.T) {
		dstNetwork := "tcp"
		dstAddress := "127.0.0.1:" + testsuite.HTTPServerPort
		opts := Options{LocalAddress: "0.0.0.1:0"}
		tranner, err := NewTranner("test", dstNetwork, dstAddress, logger.Test, &opts)
		require.NoError(t, err)

		err = tranner.Start()
		require.Error(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})
}

func TestTranner_Stop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		err := tranner.Start()
		require.NoError(t, err)

		// force close tranner
		conn, err := net.Dial("tcp", tranner.testAddress())
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		// wait tran
		time.Sleep(time.Second)

		t.Log(tranner.Status())

		tranner.Stop()
		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("close with error", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		tranner.ctx, tranner.cancel = context.WithCancel(context.Background())
		tranner.listener = testsuite.NewMockListenerWithCloseError()

		conn := &tConn{
			ctx:   tranner,
			local: testsuite.NewMockConnWithCloseError(),
		}
		tranner.trackConn(conn, true)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})
}

func TestTranner_Call(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	t.Run("List", func(t *testing.T) {
		ret, err := tranner.Call("List")
		require.NoError(t, err)
		require.NotNil(t, ret)
	})

	t.Run("Kill", func(t *testing.T) {
		// common
		ret, err := tranner.Call("Kill", "1.1.1.1:443")
		require.NoError(t, err)
		require.NotNil(t, ret)

		// no arguments
		ret, err = tranner.Call("Kill")
		require.Error(t, err)
		require.Nil(t, ret)

		// invalid argument
		ret, err = tranner.Call("Kill", 1)
		require.Error(t, err)
		require.Nil(t, ret)
	})

	t.Run("unknown method", func(t *testing.T) {
		ret, err := tranner.Call("foo")
		require.EqualError(t, err, `unknown method: "foo"`)
		require.Nil(t, ret)
	})

	tranner.Stop()

	testsuite.IsDestroyed(t, tranner)
}

func TestTranner_List(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	err := tranner.Start()
	require.NoError(t, err)

	conn := &tConn{
		ctx:   tranner,
		local: testsuite.NewMockConn(),
	}

	tranner.trackConn(conn, true)

	for _, addr := range tranner.List() {
		fmt.Println(addr)
	}

	tranner.trackConn(conn, false)

	tranner.Stop()

	testsuite.IsDestroyed(t, tranner)
}

func TestTranner_Kill(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	err := tranner.Start()
	require.NoError(t, err)

	conn := &tConn{
		ctx:   tranner,
		local: testsuite.NewMockConn(),
	}

	t.Run("kill all", func(t *testing.T) {
		tranner.trackConn(conn, true)
		defer tranner.trackConn(conn, false)

		for _, addr := range tranner.List() {
			err = tranner.Kill(addr)
			require.NoError(t, err)
		}
	})

	t.Run("not exist", func(t *testing.T) {
		tranner.trackConn(conn, true)
		defer tranner.trackConn(conn, false)

		err = tranner.Kill("foo")
		require.EqualError(t, err, `connection "foo" is not exist`)
	})

	t.Run("kill with close error", func(t *testing.T) {
		c := &tConn{
			ctx:   tranner,
			local: testsuite.NewMockConnWithCloseError(),
		}
		tranner.trackConn(c, true)
		defer tranner.trackConn(c, false)

		err = tranner.Kill(c.local.RemoteAddr().String())
		require.NoError(t, err)
	})

	tranner.Stop()

	testsuite.IsDestroyed(t, tranner)
}

func TestTranner_serve(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("accept panic", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		patch := func(net.Listener, int) net.Listener {
			return testsuite.NewMockListenerWithAcceptPanic()
		}
		pg := monkey.Patch(netutil.LimitListener, patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("failed to accept", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		patch := func(net.Listener, int) net.Listener {
			return testsuite.NewMockListenerWithAcceptError()
		}
		pg := monkey.Patch(netutil.LimitListener, patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		// wait serve() return
		tranner.wg.Wait()

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("close listener error", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		patch := func(net.Listener, int) net.Listener {
			return testsuite.NewMockListenerWithCloseError()
		}
		pg := monkey.Patch(netutil.LimitListener, patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})
}

func TestTranner_trackConn(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	t.Run("failed to add conn", func(t *testing.T) {
		ok := tranner.trackConn(nil, true)
		require.False(t, ok)
	})

	t.Run("add and delete", func(t *testing.T) {
		err := tranner.Start()
		require.NoError(t, err)

		ok := tranner.trackConn(nil, true)
		require.True(t, ok)

		ok = tranner.trackConn(nil, false)
		require.True(t, ok)
	})

	tranner.Stop()

	testsuite.IsDestroyed(t, tranner)
}

func TestTConn_Serve(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to track conn", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		c := testsuite.NewMockConn()
		conn := tranner.newConn(c)
		conn.Serve()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("close local connection error", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		c := testsuite.NewMockConnWithCloseError()
		conn := tranner.newConn(c)
		conn.Serve()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("failed to connect target", func(t *testing.T) {
		dstNetwork := "tcp"
		dstAddress := "0.0.0.0:1"
		opts := Options{LocalAddress: "127.0.0.1:0"}
		tranner, err := NewTranner("test", dstNetwork, dstAddress, logger.Test, &opts)
		require.NoError(t, err)

		err = tranner.Start()
		require.NoError(t, err)

		conn, err := net.Dial("tcp", tranner.testAddress())
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		// wait tran
		time.Sleep(time.Second)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("panic", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		patch := func(context.Context, time.Duration) (context.Context, context.CancelFunc) {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(context.WithTimeout, patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		conn, err := net.Dial("tcp", tranner.testAddress())
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		// wait tran
		time.Sleep(time.Second)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("panic from copy", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		patch := func(io.Writer, io.Reader) (int64, error) {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(io.Copy, patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		conn, err := net.Dial("tcp", tranner.testAddress())
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		// wait tran
		time.Sleep(time.Second)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})

	t.Run("close remote connection error", func(t *testing.T) {
		tranner := testGenerateTranner(t)

		dialer := new(net.Dialer)
		patch := func(interface{}, context.Context, string, string) (net.Conn, error) {
			return testsuite.NewMockConnWithCloseError(), nil
		}
		pg := monkey.PatchInstanceMethod(dialer, "DialContext", patch)
		defer pg.Unpatch()

		err := tranner.Start()
		require.NoError(t, err)

		conn := testsuite.NewMockConn()
		tranner.newConn(conn).Serve()

		// wait tran
		time.Sleep(time.Second)

		tranner.Stop()

		testsuite.IsDestroyed(t, tranner)
	})
}

func TestTranner_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tranner := testGenerateTranner(t)

	start := func() {
		_ = tranner.Start()
	}
	stop := func() {
		tranner.Stop()
	}
	restart := func() {
		_ = tranner.Restart()
	}
	info := func() {
		_ = tranner.Info()
	}
	status := func() {
		_ = tranner.Status()
	}
	track := func() {
		conn := &tConn{
			ctx:   tranner,
			local: testsuite.NewMockConn(),
		}
		tranner.trackConn(conn, true)
	}
	testsuite.RunParallel(100, nil, nil,
		start, stop, restart, info, status, track)

	tranner.Stop()

	testsuite.IsDestroyed(t, tranner)
}
