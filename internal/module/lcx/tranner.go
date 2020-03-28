package lcx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/netutil"

	"project/internal/logger"
	"project/internal/xpanic"
)

// Tranner is used to map port.
type Tranner struct {
	dstNetwork string // destination
	dstAddress string // destination
	logger     logger.Logger
	opts       *Options

	logSrc   string
	listener net.Listener
	conns    map[*tConn]struct{}
	rwm      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewTranner is used to create a tranner.
func NewTranner(
	tag string,
	dstNetwork string,
	dstAddress string,
	logger logger.Logger,
	opts *Options,
) (*Tranner, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
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
	return &Tranner{
		dstNetwork: dstNetwork,
		dstAddress: dstAddress,
		logger:     logger,
		opts:       opts,
		logSrc:     "lcx tran-" + tag,
		conns:      make(map[*tConn]struct{}),
	}, nil
}

// Start is used to start tranner.
func (t *Tranner) Start() error {
	t.rwm.Lock()
	defer t.rwm.Unlock()
	if t.listener != nil {
		return errors.New("already start lcx tranner")
	}
	listener, err := net.Listen(t.opts.LocalNetwork, t.opts.LocalAddress)
	if err != nil {
		return err
	}
	listener = netutil.LimitListener(listener, t.opts.MaxConns)
	t.listener = listener
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.wg.Add(1)
	go t.serve(listener)
	return nil
}

// Stop is used to stop tranner.
func (t *Tranner) Stop() {
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
	_ = t.listener.Close()
	t.listener = nil
	// close all connections
	for conn := range t.conns {
		_ = conn.Close()
	}
}

// Restart is used to restart tranner.
func (t *Tranner) Restart() error {
	t.Stop()
	return t.Start()
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

// testAddress is used to get listener address, it only for test.
func (t *Tranner) testAddress() string {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	if t.listener == nil {
		return ""
	}
	return t.listener.Addr().String()
}

func (t *Tranner) logf(lv logger.Level, format string, log ...interface{}) {
	t.logger.Printf(lv, t.logSrc, format, log...)
}

func (t *Tranner) log(lv logger.Level, log ...interface{}) {
	t.logger.Println(lv, t.logSrc, log...)
}

func (t *Tranner) serve(listener net.Listener) {
	defer func() { _ = listener.Close() }()
	address := listener.Addr()
	network := address.Network()
	defer func() {
		if r := recover(); r != nil {
			t.log(logger.Fatal, xpanic.Print(r, "Tranner.serve"))
		}
		t.logf(logger.Info, "listener closed (%s %s)", network, address)
		t.wg.Done()
	}()
	t.logf(logger.Info, "start listener (%s %s)", network, address)
	// start accept
	var delay time.Duration // how long to sleep on accept failure
	maxDelay := time.Second
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
			errStr := err.Error()
			if !strings.Contains(errStr, "closed") {
				t.log(logger.Error, errStr)
			}
			return
		}
		delay = 0
		t.newConn(conn).Serve()
	}
}

func (t *Tranner) newConn(c net.Conn) *tConn {
	return &tConn{
		tranner: t,
		local:   c,
	}
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

type tConn struct {
	tranner *Tranner
	local   net.Conn
}

func (c *tConn) log(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = logger.Conn(c.local).WriteTo(buf)
	c.tranner.log(lv, buf)
}

func (c *tConn) Serve() {
	c.tranner.wg.Add(1)
	go c.serve()
}

func (c *tConn) serve() {
	const title = "tConn.serve"
	defer func() {
		if r := recover(); r != nil {
			c.log(logger.Fatal, xpanic.Print(r, title))
		}
		_ = c.local.Close()
		c.tranner.wg.Done()
	}()

	if !c.tranner.trackConn(c, true) {
		return
	}
	defer c.tranner.trackConn(c, false)

	// connect the target
	ctx, cancel := context.WithTimeout(c.tranner.ctx, c.tranner.opts.ConnectTimeout)
	defer cancel()
	remote, err := new(net.Dialer).DialContext(ctx, c.tranner.dstNetwork, c.tranner.dstAddress)
	if err != nil {
		c.log(logger.Error, "failed to connect target:", err)
		return
	}
	defer func() { _ = remote.Close() }()

	// log
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, "income connection")
	_, _ = logger.Conn(c.local).WriteTo(buf)
	_, _ = fmt.Fprint(buf, "\n", c.tranner.Status())
	c.tranner.log(logger.Info, buf)

	// start copy
	_ = remote.SetDeadline(time.Time{})
	_ = c.local.SetDeadline(time.Time{})
	c.tranner.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log(logger.Fatal, xpanic.Print(r, title))
			}
			c.tranner.wg.Done()
		}()
		_, _ = io.Copy(c.local, remote)
	}()
	_, _ = io.Copy(remote, c.local)
}

func (c *tConn) Close() error {
	return c.local.Close()
}