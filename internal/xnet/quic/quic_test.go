package quic

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestListenAndDial(t *testing.T) {
	serverCfg, clientCfg := testsuite.TLSConfigPair(t)
	if testsuite.IPv4Enabled {
		const network = "udp4"
		listener, err := Listen(network, "127.0.0.1:0", serverCfg, 0)
		require.NoError(t, err)
		address := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			return Dial(network, address, clientCfg.Clone(), 0)
		}, true)
	}

	if testsuite.IPv6Enabled {
		const network = "udp6"
		listener, err := Listen(network, "[::1]:0", serverCfg, 0)
		require.NoError(t, err)
		address := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			return Dial(network, address, clientCfg.Clone(), 0)
		}, true)
	}
}

func TestListenAndDialContext(t *testing.T) {
	serverCfg, clientCfg := testsuite.TLSConfigPair(t)
	if testsuite.IPv4Enabled {
		const network = "udp4"
		listener, err := Listen(network, "127.0.0.1:0", serverCfg, 0)
		require.NoError(t, err)
		address := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return DialContext(ctx, network, address, clientCfg.Clone(), 0)
		}, true)
	}

	if testsuite.IPv6Enabled {
		const network = "udp6"
		listener, err := Listen(network, "[::1]:0", serverCfg, 0)
		require.NoError(t, err)
		address := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return DialContext(ctx, network, address, clientCfg.Clone(), 0)
		}, true)
	}
}

func TestConn_Close(t *testing.T) {
	serverCfg, clientCfg := testsuite.TLSConfigPair(t)
	listener, err := Listen("udp4", "localhost:0", serverCfg, 0)
	require.NoError(t, err)
	address := listener.Addr().String()
	var server net.Conn
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		server, err = listener.Accept()
		require.NoError(t, err)
	}()
	client, err := Dial("udp4", address, clientCfg, 0)
	require.NoError(t, err)
	wg.Wait()
	wg.Add(8)
	// server
	go func() {
		defer wg.Done()
		_ = server.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = server.Write(testsuite.Bytes())
	}()
	go func() {
		defer wg.Done()
		_ = server.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = server.Write(testsuite.Bytes())
	}()
	// client
	go func() {
		defer wg.Done()
		_ = client.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = client.Write(testsuite.Bytes())
	}()
	go func() {
		defer wg.Done()
		_ = client.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = client.Write(testsuite.Bytes())
	}()
	wg.Wait()

	testsuite.IsDestroyed(t, client)
	testsuite.IsDestroyed(t, server)
}
