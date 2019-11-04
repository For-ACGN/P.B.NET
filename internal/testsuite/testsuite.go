package testsuite

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	enableIPv4 bool
	enableIPv6 bool
)

func init() {
	initGetIPv4Address()
	initGetIPv6Address()

	// check IPv4
	for i := 0; i < 5; i++ {
		addr := getIPv4Address()
		conn, err := net.DialTimeout("tcp4", addr, 5*time.Second)
		if err == nil {
			_ = conn.Close()
			enableIPv4 = true
			break
		}
	}

	// check IPv6
	for i := 0; i < 5; i++ {
		addr := getIPv6Address()
		conn, err := net.DialTimeout("tcp6", addr, 5*time.Second)
		if err == nil {
			_ = conn.Close()
			enableIPv6 = true
			break
		}
	}

	// check network
	if !enableIPv4 && !enableIPv6 {
		fmt.Print("network unavailable")
		os.Exit(1)
	}

	// deploy pprof
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/debug/pprof/", pprof.Index)
	serverMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	serverMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	serverMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	serverMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := http.Server{Handler: serverMux}
	var (
		listener net.Listener
		err      error
	)
	listener, err = net.Listen("tcp", "localhost:9931")
	if err != nil {
		listener, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			fmt.Printf("failed to deploy pprof: %s\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("[debug] pprof: %s\n", listener.Addr())
	go func() { _ = server.Serve(listener) }()
}

// EnableIPv4 is used to determine whether IPv4 is available
func EnableIPv4() bool {
	return enableIPv4
}

// EnableIPv6 is used to determine whether IPv6 is available
func EnableIPv6() bool {
	return enableIPv6
}

func isDestroyed(object interface{}) bool {
	destroyed := make(chan struct{})
	runtime.SetFinalizer(object, func(_ interface{}) {
		close(destroyed)
	})
	// total 3 second
	for i := 0; i < 12; i++ {
		runtime.GC()
		select {
		case <-destroyed:
			return true
		case <-time.After(250 * time.Millisecond):
		}
	}
	return false
}

// IsDestroyed is used to check if the object has been recycled by the GC
func IsDestroyed(t testing.TB, object interface{}) {
	require.True(t, isDestroyed(object), "object not destroyed")
}

// RunHTTPServer is used to start a http or https server
func RunHTTPServer(t testing.TB, network string, server *http.Server) string {
	listener, err := net.Listen(network, server.Addr)
	require.NoError(t, err)
	// run
	go func() {
		if server.TLSConfig != nil {
			_ = server.ServeTLS(listener, "", "")
		} else {
			_ = server.Serve(listener)
		}
	}()
	// get port
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	return port
}

// Bytes is used to generate test data: []byte{0, 1, .... 254, 255}
func Bytes() []byte {
	testdata := make([]byte, 256)
	for i := 0; i < 256; i++ {
		testdata[i] = byte(i)
	}
	return testdata
}
