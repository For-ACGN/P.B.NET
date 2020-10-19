package testsuite

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/nettool"
	"project/internal/xpanic"
)

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
	deployPprofHTTPServer()
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

func deployPprofHTTPServer() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/debug/pprof/", pprof.Index)
	serveMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	serveMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	serveMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	serveMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := &http.Server{Handler: serveMux}
	for port := 9931; port < 65536; port++ {
		if startPprofHTTPServer(server, port) {
			fmt.Printf("[debug] pprof http server port: %d\n", port)
			return
		}
	}
	panic("failed to deploy pprof http server")
}

func startPprofHTTPServer(server *http.Server, port int) bool {
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

// CheckErrorInTestMain is used to check error in function TestMain(),
// because no t *testing.T, so we need check it self.
func CheckErrorInTestMain(err error) {
	if err != nil {
		panic(fmt.Sprintf("error occoured in TestMain:\n%s", err))
	}
}

// CheckOptions is used to check unmarshal is apply value to each field,
// it will check each field value is zero.
func CheckOptions(t *testing.T, v interface{}) {
	str := checkOptions("", v)
	require.True(t, str == "", str)
}

func checkOptions(father string, v interface{}) (result string) {
	ok, result := checkSpecialType(father, v)
	if ok {
		return
	}
	typ := reflect.TypeOf(v)
	var value reflect.Value
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "checkOptions")
			result = fmt.Sprint(father+typ.Name(), " with panic occurred")
		}
	}()
	if typ.Kind() == reflect.Ptr {
		value = reflect.ValueOf(v)
		typ = value.Type()
		if value.IsNil() { // check is nil point
			return father + typ.Name() + " is nil point"
		}
		value = value.Elem()
		typ = value.Type()
	} else {
		value = reflect.ValueOf(v)
	}
	return walkOptions(father, typ, value)
}

func checkSpecialType(father string, v interface{}) (bool, string) {
	var typ string
	switch val := v.(type) {
	case *time.Time:
		if val != nil && !val.IsZero() {
			return true, ""
		}
		typ = "time.Time"
	case time.Time:
		if !val.IsZero() {
			return true, ""
		}
		typ = "time.Time"
	default:
		return false, ""
	}
	if father == "" {
		return true, typ + " is zero value"
	}
	return true, father + " is zero value"
}

func walkOptions(father string, typ reflect.Type, value reflect.Value) string {
	for i := 0; i < value.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := value.Field(i)
		// skip unexported field
		if fieldType.PkgPath != "" && !fieldType.Anonymous {
			continue
		}
		// skip filed with check tag
		fieldTag, ok := fieldType.Tag.Lookup("check")
		if ok {
			if fieldTag == "" {
				panic(fmt.Sprintf("empty value in check tag, %s", fieldType.Name))
			}
			if fieldTag == "-" {
				continue
			}
		}
		switch fieldType.Type.Kind() {
		case reflect.Struct, reflect.Ptr:
			var f string
			if father == "" {
				f = typ.Name() + "." + fieldType.Name
			} else {
				f = father + "." + fieldType.Name
			}
			result := checkOptions(f, fieldValue.Interface())
			if result != "" {
				return result
			}
		case reflect.Chan, reflect.Func, reflect.Interface,
			reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
			continue
		default:
			if !fieldValue.IsZero() {
				continue
			}
			const format = "%s.%s is zero value"
			if father == "" {
				return fmt.Sprintf(format, typ.Name(), fieldType.Name)
			}
			return fmt.Sprintf(format, father, fieldType.Name)
		}
	}
	return ""
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

// RunParallel is used to call functions with go func().
// object with Add(), Get() ... need it for test with race.
func RunParallel(times int, init, cleanup func(), fns ...func()) {
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
