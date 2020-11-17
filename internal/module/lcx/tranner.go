package lcx

import (
	"bytes"
	"context"
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

// Tranner is used to map port.
type Tranner struct {
	dstNetwork string
	dstAddress string
	logger     logger.Logger
	opts       *Options

	logSrc   string
	listener net.Listener // used to check is closed
	conns    map[*tConn]struct{}
	rwm      sync.RWMutex // include listener

	mu     sync.Mutex // for operation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewTranner is used to create a tranner.
func NewTranner(tag, dstNetwork, dstAddress string, lg logger.Logger, opts *Options) (*Tranner, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	if dstAddress == "" {
		return nil, errors.New("empty destination address")
	}
	_, err := net.ResolveTCPAddr(dstNetwork, dstAddress)
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
	logSrc := "lcx tran"
	if tag != EmptyTag {
		logSrc += "-" + tag
	}
	return &Tranner{
		dstNetwork: dstNetwork,
		dstAddress: dstAddress,
		logger:     lg,
		opts:       opts,
		logSrc:     logSrc,
		conns:      make(map[*tConn]struct{}),
	}, nil
}

// Start is used to started tranner.
func (t *Tranner) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.start()
}

func (t *Tranner) start() error {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	if t.listener != nil {
		return errors.New("already started lcx tran")
	}
	listener, err := net.Listen(t.opts.LocalNetwork, t.opts.LocalAddress)
	if err != nil {
		return err
	}
	listener = netutil.LimitListener(listener, t.opts.MaxConns)
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.wg.Add(1)
	go t.serve(listener)
	// prevent panic before here
	t.listener = listener
	return nil
}

// Stop is used to stop tranner.
func (t *Tranner) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stop()
	t.wg.Wait()
}

func (t *Tranner) stop() {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	if t.listener == nil {
		return
	}
	t.cancel()
	// close listener
	err := t.listener.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		address := t.listener.Addr()
		network := address.Network()
		const format = "failed to close listener (%s %s): %s"
		t.logf(logger.Error, format, network, address, err)
	}
	// close all connections
	for conn := range t.conns {
		err = conn.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			t.log(logger.Error, "failed to close connection:", err)
		}
		delete(t.conns, conn)
	}
	// prevent panic before here
	t.listener = nil
}

// Restart is used to restart tranner.
func (t *Tranner) Restart() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stop()
	t.wg.Wait()
	return t.start()
}

// Name is used to get the module name.
func (t *Tranner) Name() string {
	return "lcx tran"
}

// Info is used to get the tranner information.
// "listen: tcp 0.0.0.0:1999, target: tcp 192.168.1.2:3389"
func (t *Tranner) Info() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	network := "unknown"
	address := "unknown"
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	if t.listener != nil {
		addr := t.listener.Addr()
		network = addr.Network()
		address = addr.String()
	}
	const format = "listen: %s %s, target: %s %s"
	_, _ = fmt.Fprintf(buf, format, network, address, t.dstNetwork, t.dstAddress)
	return buf.String()
}

// Status is used to return the tranner status.
// connections: 12/1000 (used/limit)
func (t *Tranner) Status() string {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	const format = "connections: %d/%d (used/limit)"
	_, _ = fmt.Fprintf(buf, format, len(t.conns), t.opts.MaxConns)
	return buf.String()
}

func (t *Tranner) logf(lv logger.Level, format string, log ...interface{}) {
	t.logger.Printf(lv, t.logSrc, format, log...)
}

func (t *Tranner) log(lv logger.Level, log ...interface{}) {
	t.logger.Println(lv, t.logSrc, log...)
}

func (t *Tranner) serve(listener net.Listener) {
	defer t.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			t.log(logger.Fatal, xpanic.Print(r, "Tranner.serve"))
		}
	}()

	address := listener.Addr()
	network := address.Network()

	defer func() {
		err := listener.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			const format = "failed to close listener (%s %s): %s"
			t.logf(logger.Error, format, network, address, err)
		}
	}()

	t.logf(logger.Info, "started listener (%s %s)", network, address)
	defer t.logf(logger.Info, "listener closed (%s %s)", network, address)

	// started accept loop
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
				t.logf(logger.Warning, "accept error: %s; retrying in %v", err, delay)
				time.Sleep(delay)
				continue
			}
			if !nettool.IsNetClosingError(err) {
				t.log(logger.Error, err)
			}
			return
		}
		delay = 0
		c := t.newConn(conn)
		c.Serve()
	}
}

func (t *Tranner) newConn(c net.Conn) *tConn {
	return &tConn{ctx: t, local: c}
}

func (t *Tranner) trackConn(conn *tConn, add bool) bool {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	if add {
		if t.listener == nil { // stopped
			return false
		}
		t.conns[conn] = struct{}{}
	} else {
		delete(t.conns, conn)
	}
	return true
}

// tran connection
type tConn struct {
	ctx   *Tranner
	local net.Conn
}

func (c *tConn) log(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = logger.Conn(c.local).WriteTo(buf)
	c.ctx.log(lv, buf)
}

func (c *tConn) Serve() {
	c.ctx.wg.Add(1)
	go c.serve()
}

func (c *tConn) serve() {
	defer c.ctx.wg.Done()

	const title = "tConn.serve"
	defer func() {
		if r := recover(); r != nil {
			c.log(logger.Fatal, xpanic.Print(r, title))
		}
	}()

	defer func() {
		err := c.local.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			c.log(logger.Error, "failed to close local connection:", err)
		}
	}()

	var ok bool
	defer func() {
		if ok {
			buf := new(bytes.Buffer)
			_, _ = fmt.Fprintln(buf, "connection closed")
			_, _ = logger.Conn(c.local).WriteTo(buf)
			_, _ = fmt.Fprint(buf, "\n", c.ctx.Status())
			c.ctx.log(logger.Info, buf)
		} else {
			c.ctx.log(logger.Info, c.ctx.Status())
		}
	}()

	if !c.ctx.trackConn(c, true) {
		return
	}
	defer c.ctx.trackConn(c, false)

	// connect the target
	ctx, cancel := context.WithTimeout(c.ctx.ctx, c.ctx.opts.ConnectTimeout)
	defer cancel()
	network := c.ctx.dstNetwork
	address := c.ctx.dstAddress
	remote, err := new(net.Dialer).DialContext(ctx, network, address)
	if err != nil {
		c.log(logger.Error, "failed to connect target:", err)
		return
	}

	defer func() {
		err := remote.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			c.log(logger.Error, "failed to close remote connection:", err)
		}
	}()

	// print latest connection status
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, "connection established")
	_, _ = logger.Conn(c.local).WriteTo(buf)
	_, _ = fmt.Fprint(buf, "\n", c.ctx.Status())
	c.ctx.log(logger.Info, buf)

	// reset deadline
	_ = remote.SetDeadline(time.Time{})
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
		_, _ = io.Copy(c.local, remote)
	}()
	_, _ = io.Copy(remote, c.local)
	ok = true
}

func (c *tConn) Close() error {
	return c.local.Close()
}

// testAddress is used to get listener address, it only for test.
func (t *Tranner) testAddress() string {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	if t.listener == nil {
		return ""
	}
	return t.listener.Addr().String()
}
