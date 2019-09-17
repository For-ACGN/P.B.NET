package quic

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/cert"
	"project/internal/xnet/testdata"
)

func TestQUIC(t *testing.T) {
	// generate cert
	certConfig := &cert.Config{
		DNSNames:    []string{"localhost"},
		IPAddresses: []string{"127.0.0.1", "::1"},
	}
	c, k, err := cert.Generate(nil, nil, certConfig)
	require.NoError(t, err)
	tlsCert, err := tls.X509KeyPair(c, k)
	require.NoError(t, err)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	listener, err := Listen("udp", "localhost:0", tlsConfig, 0)
	require.NoError(t, err)
	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		write := func() {
			data := testdata.GenerateData()
			_, err = conn.Write(data)
			require.NoError(t, err)
			// check data is changed after write
			require.Equal(t, testdata.GenerateData(), data)
		}
		read := func() {
			data := make([]byte, 256)
			_, err = io.ReadFull(conn, data)
			require.NoError(t, err)
			require.Equal(t, testdata.GenerateData(), data)
		}
		read()
		write()
		write()
		read()
	}()
	// client
	x509Cert, err := cert.Parse(c)
	require.NoError(t, err)
	tlsConfig = &tls.Config{
		RootCAs: x509.NewCertPool(),
	}
	tlsConfig.RootCAs.AddCert(x509Cert)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	conn, err := Dial("udp", "localhost:"+port, tlsConfig, 0)
	require.NoError(t, err)
	write := func() {
		data := testdata.GenerateData()
		_, err = conn.Write(data)
		require.NoError(t, err)
		// check data is changed after write
		require.Equal(t, testdata.GenerateData(), data)
	}
	read := func() {
		data := make([]byte, 256)
		_, err = io.ReadFull(conn, data)
		require.NoError(t, err)
		require.Equal(t, testdata.GenerateData(), data)
	}
	write()
	read()
	read()
	write()
}
