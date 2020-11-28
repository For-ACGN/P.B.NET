package xpprof

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/netutil"

	"project/internal/httptool"
	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/option"
	"project/internal/security"
	"project/internal/xpanic"
	"project/internal/xsync"
)

const (
	defaultTimeout  = 30 * time.Second
	defaultMaxConns = 1000
)

// Options contains options about pprof http server.
type Options struct {
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Timeout  time.Duration     `toml:"timeout"`
	MaxConns int               `toml:"max_conns"`
	Server   option.HTTPServer `toml:"server" testsuite:"-"`
}

// Server is a pprof tool over http server.
type Server struct {
	logger   logger.Logger
	https    bool
	maxConns int
	logSrc   string

	server  *http.Server
	handler *handler

	// listener addresses
	addresses    map[*net.Addr]struct{}
	addressesRWM sync.RWMutex
}

// NewHTTPServer is used to create a pprof tool over http server.
func NewHTTPServer(lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(lg, opts, false)
}

// NewHTTPSServer is used to create a pprof tool over https server.
func NewHTTPSServer(lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(lg, opts, true)
}

func newServer(lg logger.Logger, opts *Options, https bool) (*Server, error) {
	if opts == nil {
		opts = new(Options)
	}
	// apply http server option
	server, err := opts.Server.Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// log source
	var logSrc string
	if https {
		logSrc = "pprof-https"
	} else {
		logSrc = "pprof-http"
	}
	// initialize http handler
	handler := &handler{
		logger: lg,
		logSrc: logSrc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	handler.mux = mux
	if opts.Username != "" {
		if strings.Contains(opts.Username, ":") { // can not include ":"
			return nil, errors.New("username can not include character \":\"")
		}
		handler.username = security.NewString(opts.Username)
	}
	if opts.Password != "" {
		// validate password is a bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(opts.Password), []byte("123456"))
		if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
			return nil, errors.New("invalid bcrypt hash about password")
		}
		handler.password = security.NewString(opts.Password)
	}
	// set http server
	server.Handler = handler
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = defaultTimeout
	}
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
	// set pprof server
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
		srv.maxConns = defaultMaxConns
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

// Info is used to get pprof http server information.
//
// "address: tcp 127.0.0.1:1999, tcp4 127.0.0.1:2001"
// "address: tcp 127.0.0.1:1999 auth: admin:bcrypt"
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

// Close is used to close pprof http server.
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

	mux *http.ServeMux // pprof handlers

	username *security.String // raw username
	password *security.String // bcrypt hash

	counter xsync.Counter
}

// [2018-11-27 00:00:00] [info] <pprof-http> test log
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
	// <security> remove Authorization for prevent log it
	r.Header.Del("Authorization")
	h.log(logger.Info, r, "handle request")
	h.mux.ServeHTTP(w, r)
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) bool {
	if h.username == nil && h.password == nil {
		return true
	}
	authInfo := strings.Split(r.Header.Get("Authorization"), " ")
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
		passErr := ePass != nil && bcrypt.CompareHashAndPassword(ePass, pass) != nil
		if userErr || passErr {
			auth := fmt.Sprintf("%s:%s", user, pass)
			h.log(logger.Exploit, r, "invalid username or password:", auth)
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
	w.Header().Set("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
}

func (h *handler) Close() {
	h.counter.Wait()
}
