package lcx

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/netutil"

	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/xpanic"
)

// Listener is used to accept slaver's connection and user's connection.
type Listener struct {
	iNetwork string // slaver connected Listener
	iAddress string // slaver connected Listener
	logger   logger.Logger
	opts     *Options
	logSrc   string

	iListener net.Listener // accept slaver connection
	lListener net.Listener // accept user connection
	conns     map[*lConn]struct{}
	rwm       sync.RWMutex // include listener

	mu sync.Mutex // for operation
	wg sync.WaitGroup
}

// NewListener is used to create a listener.
func NewListener(tag, iNetwork, iAddress string, lg logger.Logger, opts *Options) (*Listener, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	if iAddress == "" {
		return nil, errors.New("empty income listener address")
	}
	_, err := net.ResolveTCPAddr(iNetwork, iAddress)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(Options)
	}
	opts = opts.apply()
	_, err = net.ResolveTCPAddr(opts.LocalNetwork, opts.LocalAddress)
	if err != nil {
		return nil, err
	}
	// log source
	logSrc := "lcx listen"
	if tag != EmptyTag {
		logSrc += "-" + tag
	}
	return &Listener{
		iNetwork: iNetwork,
		iAddress: iAddress,
		logger:   lg,
		opts:     opts,
		logSrc:   logSrc,
		conns:    make(map[*lConn]struct{}),
	}, nil
}

// Start is used to started listener.
func (l *Listener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.start()
}

func (l *Listener) start() error {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	if l.iListener != nil {
		return errors.New("already started lcx listen")
	}
	iListener, err := net.Listen(l.iNetwork, l.iAddress)
	if err != nil {
		return err
	}
	lListener, err := net.Listen(l.opts.LocalNetwork, l.opts.LocalAddress)
	if err != nil {
		return err
	}
	iListener = netutil.LimitListener(iListener, l.opts.MaxConns)
	lListener = netutil.LimitListener(lListener, l.opts.MaxConns)
	l.wg.Add(1)
	go l.serve(iListener, lListener)
	// prevent panic before here
	l.iListener = iListener
	l.lListener = lListener
	return nil
}

// Stop is used to stop listener.
func (l *Listener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stop()
	l.wg.Wait()
}

func (l *Listener) stop() {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	if l.iListener == nil {
		return
	}
	err := l.iListener.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		address := l.iListener.Addr()
		network := address.Network()
		const format = "failed to close income listener (%s %s): %s"
		l.logf(logger.Error, format, network, address, err)
	}
	err = l.lListener.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		address := l.lListener.Addr()
		network := address.Network()
		const format = "failed to close local listener (%s %s): %s"
		l.logf(logger.Error, format, network, address, err)
	}
	// close all connections
	for conn := range l.conns {
		err = conn.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			l.log(logger.Error, "failed to close connection:", err)
		}
		delete(l.conns, conn)
	}
	// prevent panic before here
	l.iListener = nil
	l.lListener = nil
}

// Restart is used to restart listener.
func (l *Listener) Restart() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stop()
	l.wg.Wait()
	return l.start()
}

// Name is used to get the module name.
func (l *Listener) Name() string {
	return "lcx listen"
}

// Info is used to get the listener information.
// "income: tcp 0.0.0.0:1999, local: tcp 127.0.0.1:19993"
func (l *Listener) Info() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	iNetwork := "unknown"
	iAddress := "unknown"
	lNetwork := "unknown"
	lAddress := "unknown"
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	if l.iListener != nil {
		addr := l.iListener.Addr()
		iNetwork = addr.Network()
		iAddress = addr.String()
	}
	if l.lListener != nil {
		addr := l.lListener.Addr()
		lNetwork = addr.Network()
		lAddress = addr.String()
	}
	const format = "income: %s %s, local: %s %s"
	_, _ = fmt.Fprintf(buf, format, iNetwork, iAddress, lNetwork, lAddress)
	return buf.String()
}

// Status is used to return the tranner status.
// connections: 12/1000 (used/limit)
func (l *Listener) Status() string {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	const format = "connections: %d/%d (used/limit)"
	_, _ = fmt.Fprintf(buf, format, len(l.conns), l.opts.MaxConns)
	return buf.String()
}

func (l *Listener) logf(lv logger.Level, format string, log ...interface{}) {
	l.logger.Printf(lv, l.logSrc, format, log...)
}

func (l *Listener) log(lv logger.Level, log ...interface{}) {
	l.logger.Println(lv, l.logSrc, log...)
}

