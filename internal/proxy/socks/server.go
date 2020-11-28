package socks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/netutil"

	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/security"
	"project/internal/xpanic"
	"project/internal/xsync"
)

// EmptyTag is a reserve tag that delete "-" in tag,
// "https proxy- " -> "https proxy", it is used to tool/proxy.
const EmptyTag = " "

// ErrServerClosed is returned by the Server's Serve, ListenAndServe,
// methods after a call Close.
var ErrServerClosed = fmt.Errorf("socks server closed")

// Server implemented internal/proxy.server.
type Server struct {
	logger     logger.Logger
	socks4     bool
	disableExt bool   // socks4 can't resolve domain name
	protocol   string // "socks5", "socks4a", "socks4"
	logSrc     string

	// options
	username *security.String
	password *security.String
	userID   *security.String
	timeout  time.Duration
	maxConns int

	// secondary proxy
	dialContext nettool.DialContext

	listeners  map[*net.Listener]struct{}
	conns      map[*conn]struct{}
	inShutdown int32
	rwm        sync.RWMutex

	ctx     context.Context
	cancel  context.CancelFunc
	counter xsync.Counter
}

// NewSocks5Server is used to create a socks5 server.
func NewSocks5Server(tag string, lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(tag, lg, opts, false, false)
}

// NewSocks4aServer is used to create a socks4a server.
func NewSocks4aServer(tag string, lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(tag, lg, opts, true, false)
}

// NewSocks4Server is used to create a socks4 server.
func NewSocks4Server(tag string, lg logger.Logger, opts *Options) (*Server, error) {
	return newServer(tag, lg, opts, true, true)
}

func newServer(tag string, lg logger.Logger, opts *Options, socks4, disableExt bool) (*Server, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	if opts == nil {
		opts = new(Options)
	}
	srv := Server{
		logger:      lg,
		socks4:      socks4,
		disableExt:  disableExt,
		timeout:     opts.Timeout,
		maxConns:    opts.MaxConns,
		dialContext: opts.DialContext,
		listeners:   make(map[*net.Listener]struct{}, 1),
		conns:       make(map[*conn]struct{}, 16),
	}
	// select protocol
	switch {
	case !socks4:
		srv.protocol = "socks5"
	case socks4 && disableExt:
		srv.protocol = "socks4"
	case socks4 && !disableExt:
		srv.protocol = "socks4a"
	}
	// log source
	logSrc := srv.protocol
	if tag != EmptyTag {
		logSrc += "-" + tag
	}
	srv.logSrc = logSrc
	// authentication
	if opts.Username != "" || opts.Password != "" {
		srv.username = security.NewString(opts.Username)
		srv.password = security.NewString(opts.Password)
	}
	if opts.UserID != "" {
		srv.userID = security.NewString(opts.UserID)
	}
	if srv.timeout < 1 {
		srv.timeout = defaultConnectTimeout
	}
	if srv.maxConns < 1 {
		srv.maxConns = defaultMaxConnections
	}
	if srv.dialContext == nil {
		srv.dialContext = new(net.Dialer).DialContext
	}
	srv.ctx, srv.cancel = context.WithCancel(context.Background())
	return &srv, nil
}

func (srv *Server) logf(lv logger.Level, format string, log ...interface{}) {
	srv.logger.Printf(lv, srv.logSrc, format, log...)
}

func (srv *Server) log(lv logger.Level, log ...interface{}) {
	srv.logger.Println(lv, srv.logSrc, log...)
}

func (srv *Server) shuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *Server) trackListener(listener *net.Listener, add bool) bool {
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	if add {
		if srv.shuttingDown() {
			return false
		}
		srv.listeners[listener] = struct{}{}
		srv.counter.Add(1)
	} else {
		delete(srv.listeners, listener)
		srv.counter.Done()
	}
	return true
}

// ListenAndServe is used to listen a listener and serve.
func (srv *Server) ListenAndServe(network, address string) error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
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
	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "Server.Serve")
			srv.log(logger.Fatal, err)
		}
	}()

	address := listener.Addr()
	network := address.Network()

	listener = netutil.LimitListener(listener, srv.maxConns)
	defer func() {
		err := listener.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			const format = "failed to close listener (%s %s): %s"
			srv.logf(logger.Error, format, network, address, err)
		}
	}()

	if !srv.trackListener(&listener, true) {
		return ErrServerClosed
	}
	defer srv.trackListener(&listener, false)

	srv.logf(logger.Info, "serve over listener (%s %s)", network, address)
	defer srv.logf(logger.Info, "listener closed (%s %s)", network, address)

	// start accept loop
	const maxDelay = time.Second
	var delay time.Duration // how long to sleep on accept failure
	for {
		conn, err := listener.Accept()
		if err != nil {
			// check error
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if delay > maxDelay {
					delay = maxDelay
				}
				srv.logf(logger.Warning, "accept error: %s; retrying in %v", err, delay)
				time.Sleep(delay)
				continue
			}
			if nettool.IsNetClosingError(err) {
				return nil
			}
			srv.log(logger.Error, err)
			return err
		}
		delay = 0
		c := srv.newConn(conn)
		c.Serve()
	}
}

