package netmon

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/module/netstat"
	"project/internal/nettool"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestMonitor(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	fmt.Println("Local Address    Remote Address    State    PID")
	handler := func(_ context.Context, event uint8, data interface{}) {
		switch event {
		case EventConnCreated:
			testMonitorPrintCreatedConns(t, data)
		case EventConnClosed:
			testMonitorPrintClosedConns(t, data)
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)

	monitor.Start()
	time.Sleep(2 * time.Second)
	require.True(t, monitor.IsRunning())

	monitor.SetInterval(50 * time.Millisecond)

	err = monitor.SetOptions(nil)
	require.NoError(t, err)

	monitor.Pause()
	time.Sleep(2 * time.Second)
	require.False(t, monitor.IsRunning())

	monitor.Continue()
	time.Sleep(2 * time.Second)
	require.True(t, monitor.IsRunning())

	require.NotEmpty(t, monitor.GetTCP4Conns())
	require.NotEmpty(t, monitor.GetTCP6Conns())
	require.NotEmpty(t, monitor.GetUDP4Conns())
	require.NotEmpty(t, monitor.GetUDP6Conns())

	err = monitor.Close()
	require.NoError(t, err)
	require.False(t, monitor.IsRunning())

	testsuite.IsDestroyed(t, monitor)
}

func testMonitorPrintCreatedConns(t *testing.T, conns interface{}) {
	for _, conn := range conns.([]interface{}) {
		switch conn := conn.(type) {
		case *netstat.TCP4Conn:
			fmt.Printf(
				"create TCP4 connection\n%s %s %d %d\n",
				conn.LocalAddr(), conn.RemoteAddr(), conn.State, conn.PID,
			)
		case *netstat.TCP6Conn:
			fmt.Printf(
				"create TCP6 connection\n%s %s %d %d\n",
				conn.LocalAddr(), conn.RemoteAddr(), conn.State, conn.PID,
			)
		case *netstat.UDP4Conn:
			fmt.Printf(
				"create UDP4 connection\n%s *:* %d\n",
				conn.LocalAddr(), conn.PID,
			)
		case *netstat.UDP6Conn:
			fmt.Printf(
				"create UDP6 connection\n%s *:* %d\n",
				conn.LocalAddr(), conn.PID,
			)
		default:
			t.Fatal("invalid structure:", conn)
		}
	}
}

func testMonitorPrintClosedConns(t *testing.T, conns interface{}) {
	for _, conn := range conns.([]interface{}) {
		switch conn := conn.(type) {
		case *netstat.TCP4Conn:
			fmt.Printf("close TCP4 connection\n%s:%d %s:%d %d %d\n",
				conn.LocalIP, conn.LocalPort,
				conn.RemoteIP, conn.RemotePort,
				conn.State, conn.PID,
			)
		case *netstat.TCP6Conn:
			fmt.Printf("close TCP6 connection\n[%s%%%d]:%d [%s%%%d]:%d %d %d\n",
				conn.LocalIP, conn.LocalScopeID, conn.LocalPort,
				conn.RemoteIP, conn.RemoteScopeID, conn.RemotePort,
				conn.State, conn.PID,
			)
		case *netstat.UDP4Conn:
			fmt.Printf("close UDP4 connection\n%s:%d *:* %d\n",
				conn.LocalIP, conn.LocalPort, conn.PID,
			)
		case *netstat.UDP6Conn:
			fmt.Printf("close UDP6 connection\n[%s%%%d]:%d *:* %d\n",
				conn.LocalIP, conn.LocalScopeID, conn.LocalPort, conn.PID,
			)
		default:
			t.Fatal("invalid structure:", conn)
		}
	}
}

func TestMonitor_EventConnCreated(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tcp4Listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = tcp4Listener.Close() }()

	tcp6Listener, err := net.Listen("tcp6", "[::1]:0")
	require.NoError(t, err)
	defer func() { _ = tcp6Listener.Close() }()

	tcp4Addr := tcp4Listener.Addr().String()
	tcp6Addr := tcp6Listener.Addr().String()

	fmt.Println("tcp4 listener:", tcp4Addr)
	fmt.Println("tcp6 listener:", tcp6Addr)

	var (
		findTCP4    bool
		findTCP6    bool
		findUDP4    bool
		findUDP6    bool
		createdUDP4 []string
		createdUDP6 []string
	)

	handler := func(_ context.Context, event uint8, data interface{}) {
		if event != EventConnCreated {
			return
		}
		for _, conn := range data.([]interface{}) {
			switch conn := conn.(type) {
			case *netstat.TCP4Conn:
				remoteAddr := nettool.JoinHostPort(conn.RemoteIP.String(), conn.RemotePort)
				fmt.Println("created tcp4 connection:", remoteAddr)
				if remoteAddr == tcp4Addr {
					findTCP4 = true
				}
			case *netstat.TCP6Conn:
				remoteAddr := nettool.JoinHostPort(conn.RemoteIP.String(), conn.RemotePort)
				fmt.Println("created tcp6 connection:", remoteAddr)
				if remoteAddr == tcp6Addr {
					findTCP6 = true
				}
			case *netstat.UDP4Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("created udp4 connection:", localAddr)
				createdUDP4 = append(createdUDP4, localAddr)
			case *netstat.UDP6Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("created udp6 connection:", localAddr)
				createdUDP6 = append(createdUDP6, localAddr)
			default:
				t.Fatal("invalid structure:", conn)
			}
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)
	monitor.Start()

	// wait first auto refresh
	time.Sleep(2 * defaultRefreshInterval)

	// connect test tcp listeners
	tcp4Conn, err := net.Dial("tcp4", tcp4Addr)
	require.NoError(t, err)
	defer func() { _ = tcp4Conn.Close() }()

	tcp6Conn, err := net.Dial("tcp6", tcp6Addr)
	require.NoError(t, err)
	defer func() { _ = tcp6Conn.Close() }()

	// listen test udp connection
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:123")
	require.NoError(t, err)
	udp4Conn, err := net.ListenUDP("udp4", udpAddr)
	require.NoError(t, err)
	defer func() { _ = udp4Conn.Close() }()

	udpAddr, err = net.ResolveUDPAddr("udp6", "[::1]:123")
	require.NoError(t, err)
	udp6Conn, err := net.ListenUDP("udp6", udpAddr)
	require.NoError(t, err)
	defer func() { _ = udp6Conn.Close() }()

	udp4ConnAddr := udp4Conn.LocalAddr().String()
	udp6ConnAddr := udp6Conn.LocalAddr().String()

	fmt.Println("udp4 listener:", udp4ConnAddr)
	fmt.Println("udp6 listener:", udp6ConnAddr)

	// wait refresh
	time.Sleep(2 * defaultRefreshInterval)

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)

	// check
	for i := 0; i < len(createdUDP4); i++ {
		if createdUDP4[i] == udp4ConnAddr {
			findUDP4 = true
			break
		}
	}
	for i := 0; i < len(createdUDP6); i++ {
		if createdUDP6[i] == udp6ConnAddr {
			findUDP6 = true
			break
		}
	}
	require.True(t, findTCP4, "not find expected tcp4 connection")
	require.True(t, findTCP6, "not find expected tcp6 connection")
	require.True(t, findUDP4, "not find expected udp4 connection")
	require.True(t, findUDP6, "not find expected udp6 connection")
}