func (l *Listener) serve(iListener, lListener net.Listener) {
	defer l.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			l.log(logger.Fatal, xpanic.Print(r, "Listener.serve"))
		}
	}()

	addr := iListener.Addr()
	iNetwork := addr.Network()
	iAddress := addr.String()

	addr = lListener.Addr()
	lNetwork := addr.Network()
	lAddress := addr.String()

	defer func() {
		err := iListener.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			const format = "failed to close income listener (%s %s): %s"
			l.logf(logger.Error, format, iNetwork, iAddress, err)
		}
		err = lListener.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			const format = "failed to close local listener (%s %s): %s"
			l.logf(logger.Error, format, lNetwork, lAddress, err)
		}
	}()

	const format = "started income and local listener (%s %s), (%s %s)"
	l.logf(logger.Info, format, iNetwork, iAddress, lNetwork, lAddress)
	defer func() {
		const format = "income and local listener closed (%s %s), (%s %s)"
		l.logf(logger.Info, format, iNetwork, iAddress, lNetwork, lAddress)
	}()

	// started accept loop
	buf := new(bytes.Buffer)
	for {
		// first accept remote connection
		remote := l.accept(iListener)
		if remote == nil {
			return
		}
		// log remote connection
		buf.Reset()
		_, _ = fmt.Fprintln(buf, "income slave connection")
		_, _ = logger.Conn(remote).WriteTo(buf)
		l.log(logger.Info, buf)
		// than accept local connection
		local := l.accept(lListener)
		if local == nil {
			_ = remote.Close()
			return
		}
		// log local connection
		buf.Reset()
		_, _ = fmt.Fprintln(buf, "income user connection")
		_, _ = logger.Conn(local).WriteTo(buf)
		l.log(logger.Info, buf)
		// serve
		c := l.newConn(remote, local)
		c.Serve()
	}
}

func (l *Listener) accept(listener net.Listener) net.Conn {
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
				l.logf(logger.Warning, "accept error: %s; retrying in %v", err, delay)
				time.Sleep(delay)
				continue
			}
			if !nettool.IsNetClosingError(err) {
				l.log(logger.Error, err)
			}
			return nil
		}
		return conn
	}
}

func (l *Listener) newConn(remote, local net.Conn) *lConn {
	return &lConn{ctx: l, remote: remote, local: local}
}

func (l *Listener) trackConn(conn *lConn, add bool) bool {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	if add {
		if l.iListener == nil { // stopped
			return false
		}
		l.conns[conn] = struct{}{}
	} else {
		delete(l.conns, conn)
	}
	return true
}

// listener connection
type lConn struct {
	ctx    *Listener
	remote net.Conn // slaver income connection
	local  net.Conn // user income connection
}

func (c *lConn) log(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = logger.Conn(c.remote).WriteTo(buf)
	c.ctx.log(lv, buf)
}

func (c *lConn) Serve() {
	c.ctx.wg.Add(1)
	go c.serve()
}

func (c *lConn) serve() {
	defer c.ctx.wg.Done()

	const title = "lConn.serve"
	defer func() {
		if r := recover(); r != nil {
			c.log(logger.Fatal, xpanic.Print(r, title))
		}
	}()

	defer func() {
		err := c.remote.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			c.log(logger.Error, "failed to close remote connection:", err)
		}
		err = c.local.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			c.log(logger.Error, "failed to close local connection:", err)
		}
	}()

	buf := new(bytes.Buffer)
	defer func() {
		buf.Reset()
		_, _ = fmt.Fprintln(buf, "connection closed")
		_, _ = logger.Conn(c.local).WriteTo(buf)
		_, _ = fmt.Fprint(buf, "\n", c.ctx.Status())
		c.ctx.log(logger.Info, buf)
	}()

	if !c.ctx.trackConn(c, true) {
		return
	}
	defer c.ctx.trackConn(c, false)

	// print latest connection status
	_, _ = fmt.Fprintln(buf, "connection established")
	_, _ = logger.Conn(c.local).WriteTo(buf)
	_, _ = fmt.Fprint(buf, "\n", c.ctx.Status())
	c.ctx.log(logger.Info, buf)

	// reset deadline
	_ = c.remote.SetDeadline(time.Time{})
	_ = c.local.SetDeadline(time.Time{})

	// started copy
	c.ctx.wg.Add(1)
	go func() {
		defer c.ctx.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				c.log(logger.Fatal, xpanic.Print(r, title))
			}
		}()
		_, _ = io.Copy(c.local, c.remote)
	}()
	_, _ = io.Copy(c.remote, c.local)
}

func (c *lConn) Close() error {
	lErr := c.local.Close()
	rErr := c.remote.Close()
	if rErr != nil && lErr == nil {
		return rErr
	}
	return lErr
}

// testIncomeAddress is used to get income listener address, it only for test.
func (l *Listener) testIncomeAddress() string {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	if l.iListener == nil {
		return ""
	}
	return l.iListener.Addr().String()
}

// testLocalAddress is used to get local listener address, it only for test.
func (l *Listener) testLocalAddress() string {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	if l.lListener == nil {
		return ""
	}
	return l.lListener.Addr().String()
}
