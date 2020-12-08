package testsuite

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/nettool"
	"project/internal/patch/monkey"
)

func TestPrintNetworkInfo(t *testing.T) {
	patch := func() (bool, bool) {
		return false, false
	}
	pg := monkey.Patch(nettool.IPEnabled, patch)
	defer pg.Unpatch()

	printNetworkInfo()
}

func TestDeployPprofHTTPServer(t *testing.T) {
	defer DeferForPanic(t)

	patch := func(string, string) (net.Listener, error) {
		return nil, monkey.Error
	}
	pg := monkey.Patch(net.Listen, patch)
	defer pg.Unpatch()

	deployPprofHTTPServer()
}

func TestStartPprofHTTPServer(t *testing.T) {
	t.Run("tcp4", func(t *testing.T) {
		patch := func(string, string) (net.Listener, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(net.Listen, patch)
		defer pg.Unpatch()

		ok := startPprofHTTPServer(nil, 123)
		require.False(t, ok)
	})

	t.Run("tcp6", func(t *testing.T) {
		patch := func(network, address string) (net.Listener, error) {
			if network == "tcp6" {
				return nil, monkey.Error
			}
			return nil, nil
		}
		pg := monkey.Patch(net.Listen, patch)
		defer pg.Unpatch()

		ok := startPprofHTTPServer(nil, 123)
		require.False(t, ok)
	})
}

func TestIsInGoland(t *testing.T) {
	t.Log("in Goland:", InGoland)
}

func TestBytes(t *testing.T) {
	Bytes()
}

func TestDeferForPanic(t *testing.T) {
	defer DeferForPanic(t)

	panic("test panic")
}

func TestRunGoroutines(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		done := make(chan struct{})
		RunGoroutines(func() {
			close(done)
		})
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("run goroutine timeout")
		}
	})

	t.Run("no function", func(t *testing.T) {
		RunGoroutines()
	})
}

func TestRunMultiTimes(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	test := 0
	mu := sync.Mutex{}

	f1 := func() {
		mu.Lock()
		defer mu.Unlock()

		test++
		fmt.Println("f1:", test)
	}
	f2 := func() {
		mu.Lock()
		defer mu.Unlock()

		test += 2
		fmt.Println("f2:", test)
	}
	f3 := func() {
		mu.Lock()
		defer mu.Unlock()

		test += 3
		fmt.Println("f3:", test)
	}

	t.Run("ok", func(t *testing.T) {
		RunMultiTimes(5, f1, f2, f3)

		require.Equal(t, (1+2+3)*5, test)
	})

	t.Run("no functions", func(t *testing.T) {
		RunMultiTimes(1)
	})

	t.Run("invalid times", func(t *testing.T) {
		test = 0

		RunMultiTimes(-1, f3)

		require.Equal(t, 3*100, test)
	})
}

func TestRunParallel(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	test := 0
	mu := sync.Mutex{}

	init := func() {
		test = 0
		fmt.Println("init")
	}
	f1 := func() {
		mu.Lock()
		defer mu.Unlock()

		test++
		fmt.Println("f1:", test)
	}
	f2 := func() {
		mu.Lock()
		defer mu.Unlock()

		test += 2
		fmt.Println("f2:", test)
	}
	f3 := func() {
		mu.Lock()
		defer mu.Unlock()

		test += 3
		fmt.Println("f3:", test)
	}
	cleanup := func() {
		fmt.Println("cleanup")
	}

	t.Run("ok", func(t *testing.T) {
		RunParallel(5, init, cleanup, f1, f2, f3)

		require.Equal(t, 1+2+3, test)
	})

	t.Run("no functions", func(t *testing.T) {
		RunParallel(1, nil, nil)
	})

	t.Run("invalid times", func(t *testing.T) {
		test = 0

		RunParallel(-1, nil, nil, f3)

		require.Equal(t, 3*100, test)
	})
}

func TestRunHTTPServer(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	t.Run("http", func(t *testing.T) {
		server := http.Server{Addr: "localhost:0"}
		port := RunHTTPServer(t, "tcp", &server)
		defer func() { _ = server.Close() }()
		t.Log("http server port:", port)

		client := http.Client{Transport: new(http.Transport)}
		defer client.CloseIdleConnections()
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/", port))
		require.NoError(t, err)
		_, err = io.Copy(ioutil.Discard, resp.Body)
		require.NoError(t, err)
	})

	t.Run("https", func(t *testing.T) {
		serverCfg, clientCfg := TLSConfigPair(t, "127.0.0.1")
		server := http.Server{
			Addr:      "localhost:0",
			TLSConfig: serverCfg,
		}
		port := RunHTTPServer(t, "tcp", &server)
		defer func() { _ = server.Close() }()
		t.Log("https server port:", port)

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: clientCfg,
			},
		}
		defer client.CloseIdleConnections()
		resp, err := client.Get(fmt.Sprintf("https://localhost:%s/", port))
		require.NoError(t, err)
		_, err = io.Copy(ioutil.Discard, resp.Body)
		require.NoError(t, err)
	})
}
