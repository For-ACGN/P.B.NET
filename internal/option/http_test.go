package option

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
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

func TestHTTPRequest_Body(t *testing.T) {
	hr := HTTPRequest{URL: "http://127.0.0.1/"}
	hr.Body = strings.NewReader("test")

	req, err := hr.Apply()
	require.NoError(t, err)

	post, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, "test", string(post))
}

func TestHTTPRequest_Apply(t *testing.T) {
	// empty url
	req := HTTPRequest{}
	_, err := req.Apply()
	require.Errorf(t, err, "failed to apply http request options: empty url")

	// invalid post data
	req = HTTPRequest{
		URL:  "http://localhost/",
		Post: "foo post data",
	}
	_, err = req.Apply()
	require.Error(t, err)

	// invalid method
	req.Post = "0102"
	req.Method = "invalid method"
	_, err = req.Apply()
	require.Error(t, err)
}

func TestHTTPTransportDefault(t *testing.T) {
	transport, err := new(HTTPTransport).Apply()
	require.NoError(t, err)

	require.Equal(t, 1, transport.MaxIdleConns)
	require.Equal(t, 1, transport.MaxIdleConnsPerHost)
	// require.Equal(t, 1, transport.MaxConnsPerHost)
	require.Equal(t, defaultHTTPMultiTimeout, transport.TLSHandshakeTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, transport.IdleConnTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, transport.ResponseHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, transport.ExpectContinueTimeout)
	require.Equal(t, defaultHTTPMaxResponseHeaderBytes, transport.MaxResponseHeaderBytes)
	require.Equal(t, false, transport.DisableKeepAlives)
	require.Equal(t, false, transport.DisableCompression)
	require.Empty(t, transport.ProxyConnectHeader)
	require.Nil(t, transport.Proxy)
	require.Nil(t, transport.DialContext)
}

func TestHTTPTransport(t *testing.T) {
	data, err := os.ReadFile("testdata/http_transport.toml")
	require.NoError(t, err)

	proxy := func(*http.Request) (*url.URL, error) {
		return nil, nil
	}
	dialContext := func(context.Context, string, string) (net.Conn, error) {
		return nil, nil
	}
	tr := HTTPTransport{
		Proxy:       proxy,
		DialContext: dialContext,
	}

	// check unnecessary field
	err = toml.Unmarshal(data, &tr)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, tr)

	transport, err := tr.Apply()
	require.NoError(t, err)
	const timeout = 10 * time.Second

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: 2, actual: transport.MaxIdleConns},
		{expected: 2, actual: transport.MaxIdleConnsPerHost},
		// {expected: 2, actual: transport.MaxConnsPerHost},
		{expected: timeout, actual: transport.TLSHandshakeTimeout},
		{expected: timeout, actual: transport.IdleConnTimeout},
		{expected: timeout, actual: transport.ResponseHeaderTimeout},
		{expected: timeout, actual: transport.ExpectContinueTimeout},
		{expected: int64(16384), actual: transport.MaxResponseHeaderBytes},
		{expected: true, actual: transport.DisableKeepAlives},
		{expected: true, actual: transport.DisableCompression},
		{expected: "test.com", actual: transport.TLSClientConfig.ServerName},
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
	tr := HTTPTransport{
		TLSClientConfig: testInvalidTLSConfig,
	}
	_, err := tr.Apply()
	require.Error(t, err)
}

func TestHTTPServerDefault(t *testing.T) {
	server, err := new(HTTPServer).Apply()
	require.NoError(t, err)

	require.Equal(t, time.Duration(0), server.ReadTimeout)
	require.Equal(t, time.Duration(0), server.WriteTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, server.ReadHeaderTimeout)
	require.Equal(t, defaultHTTPMultiTimeout, server.IdleTimeout)
	require.Equal(t, defaultHTTPMaxHeaderBytes, server.MaxHeaderBytes)
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
	const timeout = 10 * time.Second

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: timeout, actual: server.ReadTimeout},
		{expected: timeout, actual: server.WriteTimeout},
		{expected: timeout, actual: server.ReadHeaderTimeout},
		{expected: timeout, actual: server.IdleTimeout},
		{expected: 16384, actual: server.MaxHeaderBytes},
		{expected: "test.com", actual: server.TLSConfig.ServerName},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}

func TestHTTPServer_Apply(t *testing.T) {
	s := HTTPServer{
		TLSConfig: testInvalidTLSConfig,
	}
	_, err := s.Apply()
	require.Error(t, err)
}

func TestHTTPServerSetTimeout(t *testing.T) {
	s := HTTPServer{
		ReadTimeout:  -1,
		WriteTimeout: -1,
	}
	_, err := s.Apply()
	require.NoError(t, err)
}