func TestMonitor_EventConnClosed(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	tcp4Listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = tcp4Listener.Close() }()

	tcp6Listener, err := net.Listen("tcp6", "[::1]:0")
	require.NoError(t, err)
	defer func() { _ = tcp6Listener.Close() }()

	// listen test udp connection
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:123")
	require.NoError(t, err)
	udp4Conn, err := net.ListenUDP("udp4", udpAddr)
	require.NoError(t, err)
	defer func() { _ = udp4Conn.Close() }()

	udpAddr, err = net.ResolveUDPAddr("udp6", "[::1]:123")
	require.NoError(t, err)
	udp6Conn, err := net.ListenUDP("udp6", udpAddr)
	require.NoError(t, err)
	defer func() { _ = udp6Conn.Close() }()

	tcp4Addr := tcp4Listener.Addr().String()
	tcp6Addr := tcp6Listener.Addr().String()

	udp4ConnAddr := udp4Conn.LocalAddr().String()
	udp6ConnAddr := udp6Conn.LocalAddr().String()

	fmt.Println("tcp4 listener:", tcp4Addr)
	fmt.Println("tcp6 listener:", tcp6Addr)
	fmt.Println("udp4 listener:", udp4ConnAddr)
	fmt.Println("udp6 listener:", udp6ConnAddr)

	var (
		findTCP4 bool
		findTCP6 bool
		findUDP4 bool
		findUDP6 bool
	)

	handler := func(_ context.Context, event uint8, data interface{}) {
		if event != EventConnClosed {
			return
		}
		for _, conn := range data.([]interface{}) {
			switch conn := conn.(type) {
			case *netstat.TCP4Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("close tcp4 connection:", localAddr)
				if localAddr == tcp4Addr {
					findTCP4 = true
				}
			case *netstat.TCP6Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("close tcp6 connection:", localAddr)
				if localAddr == tcp6Addr {
					findTCP6 = true
				}
			case *netstat.UDP4Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("close udp4 connection:", localAddr)
				if localAddr == udp4ConnAddr {
					findUDP4 = true
				}
			case *netstat.UDP6Conn:
				localAddr := nettool.JoinHostPort(conn.LocalIP.String(), conn.LocalPort)
				fmt.Println("close udp6 connection:", localAddr)
				if localAddr == udp6ConnAddr {
					findUDP6 = true
				}
			default:
				t.Fatal("invalid structure:", conn)
			}
		}
	}
	monitor, err := New(logger.Test, handler, nil)
	require.NoError(t, err)
	monitor.Start()

	err = tcp4Listener.Close()
	require.NoError(t, err)
	err = tcp6Listener.Close()
	require.NoError(t, err)
	err = udp4Conn.Close()
	require.NoError(t, err)
	err = udp6Conn.Close()
	require.NoError(t, err)

	// wait auto refresh
	time.Sleep(2 * defaultRefreshInterval)

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)

	require.True(t, findTCP4, "not find expected tcp4 connection")
	require.True(t, findTCP6, "not find expected tcp6 connection")
	require.True(t, findUDP4, "not find expected udp4 connection")
	require.True(t, findUDP6, "not find expected udp6 connection")
}

func TestMonitor_refreshLoop(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to refresh", func(t *testing.T) {
		monitor, err := New(logger.Test, nil, nil)
		require.NoError(t, err)

		var m *Monitor
		patch := func(interface{}) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(m, "Refresh", patch)
		defer pg.Unpatch()

		monitor.Start()

		// wait restart
		time.Sleep(3 * time.Second)

		err = monitor.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, monitor)
	})

	t.Run("panic", func(t *testing.T) {
		monitor, err := New(logger.Test, nil, nil)
		require.NoError(t, err)

		var m *Monitor
		patch := func(interface{}) error {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(m, "Refresh", patch)
		defer pg.Unpatch()

		monitor.Start()

		// wait restart
		time.Sleep(3 * time.Second)

		err = monitor.Close()
		require.NoError(t, err)

		testsuite.IsDestroyed(t, monitor)
	})
}
