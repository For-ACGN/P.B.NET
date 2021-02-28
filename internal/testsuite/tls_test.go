package testsuite

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTLSConfigPair(t *testing.T) {
	gm := MarkGoroutines(t)
	defer gm.Compare()

	serverCfg, clientCfg := TLSConfigPair(t, "127.0.0.1")

	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	require.NoError(t, err)
	address := listener.Addr().String()

	ListenerAndDial(t, listener, func() (net.Conn, error) {
		return tls.Dial("tcp", address, clientCfg.Clone())
	}, true)
}
