package testsuite

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/nettool"
	"project/internal/xsync"
)

func TestInitHTTPServers(t *testing.T) {
	IPv4Enabled = false
	IPv6Enabled = true
	defer func() { IPv4Enabled, IPv6Enabled = nettool.IPEnabled() }()

	initHTTPServers(t)
}

type mockProxyServer struct {
	listeners  map[*net.Listener]struct{}
	inShutdown int32
	rwm        sync.RWMutex

	counter xsync.Counter
}

func newMockProxyServer() *mockProxyServer {
	return &mockProxyServer{
		listeners: make(map[*net.Listener]struct{}),
	}
}

func (mps *mockProxyServer) shuttingDown() bool {
	return atomic.LoadInt32(&mps.inShutdown) != 0
}

func (mps *mockProxyServer) trackListener(listener *net.Listener, add bool) bool {
	mps.rwm.Lock()
	defer mps.rwm.Unlock()
	if add {
		if mps.shuttingDown() {
			return false
		}
		mps.listeners[listener] = struct{}{}
		mps.counter.Add(1)
	} else {
		delete(mps.listeners, listener)
		mps.counter.Done()
	}
	return true
}

func (mps *mockProxyServer) ListenAndServe(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	return mps.Serve(listener)
}

func (mps *mockProxyServer) Serve(listener net.Listener) error {
	if !mps.trackListener(&listener, true) {
		return errors.New("mock proxy server closed")
	}
	defer mps.trackListener(&listener, false)
	for {
		conn, err := listener.Accept()
		if err != nil {
			if nettool.IsNetClosedError(err) {
				return nil
			}
			return err
		}
		_ = conn.Close()
	}
}

func (mps *mockProxyServer) Addresses() []net.Addr {
	mps.rwm.RLock()
	defer mps.rwm.RUnlock()
	addresses := make([]net.Addr, 0, len(mps.listeners))
	for listener := range mps.listeners {
		addresses = append(addresses, (*listener).Addr())
	}
	return addresses
}

func (mps *mockProxyServer) Info() string {
	return "mock proxy server information"
}

func (mps *mockProxyServer) Close() error {
	err := mps.close()
	mps.counter.Wait()
	return err
}

func (mps *mockProxyServer) close() error {
	atomic.StoreInt32(&mps.inShutdown, 1)
	var err error
	mps.rwm.Lock()
	defer mps.rwm.Unlock()
	// close all listeners
	for listener := range mps.listeners {
		e := (*listener).Close()
		if e != nil && !nettool.IsNetClosedError(e) && err == nil {
			err = e
		}
		delete(mps.listeners, listener)
	}
	return err
}

func TestWaitProxyServerServe(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	const (
		network = "tcp"
		address = "127.0.0.1:0"
	)

	t.Run("ok", func(t *testing.T) {
		server := newMockProxyServer()
		go func() {
			err := server.ListenAndServe(network, address)
			require.NoError(t, err)
		}()
		go func() {
			err := server.ListenAndServe(network, address)
			require.NoError(t, err)
		}()

		WaitProxyServerServe(t, server, 2)

		err := server.Close()
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		server := newMockProxyServer()
		go func() {
			err := server.ListenAndServe(network, address)
			require.NoError(t, err)
		}()

		ok := waitProxyServerServe(server, 2)
		require.False(t, ok, "wait proxy server serve not failed")

		err := server.Close()
		require.NoError(t, err)
	})
}

func TestNopCloser(t *testing.T) {
	closer := NewNopCloser()
	closer.Get()
	_ = closer.Close()
}
