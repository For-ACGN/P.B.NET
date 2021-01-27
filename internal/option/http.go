package option

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const defaultHTTPMultiTimeout = time.Minute

// HTTPRequest contains options about http.Request.
type HTTPRequest struct {
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	Method string `toml:"method"`

	// URL specifies either the URI being requested.
	URL string `toml:"url"`

	// Header contains the http request header.
	Header http.Header `toml:"header"`

	// Host optionally overrides the Host header to send.
	Host string `toml:"host"`

	// Close is used to prevent re-use of connections between
	// requests to the same hosts.
	Close bool `toml:"close"`

	// Post is the http request body that encoded by hex.
	Post string `toml:"post"`

	// Body is the http request body, if it is set, Apply will
	// use it to replace Post field.
	Body io.Reader `toml:"-" msgpack:"-"`
}

// Apply is used to create *http.Request.
func (hr *HTTPRequest) Apply() (*http.Request, error) {
	if hr.URL == "" {
		return nil, hr.error("empty url")
	}
	body := hr.Body
	if hr.Body == nil {
		post, err := hex.DecodeString(hr.Post)
		if err != nil {
			return nil, hr.error(err)
		}
		body = bytes.NewReader(post)
	}
	req, err := http.NewRequest(hr.Method, hr.URL, body)
	if err != nil {
		return nil, hr.error(err)
	}
	if hr.Header != nil {
		req.Header = hr.Header.Clone()
	}
	req.Host = hr.Host
	req.Close = hr.Close
	return req, nil
}

func (hr *HTTPRequest) error(err interface{}) error {
	return fmt.Errorf("failed to apply http request option: %s", err)
}

// HTTPTransport contains options about http.Transport.
type HTTPTransport struct {
	// TLSClientConfig contain TLS configuration.
	TLSClientConfig TLSConfig `toml:"tls_config" testsuite:"-"`

	// TLSHandshakeTimeout specifies the maximum amount of time waiting
	// to wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration `toml:"tls_handshake_timeout"`

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `toml:"max_idle_conns"`

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `toml:"max_idle_conns_per_host"`

	// MaxConnsPerHost optionally limits the total number of connections
	// per host, including connections in the dialing, active, and idle
	// states. On limit violation, dials will block. Zero means no limit.
	MaxConnsPerHost int `toml:"max_conns_per_host"`

	// IdleConnTimeout is the maximum amount of time an idle (keep-alive)
	// connection will remain idle before closing itself. Zero means no limit.
	IdleConnTimeout time.Duration `toml:"idle_conn_timeout"`

	// ResponseHeaderTimeout, if non-zero, specifies the amount of time to
	// wait for a server's response headers after fully writing the request
	// (including its body, if any). This time does not include the time to
	// read the response body.
	ResponseHeaderTimeout time.Duration `toml:"response_header_timeout"`

	// ExpectContinueTimeout, if non-zero, specifies the amount of time to
	// wait for a server's first response headers after fully writing the
	// request headers if the request has an "Expect: 100-continue" header.
	// Zero means no timeout and causes the body to be sent immediately,
	// without waiting for the server to approve. This time does not include
	// the time to send the request header.
	ExpectContinueTimeout time.Duration `toml:"expect_continue_timeout"`

	// MaxResponseHeaderBytes specifies a limit on how many response bytes
	// are allowed in the server's response header. Zero means to use a
	// default limit.
	MaxResponseHeaderBytes int64 `toml:"max_response_header_bytes"`

	// DisableKeepAlives, if true, disables HTTP keep-alives and will only
	// use the connection to the server for a single HTTP request. This is
	// unrelated to the similarly named TCP keep-alives.
	DisableKeepAlives bool `toml:"disable_keep_alives"`

	// DisableCompression, if true, prevents the Transport from requesting
	// compression with an "Accept-Encoding: gzip" request header when the
	// Request contains no existing Accept-Encoding value. If the Transport
	// requests gzip on its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user explicitly requested
	// gzip it is not automatically uncompressed.
	DisableCompression bool `toml:"disable_compression"`

	// ProxyConnectHeader optionally specifies headers to send to proxies
	// during CONNECT requests.
	ProxyConnectHeader http.Header `toml:"proxy_connect_header"`

	// Proxy specifies a function to return a proxy for a given Request.
	// If the function returns a non-nil error, the request is aborted
	// with the provided error.
	//
	// The proxy type is determined by the URL scheme. "http", "https", and
	// "socks5" are supported. If the scheme is empty, "http" is assumed.
	//
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	Proxy func(*http.Request) (*url.URL, error) `toml:"-" msgpack:"-"`

	// DialContext specifies the dial function for creating unencrypted TCP.
	// connections. If DialContext is nil (and the deprecated Dial below is
	// also nil), then the transport dials using package net.
	//
	// DialContext runs concurrently with calls to RoundTrip. A RoundTrip
	// call that initiates a dial may end up using a connection dialed
	// previously when the earlier connection becomes idle before the later
	// DialContext completes.
	DialContext func(context.Context, string, string) (net.Conn, error) `toml:"-" msgpack:"-"`
}

