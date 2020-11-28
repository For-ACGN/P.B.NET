package http

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/netutil"

	"project/internal/httptool"
	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/security"
	"project/internal/xpanic"
	"project/internal/xsync"
)

// EmptyTag is a reserve tag that delete "-" in tag,
// "https proxy- " -> "https proxy", it is used to tool/proxy.
const EmptyTag = " "

// Server implemented internal/proxy.server.
type Server struct {
	logger   logger.Logger
	https    bool
	maxConns int
	logSrc   string

	// accept client request
	server  *http.Server
	handler *handler

	// listener addresses
	addresses    map[*net.Addr]struct{}
	addressesRWM sync.RWMutex
}

// NewHTTPServer is used to create a HTTP proxy server.
func NewHTTPServer(tag string, lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(tag, lg, opts, false)
}

// NewHTTPSServer is used to create a HTTPS proxy server.
func NewHTTPSServer(tag string, lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(tag, lg, opts, true)
}

func newServer(tag string, lg logger.Logger, opts *Options, https bool) (*Server, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	if opts == nil {
		opts = new(Options)
	}
	// apply options about http
	server, err := opts.Server.Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// set transport
	transport, err := opts.Transport.Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	transport.DialContext = opts.DialContext
	// log source
	var logSrc string
	if https {
		logSrc = "https proxy"
	} else {
		logSrc = "http proxy"
	}
	if tag != EmptyTag {
		logSrc += "-" + tag
	}
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = defaultConnectTimeout
	}
	handler := &handler{
		logger:      lg,
		logSrc:      logSrc,
		timeout:     timeout,
		transport:   transport,
		dialContext: opts.DialContext,
	}
	if handler.dialContext == nil {
		handler.dialContext = new(net.Dialer).DialContext
	}
	if opts.Username != "" {
		if strings.Contains(opts.Username, ":") { // can not include ":"
			return nil, errors.New("username can not include character \":\"")
		}
		handler.username = security.NewString(opts.Username)
	}
	if opts.Password != "" {
		handler.password = security.NewString(opts.Password)
	}
	handler.ctx, handler.cancel = context.WithCancel(context.Background())
	// set http server
	server.Handler = handler
	server.ReadTimeout = timeout
	server.WriteTimeout = timeout
	server.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			handler.counter.Add(1)
		case http.StateHijacked, http.StateClosed:
			handler.counter.Done()
		}
	}
	server.ErrorLog = logger.Wrap(logger.Warning, logSrc, lg)
	// set proxy server
	srv := Server{
		logger:    lg,
		https:     https,
		maxConns:  opts.MaxConns,
		logSrc:    logSrc,
		server:    server,
		handler:   handler,
		addresses: make(map[*net.Addr]struct{}, 1),
	}
	if srv.maxConns < 1 {
		srv.maxConns = defaultMaxConnections
	}
	return &srv, nil
}

func (srv *Server) logf(lv logger.Level, format string, log ...interface{}) {
	srv.logger.Printf(lv, srv.logSrc, format, log...)
}

func (srv *Server) log(lv logger.Level, log ...interface{}) {
	srv.logger.Println(lv, srv.logSrc, log...)
}

func (srv *Server) addListenerAddress(addr *net.Addr) {
	srv.addressesRWM.Lock()
	defer srv.addressesRWM.Unlock()
	srv.addresses[addr] = struct{}{}
}

func (srv *Server) deleteListenerAddress(addr *net.Addr) {
	srv.addressesRWM.Lock()
	defer srv.addressesRWM.Unlock()
	delete(srv.addresses, addr)
}

// ListenAndServe is used to listen a listener and serve.
func (srv *Server) ListenAndServe(network, address string) error {
	err := nettool.IsTCPNetwork(network)
	if err != nil {
		return errors.WithStack(err)
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return errors.WithStack(err)
	}
	return srv.Serve(listener)
}

