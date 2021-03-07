package testsuite

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/nettool"
)

// InvalidFilePath is a invalid file path.
const InvalidFilePath = "testdata/<</file"

var (
	// IPv4Enabled is used to tell tests current system is enable IPv4.
	IPv4Enabled bool

	// IPv6Enabled is used to tell tests current system is enable IPv6.
	IPv6Enabled bool

	// InGoland is used to tell tests in run by Goland.
	InGoland bool
)

func init() {
	printNetworkInfo()
	deployPPROFHTTPServer()
	isInGoland()
}

func printNetworkInfo() {
	IPv4Enabled, IPv6Enabled = nettool.IPEnabled()
	if IPv4Enabled || IPv6Enabled {
		const format = "[debug] network: IPv4-%t IPv6-%t"
		str := fmt.Sprintf(format, IPv4Enabled, IPv6Enabled)
		str = strings.ReplaceAll(str, "true", "Enabled")
		str = strings.ReplaceAll(str, "false", "Disabled")
		fmt.Println(str)
		return
	}
	fmt.Println("[debug] network unavailable")
}

func deployPPROFHTTPServer() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/debug/pprof/", pprof.Index)
	serveMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	serveMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	serveMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	serveMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := &http.Server{Handler: serveMux}
	for port := 9931; port < 65536; port++ {
		if startPPROFHTTPServer(server, port) {
			fmt.Printf("[debug] pprof http server port: %d\n", port)
			return
		}
	}
	panic("failed to deploy pprof http server")
}

func startPPROFHTTPServer(server *http.Server, port int) bool {
	var (
		ipv4 net.Listener
		ipv6 net.Listener
		err  error
	)
	address := fmt.Sprintf("localhost:%d", port)
	ipv4, err = net.Listen("tcp4", address)
	if err != nil {
		return false
	}
	ipv6, err = net.Listen("tcp6", address)
	if err != nil {
		return false
	}
	RunGoroutines(
		func() { _ = server.Serve(ipv4) },
		func() { _ = server.Serve(ipv6) },
	)
	return true
}

func isInGoland() {
	for _, value := range os.Environ() {
		if strings.Contains(value, "IDEA") {
			InGoland = true
			break
		}
	}
}

// TestDataSize is the size of Bytes().
const TestDataSize = 256

// Bytes is used to generate test data: []byte{0, 1, .... 254, 255}.
func Bytes() []byte {
	testdata := make([]byte, TestDataSize)
	for i := 0; i < TestDataSize; i++ {
		testdata[i] = byte(i)
	}
	return testdata
}

// DeferForPanic is used to add recover and log panic in defer function,
// it used to some tests like this:
//
// defer func() {
//      r := recover()
//      require.NotNil(t, r)
//      t.Log(r)
// }()
func DeferForPanic(t testing.TB) {
	r := recover()
	require.NotNil(t, r)
	t.Logf("\npanic in %s:\n%s\n", t.Name(), r)
}

// RunGoroutines is used to make sure goroutine is running.
// Because when you call "go" maybe this goroutine is not in running.
// Usually use it with testsuite.MarkGoroutines().
func RunGoroutines(fns ...func()) {
	l := len(fns)
	if l == 0 {
		return
	}
	done := make(chan struct{}, l)
	for i := 0; i < l; i++ {
		go func(i int) {
			done <- struct{}{}
			fns[i]()
		}(i)
	}
	for i := 0; i < l; i++ {
		<-done
	}
}

// RunMultiTimes is used to call functions with n times in the same time.
func RunMultiTimes(times int, fns ...func()) {
	l := len(fns)
	if l == 0 {
		return
	}
	if times < 1 || times > 1000 {
		times = 100
	}
	wg := sync.WaitGroup{}
	for i := 0; i < l; i++ {
		for j := 0; j < times; j++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				fns[i]()
				// trigger data race better
				time.Sleep(10 * time.Millisecond)
			}(i)
		}
	}
	wg.Wait()
}

// RunParallelTest is used to run parallel for test with race.. init will be called
// before each test, and cleanup will be called after all functions returned.
func RunParallelTest(times int, init, cleanup func(), fns ...func()) {
	l := len(fns)
	if l == 0 {
		return
	}
	if times < 1 || times > 1000 {
		times = 100
	}
	wg := sync.WaitGroup{}
	for i := 0; i < times; i++ {
		// initialize before call
		if init != nil {
			init()
		}
		// call functions
		for j := 0; j < l; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				fns[j]()
				// trigger data race better
				time.Sleep(10 * time.Millisecond)
			}(j)
		}
		wg.Wait()
		// clean after call
		if cleanup != nil {
			cleanup()
		}
	}
}

// RunHTTPServer is used to start a http or https server and return port.
func RunHTTPServer(t testing.TB, network string, server *http.Server) string {
	listener, err := net.Listen(network, server.Addr)
	require.NoError(t, err)
	// start serve
	RunGoroutines(func() {
		if server.TLSConfig != nil {
			_ = server.ServeTLS(listener, "", "")
		} else {
			_ = server.Serve(listener)
		}
	})
	// get port
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	return port
}
