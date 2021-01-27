package option

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
)

func TestHTTPRequestDefault(t *testing.T) {
	const URL = "http://127.0.0.1/"

	hr := HTTPRequest{URL: URL}
	req, err := hr.Apply()
	require.NoError(t, err)

	require.Equal(t, http.MethodGet, req.Method)
	require.Equal(t, URL, req.URL.String())
	require.NotNil(t, req.Header)
	require.Zero(t, req.Host)
	require.Equal(t, false, req.Close)
	require.Equal(t, http.NoBody, req.Body)
}

func TestHTTPRequest(t *testing.T) {
	data, err := os.ReadFile("testdata/http_request.toml")
	require.NoError(t, err)

	// check unnecessary field
	hr := HTTPRequest{}
	err = toml.Unmarshal(data, &hr)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, hr)

	req, err := hr.Apply()
	require.NoError(t, err)
	post, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: http.MethodPost, actual: req.Method},
		{expected: "https://127.0.0.1/", actual: req.URL.String()},
		{expected: "keep-alive", actual: req.Header.Get("Connection")},
		{expected: 7, actual: len(req.Header)},
		{expected: "localhost", actual: req.Host},
		{expected: true, actual: req.Close},
		{expected: []byte{1, 2}, actual: post},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}

func TestHTTPRequest_Apply(t *testing.T) {
	const URL = "http://127.0.0.1/"

	t.Run("empty url", func(t *testing.T) {
		hr := HTTPRequest{}
		_, err := hr.Apply()
		require.EqualError(t, err, "failed to apply http request option: empty url")
	})

	t.Run("with body", func(t *testing.T) {
		hr := HTTPRequest{URL: URL}
		hr.Body = strings.NewReader("test")

		req, err := hr.Apply()
		require.NoError(t, err)

		post, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, "test", string(post))
	})

	t.Run("invalid post data", func(t *testing.T) {
		hr := HTTPRequest{
			URL:  URL,
			Post: "foo post data",
		}
		_, err := hr.Apply()
		require.Error(t, err)
	})

	t.Run("invalid method", func(t *testing.T) {
		hr := HTTPRequest{
			Method: "invalid method",
			URL:    URL,
			Post:   "0102",
		}
		_, err := hr.Apply()
		require.Error(t, err)
	})
}

func TestHTTPTransportDefault(t *testing.T) {
	ht, err := new(HTTPTransport).Apply()
	require.NoError(t, err)

	require.Equal(t, defaultHTTPMultiTimeout, ht.TLSHandshakeTimeout)
	require.Equal(t, 4, ht.MaxIdleConns)
	require.Equal(t, 1, ht.MaxIdleConnsPerHost)
	require.Equal(t, 0, ht.MaxConnsPerHost)
	require.Equal(t, defaultHTTPMultiTimeout, ht.IdleConnTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, ht.ResponseHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, ht.ExpectContinueTimeout)
	require.Equal(t, int64(1024*1024), ht.MaxResponseHeaderBytes)
	require.Equal(t, false, ht.DisableKeepAlives)
	require.Equal(t, false, ht.DisableCompression)
	require.Empty(t, ht.ProxyConnectHeader)
	require.Nil(t, ht.Proxy)
	require.Nil(t, ht.DialContext)
}

func TestHTTPTransport(t *testing.T) {
	data, err := os.ReadFile("testdata/http_transport.toml")
	require.NoError(t, err)

	// check unnecessary field
	ht := HTTPTransport{}
	err = toml.Unmarshal(data, &ht)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, ht)

	transport, err := ht.Apply()
	require.NoError(t, err)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: "test.com", actual: transport.TLSClientConfig.ServerName},
		{expected: 20 * time.Second, actual: transport.TLSHandshakeTimeout},
		{expected: 2, actual: transport.MaxIdleConns},
		{expected: 4, actual: transport.MaxIdleConnsPerHost},
		{expected: 8, actual: transport.MaxConnsPerHost},
		{expected: 10 * time.Second, actual: transport.IdleConnTimeout},
		{expected: 12 * time.Second, actual: transport.ResponseHeaderTimeout},
		{expected: 14 * time.Second, actual: transport.ExpectContinueTimeout},
		{expected: int64(16384), actual: transport.MaxResponseHeaderBytes},
		{expected: true, actual: transport.DisableKeepAlives},
		{expected: true, actual: transport.DisableCompression},
		{expected: []string{"testdata"}, actual: transport.ProxyConnectHeader["Test"]},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}

var testInvalidTLSConfig = TLSConfig{
	RootCAs:   []string{"foo data"},
	ClientCAs: []string{"foo data"},
}

func TestHTTPTransport_Apply(t *testing.T) {
	t.Run("invalid tls config", func(t *testing.T) {
		ht := HTTPTransport{
			TLSClientConfig: testInvalidTLSConfig,
		}
		_, err := ht.Apply()
		require.Error(t, err)
	})

	t.Run("invalid MaxConnsPerHost", func(t *testing.T) {
		ht := HTTPTransport{
			MaxConnsPerHost: -1,
		}
		tr, err := ht.Apply()
		require.NoError(t, err)

		require.Equal(t, 16, tr.MaxConnsPerHost)
	})
}

func TestHTTPServerDefault(t *testing.T) {
	server, err := new(HTTPServer).Apply()
	require.NoError(t, err)

	require.Equal(t, time.Duration(0), server.ReadTimeout)
	require.Equal(t, time.Duration(0), server.WriteTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, server.ReadHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, server.IdleTimeout)
	require.Equal(t, 1024*1024, server.MaxHeaderBytes)
}

func TestHTTPServer(t *testing.T) {
	data, err := os.ReadFile("testdata/http_server.toml")
	require.NoError(t, err)

	// check unnecessary field
	hs := HTTPServer{}
	err = toml.Unmarshal(data, &hs)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, hs)

	server, err := hs.Apply()
	require.NoError(t, err)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: 10 * time.Second, actual: server.ReadTimeout},
		{expected: 12 * time.Second, actual: server.WriteTimeout},
		{expected: 14 * time.Second, actual: server.ReadHeaderTimeout},
		{expected: 16 * time.Second, actual: server.IdleTimeout},
		{expected: 16384, actual: server.MaxHeaderBytes},
		{expected: "test.com", actual: server.TLSConfig.ServerName},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}

func TestHTTPServer_Apply(t *testing.T) {
	t.Run("invalid tls config", func(t *testing.T) {
		hs := HTTPServer{
			TLSConfig: testInvalidTLSConfig,
		}
		_, err := hs.Apply()
		require.Error(t, err)
	})

	t.Run("invalid timeout", func(t *testing.T) {
		hs := HTTPServer{
			ReadTimeout:  -1,
			WriteTimeout: -1,
		}
		server, err := hs.Apply()
		require.NoError(t, err)

		require.Equal(t, defaultHTTPMultiTimeout, server.ReadTimeout)
		require.Equal(t, defaultHTTPMultiTimeout, server.WriteTimeout)
	})
}
