package option

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// http transport and server default timeout
	httpDefaultTimeout = time.Minute

	// http server income request max header size
	httpDefaultMaxHeaderBytes = 512 * 1024

	// http transport max response header size
	httpDefaultMaxResponseHeaderBytes int64 = 512 * 1024
)

// HTTPRequest contains options about http.Request.
type HTTPRequest struct {
	Method string      `toml:"method"`
	URL    string      `toml:"url"`
	Post   string      `toml:"post"` // hex
	Header http.Header `toml:"header"`
	Host   string      `toml:"host"`
	Close  bool        `toml:"close"`
}

func (hr *HTTPRequest) error(err error) error {
	return fmt.Errorf("failed to apply http request options: %s", err)
}

// Apply is used to create *http.Request.
func (hr *HTTPRequest) Apply() (*http.Request, error) {
	if hr.URL == "" {
		return nil, hr.error(errors.New("empty url"))
	}
	post, err := hex.DecodeString(hr.Post)
	if err != nil {
		return nil, hr.error(err)
	}
	req, err := http.NewRequest(hr.Method, hr.URL, bytes.NewReader(post))
	if err != nil {
		return nil, hr.error(err)
	}
	req.Header = hr.Header.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Host = hr.Host
	req.Close = hr.Close
	return req, nil
}

// HTTPTransport contains options about http.Transport.
type HTTPTransport struct {
	TLSClientConfig        TLSConfig     `toml:"tls_config" testsuite:"-"`
	MaxIdleConns           int           `toml:"max_idle_conns"`
	MaxIdleConnsPerHost    int           `toml:"max_idle_conns_per_host"`
	MaxConnsPerHost        int           `toml:"max_conns_per_host" testsuite:"-"`
	TLSHandshakeTimeout    time.Duration `toml:"tls_handshake_timeout"`
	IdleConnTimeout        time.Duration `toml:"idle_conn_timeout"`
	ResponseHeaderTimeout  time.Duration `toml:"response_header_timeout"`
	ExpectContinueTimeout  time.Duration `toml:"expect_continue_timeout"`
	MaxResponseHeaderBytes int64         `toml:"max_response_header_bytes"`
	DisableKeepAlives      bool          `toml:"disable_keep_alives"`
	DisableCompression     bool          `toml:"disable_compression"`
	ProxyConnectHeader     http.Header   `toml:"proxy_connect_header"`
}

// Apply is used to create *http.Transport.
//
// TODO [external] go internal bug: MaxConnsPerHost
// when set MaxConnsPerHost, use HTTP/2 get test.com will panic, wait golang fix it.
func (ht *HTTPTransport) Apply() (*http.Transport, error) {
	tr := http.Transport{
		MaxIdleConns:        ht.MaxIdleConns,
		MaxIdleConnsPerHost: ht.MaxIdleConnsPerHost,
		// MaxConnsPerHost:        ht.MaxConnsPerHost,
		TLSHandshakeTimeout:    ht.TLSHandshakeTimeout,
		IdleConnTimeout:        ht.IdleConnTimeout,
		ResponseHeaderTimeout:  ht.ResponseHeaderTimeout,
		ExpectContinueTimeout:  ht.ExpectContinueTimeout,
		MaxResponseHeaderBytes: ht.MaxResponseHeaderBytes,
		DisableKeepAlives:      ht.DisableKeepAlives,
		DisableCompression:     ht.DisableCompression,
		ProxyConnectHeader:     ht.ProxyConnectHeader.Clone(),
	}
	// about tls configuration
	var err error
	tr.TLSClientConfig, err = ht.TLSClientConfig.Apply()
	if err != nil {
		return nil, err
	}
	// about connection
	if tr.MaxIdleConns < 1 {
		tr.MaxIdleConns = 1
	}
	if tr.MaxIdleConnsPerHost < 1 {
		tr.MaxIdleConnsPerHost = 1
	}
	// if tr.MaxConnsPerHost < 1 {
	// 	tr.MaxConnsPerHost = 1
	// }
	// about timeout
	if tr.TLSHandshakeTimeout < 1 {
		tr.TLSHandshakeTimeout = httpDefaultTimeout
	}
	if tr.IdleConnTimeout < 1 {
		tr.IdleConnTimeout = httpDefaultTimeout
	}
	if tr.ResponseHeaderTimeout < 1 {
		tr.ResponseHeaderTimeout = httpDefaultTimeout
	}
	if tr.ExpectContinueTimeout < 1 {
		tr.ExpectContinueTimeout = httpDefaultTimeout
	}
	// max header bytes
	if tr.MaxResponseHeaderBytes < 1 {
		tr.MaxResponseHeaderBytes = httpDefaultMaxResponseHeaderBytes
	}
	return &tr, nil
}

// HTTPServer contains options about http.Server.
type HTTPServer struct {
	TLSConfig         TLSConfig     `toml:"tls_config" testsuite:"-"`
	ReadTimeout       time.Duration `toml:"read_timeout"`  // warning
	WriteTimeout      time.Duration `toml:"write_timeout"` // warning
	ReadHeaderTimeout time.Duration `toml:"read_header_timeout"`
	IdleTimeout       time.Duration `toml:"idle_timeout"`
	MaxHeaderBytes    int           `toml:"max_header_bytes"`
	DisableKeepAlive  bool          `toml:"disable_keep_alive"`
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
	hs.TLSConfig.ServerSide = true
	// tls config
	var err error
	srv.TLSConfig, err = hs.TLSConfig.Apply()
	if err != nil {
		return nil, err
	}
	// timeout
	if srv.ReadTimeout < 0 {
		srv.ReadTimeout = httpDefaultTimeout
	}
	if srv.WriteTimeout < 0 {
		srv.WriteTimeout = httpDefaultTimeout
	}
	if srv.ReadHeaderTimeout < 1 {
		srv.ReadHeaderTimeout = httpDefaultTimeout
	}
	if srv.IdleTimeout < 1 {
		srv.IdleTimeout = httpDefaultTimeout
	}
	// max header bytes
	if srv.MaxHeaderBytes < 1 {
		srv.MaxHeaderBytes = httpDefaultMaxHeaderBytes
	}
	srv.SetKeepAlivesEnabled(!hs.DisableKeepAlive)
	return &srv, nil
}
