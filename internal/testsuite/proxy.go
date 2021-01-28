package testsuite

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/cert/certpool"
)

// for test proxy client.
type proxyClient interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
	Connect(ctx context.Context, conn net.Conn, network, address string) (net.Conn, error)
	HTTP(t *http.Transport)
	Timeout() time.Duration
	Server() (network string, address string)
	Info() string
}

// for test proxy server.
type proxyServer interface {
	ListenAndServe(network, address string) error
	Serve(listener net.Listener) error
	Addresses() []net.Addr
	Info() string
	Close() error
}

// HTTPServerPort is the test HTTP server port, some tests in internal/proxy need it.
var (
	HTTPServerPort  string
	HTTPSServerPort string
)

// testHandlerData is the test http server handler returned data.
var testHandlerData = []byte("hello")

var (
	httpServer  http.Server
	httpsServer http.Server

	// Root CA about server side
	httpsServerCA *x509.Certificate

	// certificate about client side
	httpsClientCert tls.Certificate

	initHTTPServerOnce sync.Once
)

// InitHTTPServers is used to create http test servers.
// Must call it before testsuite.MarkGoroutines().
func InitHTTPServers(t testing.TB) {
	initHTTPServerOnce.Do(func() { initHTTPServers(t) })
}

func initHTTPServers(t testing.TB) {
	// set handler
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/t", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(testHandlerData)
	})

	// initialize HTTP server
	httpServer.Handler = serveMux

	// initialize HTTPS server
	httpsServer.Handler = serveMux

	// server side certificate
	caASN1, certPEMBlock, keyPEMBlock := TLSCertificate(t, "127.0.0.1")
	serverCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	require.NoError(t, err)
	// require client certificate
	httpsServer.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	httpsServerCA, err = x509.ParseCertificate(caASN1)
	require.NoError(t, err)

	// client side certificate
	caASN1, certPEMBlock, keyPEMBlock = TLSCertificate(t, "127.0.0.1")
	httpsClientCert, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	require.NoError(t, err)
	httpsServer.TLSConfig.ClientCAs = x509.NewCertPool()
	httpsClientCA, err := x509.ParseCertificate(caASN1)
	require.NoError(t, err)
	httpsServer.TLSConfig.ClientCAs.AddCert(httpsClientCA)

	// start HTTP and HTTPS Server Listeners
	var (
		l1 net.Listener // HTTP  IPv4
		l2 net.Listener // HTTPS IPv4
		l3 net.Listener // HTTP  IPv6
		l4 net.Listener // HTTPS IPv6
	)

	if IPv4Enabled {
		l1, err = net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		_, HTTPServerPort, err = net.SplitHostPort(l1.Addr().String())
		require.NoError(t, err)

		l2, err = net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		_, HTTPSServerPort, err = net.SplitHostPort(l2.Addr().String())
		require.NoError(t, err)

		RunGoroutines(
			func() { _ = httpServer.Serve(l1) },
			func() { _ = httpsServer.ServeTLS(l2, "", "") },
		)
	}

	if IPv6Enabled {
		if HTTPServerPort != "" {
			l3, err = net.Listen("tcp", "[::1]:"+HTTPServerPort)
			require.NoError(t, err)

			l4, err = net.Listen("tcp", "[::1]:"+HTTPSServerPort)
			require.NoError(t, err)
		} else { // IPv6 Only
			l3, err = net.Listen("tcp", "[::1]:0")
			require.NoError(t, err)
			_, HTTPServerPort, err = net.SplitHostPort(l3.Addr().String())
			require.NoError(t, err)

			l4, err = net.Listen("tcp", "[::1]:0")
			require.NoError(t, err)
			_, HTTPSServerPort, err = net.SplitHostPort(l4.Addr().String())
			require.NoError(t, err)
		}

		RunGoroutines(
			func() { _ = httpServer.Serve(l3) },
			func() { _ = httpsServer.ServeTLS(l4, "", "") },
		)
	}

	// print proxy server addresses
	fmt.Printf("[debug] HTTP Server Port:  %s\n", HTTPServerPort)
	fmt.Printf("[debug] HTTPS Server Port: %s\n", HTTPSServerPort)
}

// WaitProxyServerServe is used to wait proxy server until is serving.
func WaitProxyServerServe(t *testing.T, server proxyServer, addressNum int) {
	ok := waitProxyServerServe(server, addressNum)
	require.True(t, ok, "wait proxy server serve timeout")
}