// Apply is used to create *http.Transport.
func (ht *HTTPTransport) Apply() (*http.Transport, error) {
	tr := http.Transport{
		TLSHandshakeTimeout:    ht.TLSHandshakeTimeout,
		MaxIdleConns:           ht.MaxIdleConns,
		MaxIdleConnsPerHost:    ht.MaxIdleConnsPerHost,
		MaxConnsPerHost:        ht.MaxConnsPerHost,
		IdleConnTimeout:        ht.IdleConnTimeout,
		ResponseHeaderTimeout:  ht.ResponseHeaderTimeout,
		ExpectContinueTimeout:  ht.ExpectContinueTimeout,
		MaxResponseHeaderBytes: ht.MaxResponseHeaderBytes,
		DisableKeepAlives:      ht.DisableKeepAlives,
		DisableCompression:     ht.DisableCompression,
		ProxyConnectHeader:     ht.ProxyConnectHeader.Clone(),
		Proxy:                  ht.Proxy,
		DialContext:            ht.DialContext,
	}
	// about TLS configuration
	var err error
	tr.TLSClientConfig, err = ht.TLSClientConfig.Apply()
	if err != nil {
		return nil, err
	}
	if tr.TLSHandshakeTimeout < 1 {
		tr.TLSHandshakeTimeout = defaultHTTPMultiTimeout
	}
	// about maximum connection
	if tr.MaxIdleConns < 1 {
		tr.MaxIdleConns = 4
	}
	if tr.MaxIdleConnsPerHost < 1 {
		tr.MaxIdleConnsPerHost = 1
	}
	if tr.MaxConnsPerHost < 0 {
		tr.MaxConnsPerHost = 16
	}
	// about timeout
	if tr.IdleConnTimeout < 1 {
		tr.IdleConnTimeout = defaultHTTPMultiTimeout
	}
	if tr.ResponseHeaderTimeout < 1 {
		tr.ResponseHeaderTimeout = defaultHTTPMultiTimeout
	}
	if tr.ExpectContinueTimeout < 1 {
		tr.ExpectContinueTimeout = defaultHTTPMultiTimeout
	}
	// max header bytes
	if tr.MaxResponseHeaderBytes < 1 {
		tr.MaxResponseHeaderBytes = 1024 * 1024
	}
	return &tr, nil
}

// HTTPServer contains options about http.Server.
type HTTPServer struct {
	// TLSConfig contain TLS configuration.
	TLSConfig TLSConfig `toml:"tls_config" testsuite:"-"`

	// ReadTimeout is the maximum duration for reading the entire request,
	// including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request decisions
	// on each request body's acceptable deadline or upload rate, most users
	// will prefer to use ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `toml:"read_timeout"`

	// WriteTimeout is the maximum duration before timing out writes of the
	// response. It is reset whenever a new request's header is read. it does
	// not let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration `toml:"write_timeout"`

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and
	// the Handler can decide what is considered too slow for the body. If
	// ReadHeaderTimeout is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration `toml:"read_header_timeout"`

	// IdleTimeout is the maximum amount of time to wait for the next request
	// when keep-alives are enabled. If IdleTimeout is zero, the value of
	// ReadTimeout is used. If both are zero, there is no timeout.
	IdleTimeout time.Duration `toml:"idle_timeout"`

	// MaxHeaderBytes controls the maximum number of bytes the server will read
	// parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body.
	//
	// If zero, http.DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `toml:"max_header_bytes"`

	// DisableKeepAlive controls whether HTTP keep-alives are enabled. By default,
	// keep-alives are always enabled. Only very resource-constrained environments
	// or servers in the process of shutting down should disable them.
	DisableKeepAlive bool `toml:"disable_keep_alive"`
}

// Apply is used to create *http.Server.
func (hs *HTTPServer) Apply() (*http.Server, error) {
	srv := http.Server{
		ReadTimeout:       hs.ReadTimeout,
		WriteTimeout:      hs.WriteTimeout,
		ReadHeaderTimeout: hs.ReadHeaderTimeout,
		IdleTimeout:       hs.IdleTimeout,
		MaxHeaderBytes:    hs.MaxHeaderBytes,
	}
	// force set it to server side
	tlsConfig := hs.TLSConfig
	tlsConfig.ServerSide = true
	// about TLS configuration
	var err error
	srv.TLSConfig, err = tlsConfig.Apply()
	if err != nil {
		return nil, err
	}
	// about timeout
	if srv.ReadTimeout < 0 {
		srv.ReadTimeout = defaultHTTPMultiTimeout
	}
	if srv.WriteTimeout < 0 {
		srv.WriteTimeout = defaultHTTPMultiTimeout
	}
	if srv.ReadHeaderTimeout < 1 {
		srv.ReadHeaderTimeout = defaultHTTPMultiTimeout
	}
	if srv.IdleTimeout < 1 {
		srv.IdleTimeout = defaultHTTPMultiTimeout
	}
	// max header bytes
	if srv.MaxHeaderBytes < 1 {
		srv.MaxHeaderBytes = 1024 * 1024
	}
	srv.SetKeepAlivesEnabled(!hs.DisableKeepAlive)
	return &srv, nil
}