func (srv *Server) newConn(c net.Conn) *conn {
	return &conn{ctx: srv, local: c}
}

func (srv *Server) trackConn(conn *conn, add bool) bool {
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	if add {
		if srv.shuttingDown() {
			return false
		}
		srv.conns[conn] = struct{}{}
	} else {
		delete(srv.conns, conn)
	}
	return true
}

// Addresses is used to get listener addresses.
func (srv *Server) Addresses() []net.Addr {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	addresses := make([]net.Addr, 0, len(srv.listeners))
	for listener := range srv.listeners {
		addresses = append(addresses, (*listener).Addr())
	}
	return addresses
}

// Info is used to get socks server information.
// "socks5"
// "socks5, auth: admin:123456"
// "socks5, address: [tcp 127.0.0.1:1999], auth: admin:123456"
// "socks4a, address: [tcp 127.0.0.1:1999, tcp4 127.0.0.1:2001], user id: test"
// "socks4, address: [tcp 127.0.0.1:1999]"
func (srv *Server) Info() string {
	buf := new(bytes.Buffer)
	// protocol
	buf.WriteString(srv.protocol)
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
	if srv.socks4 {
		if srv.userID != nil {
			userID := srv.userID.Get()
			defer srv.userID.Put(userID)
			_, _ = fmt.Fprintf(buf, ", user id: %s", userID)
		}
	} else {
		if srv.username != nil {
			username := srv.username.Get()
			defer srv.username.Put(username)
			password := srv.password.Get()
			defer srv.password.Put(password)
			_, _ = fmt.Fprintf(buf, ", auth: %s:%s", username, password)
		}
	}
	return buf.String()
}

// Close is used to close socks server.
func (srv *Server) Close() error {
	err := srv.close()
	srv.counter.Wait()
	return err
}

func (srv *Server) close() error {
	atomic.StoreInt32(&srv.inShutdown, 1)
	srv.cancel()
	var err error
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	// close all listeners
	for listener := range srv.listeners {
		e := (*listener).Close()
		if e != nil && !nettool.IsNetClosingError(e) && err == nil {
			err = e
		}
		delete(srv.listeners, listener)
	}
	// close all connections
	for conn := range srv.conns {
		e := conn.Close()
		if e != nil && !nettool.IsNetClosingError(e) && err == nil {
			err = e
		}
		delete(srv.conns, conn)
	}
	return err
}

type conn struct {
	ctx    *Server
	local  net.Conn // listener accepted conn
	remote net.Conn // dial target host
}

func (conn *conn) logf(lv logger.Level, format string, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, format, log...)
	buf.WriteString("\n")
	nettool.FprintConn(buf, conn.local)
	conn.ctx.log(lv, buf)
}

func (conn *conn) log(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	nettool.FprintConn(buf, conn.local)
	conn.ctx.log(lv, buf)
}

func (conn *conn) Serve() {
	conn.ctx.counter.Add(1)
	go conn.serve()
}

func (conn *conn) serve() {
	defer conn.ctx.counter.Done()

	const title = "conn.serve()"
	defer func() {
		if r := recover(); r != nil {
			conn.log(logger.Fatal, xpanic.Print(r, title))
		}
	}()

	defer func() {
		err := conn.local.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			conn.log(logger.Error, "failed to close local connection:", err)
		}
	}()

	if !conn.ctx.trackConn(conn, true) {
		return
	}
	defer conn.ctx.trackConn(conn, false)

	// handle
	_ = conn.local.SetDeadline(time.Now().Add(conn.ctx.timeout))
	if conn.ctx.socks4 {
		conn.serveSocks4()
	} else {
		conn.serveSocks5()
	}
	if conn.remote == nil {
		return
	}

	defer func() {
		err := conn.remote.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			conn.log(logger.Error, "failed to close remote connection:", err)
		}
	}()

	// reset deadline
	_ = conn.remote.SetDeadline(time.Time{})
	_ = conn.local.SetDeadline(time.Time{})

	// start copy
	conn.ctx.counter.Add(1)
	go func() {
		defer conn.ctx.counter.Done()
		defer func() {
			if r := recover(); r != nil {
				conn.log(logger.Fatal, xpanic.Print(r, title))
			}
		}()
		_, _ = io.Copy(conn.local, conn.remote)
	}()
	_, _ = io.Copy(conn.remote, conn.local)
}

func (conn *conn) Close() error {
	return conn.local.Close()
}