func waitProxyServerServe(server proxyServer, addressNum int) bool {
	for i := 0; i < 300; i++ {
		if len(server.Addresses()) == addressNum {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// HTTPClient is used to get target and compare result.
func HTTPClient(t *testing.T, transport *http.Transport, hostname string) {
	InitHTTPServers(t)

	// add certificate
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = new(tls.Config)
	}
	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}
	transport.TLSClientConfig.RootCAs.AddCert(httpsServerCA)
	tlsCerts := transport.TLSClientConfig.Certificates
	transport.TLSClientConfig.Certificates = append(tlsCerts, httpsClientCert)

	// make http client
	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	defer client.CloseIdleConnections()

	wg := sync.WaitGroup{}
	do := func(req *http.Request) {
		defer wg.Done()

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, testHandlerData, data)
	}

	const format = "://%s:%s/t"

	t.Run("http target", func(t *testing.T) {
		url := fmt.Sprintf("http"+format, hostname, HTTPServerPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		wg.Add(1)
		go do(req)
	})

	t.Run("https target", func(t *testing.T) {
		url := fmt.Sprintf("https"+format, hostname, HTTPSServerPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		wg.Add(1)
		go do(req)
	})

	wg.Wait()
}

// NopCloser is a nop closer.
type NopCloser struct {
	padding string
}

// Get is a padding method.
func (nc *NopCloser) Get() string {
	return nc.padding
}

// Close is a padding method.
func (nc *NopCloser) Close() error {
	nc.padding = "padding"
	return nil
}

// NewNopCloser is used to create a nop closer for ProxyServer.
func NewNopCloser() *NopCloser {
	return new(NopCloser)
}

// ProxyServer is used to test proxy server.
func ProxyServer(t *testing.T, server io.Closer, transport *http.Transport) {
	if IPv4Enabled {
		t.Run("IPv4 Only", func(t *testing.T) {
			HTTPClient(t, transport, "127.0.0.1")
		})
	}
	if IPv6Enabled {
		t.Run("IPv6 Only", func(t *testing.T) {
			HTTPClient(t, transport, "[::1]")
		})
	}
	t.Run("double stack", func(t *testing.T) {
		HTTPClient(t, transport, "localhost")
	})
	err := server.Close()
	require.NoError(t, err)
	err = server.Close()
	require.NoError(t, err)

	IsDestroyed(t, server)
}

// SendHTTPRequest is used to send a GET request to a connection.
func SendHTTPRequest(t testing.TB, conn net.Conn) {
	_ = conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	// send http request
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprint(buf, "GET /t HTTP/1.1\r\n")
	_, _ = fmt.Fprint(buf, "Host: localhost\r\n\r\n")
	_, err := buf.WriteTo(conn)
	require.NoError(t, err)
}

// ProxyConn is used to check proxy client Dial.
func ProxyConn(t testing.TB, conn net.Conn) {
	defer func() { _ = conn.Close() }()

	SendHTTPRequest(t, conn)

	// get response
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	buf := bytes.Buffer{}
	buffer := make([]byte, 1)
	for {
		n, err := conn.Read(buffer)
		require.NoError(t, err)
		buf.Write(buffer[:n])
		if buf.Len() > 4 {
			if bytes.Equal(buf.Bytes()[buf.Len()-4:], []byte("\r\n\r\n")) {
				break
			}
		}
	}

	// read body
	body := make([]byte, len(testHandlerData))
	_, err := io.ReadFull(conn, body)
	require.NoError(t, err)
	require.Equal(t, testHandlerData, body)
}

// ProxyClient is used to test proxy client.
func ProxyClient(t *testing.T, server io.Closer, client proxyClient) {
	InitHTTPServers(t)

	if IPv4Enabled {
		t.Run("IPv4 Only", func(t *testing.T) {
			proxyClientIPv4Only(t, client)
		})
	}

	// except mode socks4a, socks4
	if IPv6Enabled && strings.Contains(client.Info(), "socks5") {
		t.Run("IPv6 Only", func(t *testing.T) {
			proxyClientIPv6Only(t, client)
		})
	}

	// except mode socks4
	if !strings.Contains(client.Info(), "socks4, ") {
		t.Run("double stack", func(t *testing.T) {
			transport := new(http.Transport)
			client.HTTP(transport)
			HTTPClient(t, transport, "localhost")
			client.HTTP(transport)
			HTTPClient(t, transport, "localhost")
		})
	}

	t.Log("timeout:", client.Timeout())
	network, address := client.Server()
	t.Log("server:", network, address)
	t.Log("info:", client.Info())

	IsDestroyed(t, client)

	err := server.Close()
	require.NoError(t, err)
	IsDestroyed(t, server)
}

func proxyClientIPv4Only(t *testing.T, client proxyClient) {
	const network = "tcp4"
	address := "127.0.0.1:" + HTTPServerPort

	wg := sync.WaitGroup{}

	t.Run("dial", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.Dial(network, address)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("dial context", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.DialContext(context.Background(), network, address)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("dial timeout", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.DialTimeout(network, address, 0)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	// except mode socks4
	if !strings.Contains(client.Info(), "socks4, ") {
		t.Run("dial with host name", func(t *testing.T) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				address := "localhost:" + HTTPServerPort
				conn, err := client.DialTimeout(network, address, 0)
				require.NoError(t, err)
				ProxyConn(t, conn)
			}()
		})
	}

	wg.Wait()

	t.Run("set transport twice", func(t *testing.T) {
		transport := new(http.Transport)
		client.HTTP(transport)
		HTTPClient(t, transport, "127.0.0.1")
		client.HTTP(transport)
		HTTPClient(t, transport, "127.0.0.1")
	})
}

func proxyClientIPv6Only(t *testing.T, client proxyClient) {
	const network = "tcp6"
	address := "[::1]:" + HTTPServerPort

	wg := sync.WaitGroup{}

	t.Run("dial", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.Dial(network, address)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("dial context", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.DialContext(context.Background(), network, address)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("dial timeout", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := client.DialTimeout(network, address, 0)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("dial with host name", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			address := "localhost:" + HTTPServerPort
			conn, err := client.DialTimeout(network, address, 0)
			require.NoError(t, err)
			ProxyConn(t, conn)
		}()
	})

	t.Run("set transport twice", func(t *testing.T) {
		transport := new(http.Transport)
		client.HTTP(transport)
		HTTPClient(t, transport, "[::1]")
		client.HTTP(transport)
		HTTPClient(t, transport, "[::1]")
	})
}

// ProxyClientCancelConnect is used to cancel proxy client Connect().
func ProxyClientCancelConnect(t testing.TB, server io.Closer, client proxyClient) {
	InitHTTPServers(t)

	conn, err := net.Dial(client.Server())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cancel
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		cancel()
	}()
	time.Sleep(10 * time.Millisecond)

	address := "127.0.0.1:" + HTTPServerPort
	_, err = client.Connect(ctx, conn, "tcp", address)
	require.Error(t, err)
	wg.Wait()

	IsDestroyed(t, client)

	err = server.Close()
	require.NoError(t, err)
	IsDestroyed(t, server)
}

// ProxyClientWithHTTPSTarget is used to test proxy client with https target.
func ProxyClientWithHTTPSTarget(t testing.TB, client proxyClient) {
	transport := new(http.Transport)
	certPool, err := certpool.System()
	require.NoError(t, err)

	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    certPool,
	}
	client.HTTP(transport)
	httpClient := http.Client{Transport: transport}
	defer httpClient.CloseIdleConnections()

	resp, err := httpClient.Get("https://www.cloudflare.com/")
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
}

// ProxyClientWithUnreachableProxyServer is used to test proxy client can't connect proxy server.
func ProxyClientWithUnreachableProxyServer(t *testing.T, client proxyClient) {
	ctx := context.Background()

	t.Run("unknown network", func(t *testing.T) {
		const (
			network = "foo"
			address = "127.0.0.1:1234"
		)

		t.Run("Dial", func(t *testing.T) {
			_, err := client.Dial(network, address)
			require.Error(t, err)
			t.Log("Dial:\n", err)
		})

		t.Run("DialContext", func(t *testing.T) {
			_, err := client.DialContext(ctx, network, address)
			require.Error(t, err)
			t.Log("DialContext:\n", err)
		})

		t.Run("DialTimeout", func(t *testing.T) {
			_, err := client.DialTimeout(network, address, time.Second)
			require.Error(t, err)
			t.Log("DialTimeout:\n", err)
		})
	})

	t.Run("unreachable proxy server", func(t *testing.T) {
		const (
			network = "tcp"
			address = "127.0.0.1:1234"
		)

		t.Run("Dial", func(t *testing.T) {
			_, err := client.Dial(network, address)
			require.Error(t, err)
			t.Log("Dial:\n", err)
		})

		t.Run("DialContext", func(t *testing.T) {
			_, err := client.DialContext(ctx, network, address)
			require.Error(t, err)
			t.Log("DialContext:\n", err)
		})

		t.Run("DialTimeout", func(t *testing.T) {
			_, err := client.DialTimeout(network, address, time.Second)
			require.Error(t, err)
			t.Log("DialTimeout:\n", err)
		})
	})

	IsDestroyed(t, client)
}

// ProxyClientWithUnreachableTarget is used to test proxy client connect unreachable target.
func ProxyClientWithUnreachableTarget(t *testing.T, server io.Closer, client proxyClient) {
	const (
		network = "tcp"
		address = "0.0.0.0:1"
	)

	t.Run("Dial", func(t *testing.T) {
		_, err := client.Dial(network, address)
		require.Error(t, err)
		t.Log("Dial -> Connect:\n", err)
	})

	t.Run("DialContext", func(t *testing.T) {
		_, err := client.DialContext(context.Background(), network, address)
		require.Error(t, err)
		t.Log("DialContext -> Connect:\n", err)
	})

	t.Run("DialTimeout", func(t *testing.T) {
		_, err := client.DialTimeout(network, address, time.Second)
		require.Error(t, err)
		t.Log("DialTimeout -> Connect:\n", err)
	})

	IsDestroyed(t, client)

	err := server.Close()
	require.NoError(t, err)
	IsDestroyed(t, server)
}
