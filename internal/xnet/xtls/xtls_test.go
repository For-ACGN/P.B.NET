package xtls

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestXTLS(t *testing.T) {
	serverCfg, clientCfg := testsuite.TLSConfigPair(t)
	if testsuite.IPv4Enabled {
		listener, err := Listen("tcp4", "localhost:0", serverCfg, 0)
		require.NoError(t, err)
		addr := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			return Dial("tcp4", addr, clientCfg, 0, nil)
		}, true)
	}

	if testsuite.IPv6Enabled {
		listener, err := Listen("tcp6", "localhost:0", serverCfg, 0)
		require.NoError(t, err)
		addr := listener.Addr().String()
		testsuite.ListenerAndDial(t, listener, func() (net.Conn, error) {
			return Dial("tcp6", addr, clientCfg, 0, nil)
		}, true)
	}
}

func TestXTLSConn(t *testing.T) {
	serverCfg, clientCfg := testsuite.TLSConfigPair(t)
	clientCfg.ServerName = "localhost"

	server, client := net.Pipe()
	server = Server(context.Background(), server, serverCfg, 0)
	client = Client(context.Background(), client, clientCfg, 0)
	testsuite.ConnSC(t, server, client, false)
	testsuite.IsDestroyed(t, server)
	testsuite.IsDestroyed(t, client)

	server, client = net.Pipe()
	server = Server(context.Background(), server, serverCfg, 0)
	client = Client(context.Background(), client, clientCfg, 0)
	testsuite.ConnCS(t, server, client, false)
	testsuite.IsDestroyed(t, server)
	testsuite.IsDestroyed(t, client)
}
