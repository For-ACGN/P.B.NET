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
	tr, err := new(HTTPTransport).Apply()
	require.NoError(t, err)

	require.Equal(t, defaultHTTPMultiTimeout, tr.TLSHandshakeTimeout)
	require.Equal(t, 4, tr.MaxIdleConns)
	require.Equal(t, 1, tr.MaxIdleConnsPerHost)
	require.Equal(t, 0, tr.MaxConnsPerHost)
	require.Equal(t, defaultHTTPMultiTimeout, tr.IdleConnTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, tr.ResponseHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, tr.ExpectContinueTimeout)
	require.Equal(t, int64(1024*1024), tr.MaxResponseHeaderBytes)
	require.Equal(t, false, tr.DisableKeepAlives)
	require.Equal(t, false, tr.DisableCompression)
	require.Empty(t, tr.ProxyConnectHeader)
	require.Nil(t, tr.Proxy)
	require.Nil(t, tr.DialContext)
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

	tr, err := ht.Apply()
	require.NoError(t, err)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: "test.com", actual: tr.TLSClientConfig.ServerName},
		{expected: 20 * time.Second, actual: tr.TLSHandshakeTimeout},
		{expected: 2, actual: tr.MaxIdleConns},
		{expected: 4, actual: tr.MaxIdleConnsPerHost},
		{expected: 8, actual: tr.MaxConnsPerHost},
		{expected: 10 * time.Second, actual: tr.IdleConnTimeout},
		{expected: 12 * time.Second, actual: tr.ResponseHeaderTimeout},
		{expected: 14 * time.Second, actual: tr.ExpectContinueTimeout},
		{expected: int64(16384), actual: tr.MaxResponseHeaderBytes},
		{expected: true, actual: tr.DisableKeepAlives},
		{expected: true, actual: tr.DisableCompression},
		{expected: []string{"testdata"}, actual: tr.ProxyConnectHeader["Test"]},
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
	srv, err := new(HTTPServer).Apply()
	require.NoError(t, err)

	require.Equal(t, time.Duration(0), srv.ReadTimeout)
	require.Equal(t, time.Duration(0), srv.WriteTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, srv.ReadHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, srv.IdleTimeout)
	require.Equal(t, 1024*1024, srv.MaxHeaderBytes)
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

	srv, err := hs.Apply()
	require.NoError(t, err)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: 10 * time.Second, actual: srv.ReadTimeout},
		{expected: 12 * time.Second, actual: srv.WriteTimeout},
		{expected: 14 * time.Second, actual: srv.ReadHeaderTimeout},
		{expected: 16 * time.Second, actual: srv.IdleTimeout},
		{expected: 16384, actual: srv.MaxHeaderBytes},
		{expected: "test.com", actual: srv.TLSConfig.ServerName},
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
		srv, err := hs.Apply()
		require.NoError(t, err)

		require.Equal(t, defaultHTTPMultiTimeout, srv.ReadTimeout)
		require.Equal(t, defaultHTTPMultiTimeout, srv.WriteTimeout)
	})
}