// Serve accepts incoming connections on the listener.
func (srv *Server) Serve(listener net.Listener) (err error) {
	srv.handler.counter.Add(1)
	defer srv.handler.counter.Done()

	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "Server.Serve")
			srv.log(logger.Fatal, err)
		}
	}()

	listener = netutil.LimitListener(listener, srv.maxConns)
	defer func() { _ = listener.Close() }()

	address := listener.Addr()
	network := address.Network()
	srv.addListenerAddress(&address)
	defer srv.deleteListenerAddress(&address)
	srv.logf(logger.Info, "serve over listener (%s %s)", network, address)
	defer srv.logf(logger.Info, "listener closed (%s %s)", network, address)

	if srv.https {
		err = srv.server.ServeTLS(listener, "", "")
	} else {
		err = srv.server.Serve(listener)
	}
	if nettool.IsNetClosingError(err) || err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Addresses is used to get listener addresses.
func (srv *Server) Addresses() []net.Addr {
	srv.addressesRWM.RLock()
	defer srv.addressesRWM.RUnlock()
	addresses := make([]net.Addr, 0, len(srv.addresses))
	for address := range srv.addresses {
		addresses = append(addresses, *address)
	}
	return addresses
}

// Info is used to get http proxy server information.
//
// "https, address: [tcp 127.0.0.1:1999, tcp4 127.0.0.1:2001]"
// "http, address: [tcp 127.0.0.1:1999], auth: admin:123456"
func (srv *Server) Info() string {
	buf := new(bytes.Buffer)
	// protocol
	if srv.https {
		buf.WriteString("https")
	} else {
		buf.WriteString("http")
	}
	// listener address
	addresses := srv.Addresses()
	l := len(addresses)
	if l > 0 {
		buf.WriteString(", address: [")
		for i := 0; i < l; i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			network := addresses[i].Network()
			address := addresses[i].String()
			_, _ = fmt.Fprintf(buf, "%s %s", network, address)
		}
		buf.WriteString("]")
	}
	// username and password
	var (
		user string
		pass string
	)
	username := srv.handler.username
	if username != nil {
		user = username.Get()
		defer username.Put(user)
	}
	password := srv.handler.password
	if password != nil {
		pass = password.Get()
		defer password.Put(pass)
	}
	if user != "" || pass != "" {
		_, _ = fmt.Fprintf(buf, ", auth: %s:%s", user, pass)
	}
	return buf.String()
}

// Close is used to close HTTP proxy server.
func (srv *Server) Close() error {
	err := srv.server.Close()
	srv.handler.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		return err
	}
	return nil
}

type handler struct {
	logger logger.Logger
	logSrc string

	// dial timeout
	timeout time.Duration

	// proxy server http client
	transport *http.Transport

	// secondary proxy
	dialContext nettool.DialContext

	username *security.String
	password *security.String

	ctx     context.Context
	cancel  context.CancelFunc
	counter xsync.Counter
}

// [2018-11-27 00:00:00] [info] <http proxy-tag> test log
// remote: 127.0.0.1:1234
// POST /index HTTP/1.1
// Host: github.com
// Accept: text/html
// Connection: keep-alive
// User-Agent: Mozilla
//
// post data...
// post data...
func (h *handler) log(lv logger.Level, r *http.Request, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = httptool.FprintRequest(buf, r)
	h.logger.Println(lv, h.logSrc, buf)
}

// ServeHTTP implement http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			h.log(logger.Fatal, r, xpanic.Print(rec, "server.ServeHTTP()"))
		}
	}()
	if !h.authenticate(w, r) {
		return
	}
	// <security> remove Proxy-Authorization for
	// prevent log it or remote server watch it.
	r.Header.Del("Proxy-Authorization")
	h.log(logger.Info, r, "handle request")
	if r.Method == http.MethodConnect {
		h.handleConnectRequest(w, r)
	} else {
		h.handleCommonRequest(w, r)
	}
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) bool {
	if h.username == nil && h.password == nil {
		return true
	}
	authInfo := strings.Split(r.Header.Get("Proxy-Authorization"), " ")
	if len(authInfo) != 2 {
		h.failedToAuth(w)
		return false
	}
	authMethod := authInfo[0]
	authBase64 := authInfo[1]
	switch authMethod {
	case "Basic":
		auth, err := base64.StdEncoding.DecodeString(authBase64)
		if err != nil {
			h.log(logger.Exploit, r, "invalid basic base64 data:", err)
			h.failedToAuth(w)
			return false
		}
		userPass := strings.SplitN(string(auth), ":", 2)
		if len(userPass) == 1 {
			userPass = append(userPass, "")
		}
		var (
			eUser []byte
			ePass []byte
		)
		user := []byte(userPass[0])
		pass := []byte(userPass[1])
		if h.username != nil {
			eUser = h.username.GetBytes()
			defer h.username.PutBytes(eUser)
		}
		if h.password != nil {
			ePass = h.password.GetBytes()
			defer h.password.PutBytes(ePass)
		}
		userErr := subtle.ConstantTimeCompare(eUser, user) != 1
		passErr := subtle.ConstantTimeCompare(ePass, pass) != 1
		if userErr || passErr {
			userInfo := fmt.Sprintf("%s:%s", user, pass)
			h.log(logger.Exploit, r, "invalid username or password:", userInfo)
			h.failedToAuth(w)
			return false
		}
		return true
	default:
		h.log(logger.Exploit, r, "unsupported authentication method:", authMethod)
		h.failedToAuth(w)
		return false
	}
}

func (h *handler) failedToAuth(w http.ResponseWriter) {
	w.Header().Set("Proxy-Authenticate", "Basic")
	w.WriteHeader(http.StatusProxyAuthRequired)
}

func (h *handler) handleConnectRequest(w http.ResponseWriter, r *http.Request) {
	// check http.ResponseWriter is implemented http.Hijacker
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return
	}

	// dial target
	ctx, cancel := context.WithTimeout(h.ctx, h.timeout)
	defer cancel()
	remote, err := h.dialContext(ctx, "tcp", r.URL.Host)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		h.log(logger.Error, r, "failed to connect target", err)
		return
	}
	defer func() {
		err = remote.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			h.log(logger.Error, r, "failed to close remote connection:", err)
		}
	}()

	// hijack client conn
	wc, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log(logger.Error, r, "failed to hijack:", err)
		return
	}
	defer func() {
		err = wc.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			h.log(logger.Error, r, "failed to close hijacked connection:", err)
		}
	}()

	// write connection established response
	_, err = wc.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		h.log(logger.Error, r, "failed to write response:", err)
		return
	}

	const title = "handler.handleConnectRequest"
	// http.Server.Close() will not close hijacked conn
	// so we need create a goroutine to close it if the
	// handler.ctx is Done.
	done := make(chan struct{})
	defer close(done)
	h.counter.Add(1)
	go func() {
		defer h.counter.Done()
		defer func() {
			if rec := recover(); rec != nil {
				h.log(logger.Fatal, r, xpanic.Print(rec, title))
			}
		}()
		select {
		case <-done:
		case <-h.ctx.Done():
			_ = wc.Close()
			_ = remote.Close()
		}
	}()

	// reset deadline
	_ = remote.SetDeadline(time.Time{})
	_ = wc.SetDeadline(time.Time{})

	// start copy
	h.counter.Add(1)
	go func() {
		defer h.counter.Done()
		defer func() {
			if rec := recover(); rec != nil {
				h.log(logger.Fatal, r, xpanic.Print(rec, title))
			}
		}()
		_, _ = io.Copy(wc, remote)
	}()
	_, _ = io.Copy(remote, wc)
}

func (h *handler) handleCommonRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()
	resp, err := h.transport.RoundTrip(r.WithContext(ctx))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		h.log(logger.Error, r, err)
		return
	}
	defer func() {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	// copy header
	header := w.Header()
	for k, v := range resp.Header {
		header.Set(k, v[0])
	}
	// write status and copy body
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *handler) Close() {
	h.cancel()
	h.counter.Wait()
	h.transport.CloseIdleConnections()
}
