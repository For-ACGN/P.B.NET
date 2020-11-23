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

	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/random"
	"project/internal/xpanic"
)

// Slaver is used to connect the target and connect the Listener.
type Slaver struct {
	lNetwork   string // Listener
	lAddress   string // Listener
	dstNetwork string // destination
	dstAddress string // destination
	logger     logger.Logger
	opts       *Options

	logSrc  string
	dialer  net.Dialer
	sleeper *random.Sleeper
	stopped bool
	online  bool // prevent record a lot of logs about dail failure
	conns   map[*sConn]struct{}
	rwm     sync.RWMutex

	mu     sync.Mutex // for operation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSlaver is used to create a slaver.
func NewSlaver(
	tag string,
	lNetwork string,
	lAddress string,
	dstNetwork string,
	dstAddress string,
	logger logger.Logger,
	opts *Options,
) (*Slaver, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	if lAddress == "" {
		return nil, errors.New("empty listener address")
	}
	if dstAddress == "" {
		return nil, errors.New("empty destination address")
	}
	_, err := net.ResolveTCPAddr(lNetwork, lAddress)
	if err != nil {
		return nil, err
	}
	_, err = net.ResolveTCPAddr(dstNetwork, dstAddress)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(Options)
	}
	opts = opts.apply()
	// log source
	logSrc := "lcx slave"
	if tag != EmptyTag {
		logSrc += "-" + tag
	}
	return &Slaver{
		lNetwork:   lNetwork,
		lAddress:   lAddress,
		dstNetwork: dstNetwork,
		dstAddress: dstAddress,
		logger:     logger,
		opts:       opts,
		logSrc:     logSrc,
		sleeper:    random.NewSleeper(),
		stopped:    true,
		conns:      make(map[*sConn]struct{}),
	}, nil
}

// Start is used to started slaver.
func (s *Slaver) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.start()
}

func (s *Slaver) start() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	if !s.stopped {
		return errors.New("already started lcx slave")
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)
	go s.serve()
	s.stopped = false
	return nil
}

// Stop is used to stop slaver.
func (s *Slaver) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stop()
	s.wg.Wait()
}

func (s *Slaver) stop() {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	if s.stopped {
		return
	}
	s.cancel()
	// close all connections
	for conn := range s.conns {
		err := conn.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			s.log(logger.Error, "failed to close connection:", err)
		}
		delete(s.conns, conn)
	}
	// prevent panic before here
	s.stopped = true
}

// Restart is used to restart slaver.
func (s *Slaver) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stop()
	s.wg.Wait()
	return s.start()
}

// Name is used to get the module name.
func (s *Slaver) Name() string {
	return "lcx slave"
}

// Info is used to get the slaver information.
// "listener: tcp 0.0.0.0:1999, target: tcp 192.168.1.2:3389"
func (s *Slaver) Info() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	const format = "listener: %s %s, target: %s %s"
	_, _ = fmt.Fprintf(buf, format, s.lNetwork, s.lAddress, s.dstNetwork, s.dstAddress)
	return buf.String()
}

// Status is used to return the slaver status.
// connections: 12/1000 (used/limit)
func (s *Slaver) Status() string {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	const format = "connections: %d/%d (used/limit)"
	_, _ = fmt.Fprintf(buf, format, len(s.conns), s.opts.MaxConns)
	return buf.String()
}

func (s *Slaver) logf(lv logger.Level, format string, log ...interface{}) {
	s.logger.Printf(lv, s.logSrc, format, log...)
}

func (s *Slaver) log(lv logger.Level, log ...interface{}) {
	s.logger.Println(lv, s.logSrc, log...)
}

func (s *Slaver) serve() {
	defer s.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			s.log(logger.Fatal, xpanic.Print(r, "Slaver.serve"))
		}
	}()

	s.logf(logger.Info, "started connect listener (%s %s)", s.lNetwork, s.lAddress)
	defer s.logf(logger.Info, "stop connect listener (%s %s)", s.lNetwork, s.lAddress)

	// dial loop
	for {
		if s.full() {
			if s.online {
				s.log(logger.Warning, "full connection")
				s.online = false
			}
			select {
			case <-s.sleeper.Sleep(1, 3):
			case <-s.ctx.Done():
				return
			}
			continue
		}
		if s.isStopped() {
			return
		}
		conn, err := s.connectToListener()
		if err != nil {
			if s.online {
				s.log(logger.Error, "failed to connect listener:", err)
				s.online = false
			}
			select {
			case <-s.sleeper.Sleep(1, 10):
			case <-s.ctx.Done():
				return
			}
			continue
		}
		c := s.newConn(conn)
		c.Serve()
		s.online = true
	}
}

func (s *Slaver) full() bool {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return len(s.conns) >= s.opts.MaxConns
}

func (s *Slaver) isStopped() bool {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.stopped
}

func (s *Slaver) connectToListener() (net.Conn, error) {
	ctx, cancel := context.WithTimeout(s.ctx, s.opts.DialTimeout)
	defer cancel()
	return s.dialer.DialContext(ctx, s.lNetwork, s.lAddress)
}

func (s *Slaver) trackConn(conn *sConn, add bool) bool {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	if add {
		if s.stopped {
			return false
		}
		s.conns[conn] = struct{}{}
	} else {
		delete(s.conns, conn)
	}
	return true
}

// salver connection
type sConn struct {
	ctx   *Slaver
	local net.Conn
}

func (s *Slaver) newConn(c net.Conn) *sConn {
	return &sConn{
		ctx:   s,
		local: c,
	}
}

func (c *sConn) log(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	nettool.FprintConn(buf, c.local)
	c.ctx.log(lv, buf)
}

func (c *sConn) Serve() {
	done := make(chan byte, 1+1+2)
	c.ctx.wg.Add(1)
	go c.serve(done)
	// receive read and write ok or finish signal.
	var state byte // local or remote read ok, send 1,
	select {
	case state = <-done:
	case <-c.ctx.ctx.Done():
	}
	select {
	case state = <-done:
	case <-c.ctx.ctx.Done():
	}
	// print current status
	if state == 1 {
		c.ctx.log(logger.Info, c.ctx.Status(), "connection established")
	}
}

func (c *sConn) serve(done chan<- byte) {
	defer c.ctx.wg.Done()
	defer c.sendFailedState(done)
	defer func() {
		if r := recover(); r != nil {
			c.log(logger.Fatal, xpanic.Print(r, "sConn.serve"))
			// must wait or make dial storm
			time.Sleep(time.Second)
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
			c.ctx.log(logger.Info, c.ctx.Status(), "connection closed")
		} else {
			c.ctx.log(logger.Info, c.ctx.Status(), "connection killed")
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

	// start another goroutine to copy
	c.ctx.wg.Add(1)
	go c.serveRemote(done, remote)

	// read one byte for block it, prevent slaver burst connect listener.
	oneByte := make([]byte, 1)
	_ = c.local.SetReadDeadline(time.Now().Add(10 * time.Minute))
	_, err = c.local.Read(oneByte)
	if err != nil {
		return
	}
	_ = remote.SetWriteDeadline(time.Now().Add(c.ctx.opts.ConnectTimeout))
	_, err = remote.Write(oneByte)
	if err != nil {
		c.log(logger.Error, "failed to write to remote connection:", err)
		return
	}

	// send local connection read ok signal
	select {
	case done <- 1:
	case <-c.ctx.ctx.Done():
		return
	}

	// continue copy
	_ = c.local.SetReadDeadline(time.Time{})
	_ = remote.SetWriteDeadline(time.Time{})

	_, _ = io.Copy(remote, c.local)
	ok = true
}

func (c *sConn) serveRemote(done chan<- byte, remote net.Conn) {
	defer c.ctx.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			c.log(logger.Fatal, xpanic.Print(r, "sConn.serveRemote"))
		}
	}()

	// read one byte for block it, prevent slaver burst connect listener.
	oneByte := make([]byte, 1)
	_ = remote.SetReadDeadline(time.Now().Add(10 * time.Minute))
	_, err := remote.Read(oneByte)
	if err != nil {
		return
	}
	_ = c.local.SetWriteDeadline(time.Now().Add(c.ctx.opts.ConnectTimeout))
	_, err = c.local.Write(oneByte)
	if err != nil {
		c.log(logger.Error, "failed to write to listener connection:", err)
		return
	}

	// send remote connection read ok signal
	select {
	case done <- 1:
	case <-c.ctx.ctx.Done():
		return
	}

	// continue copy
	_ = remote.SetReadDeadline(time.Time{})
	_ = c.local.SetWriteDeadline(time.Time{})

	_, _ = io.Copy(c.local, remote)
}

func (c *sConn) sendFailedState(done chan<- byte) {
	select {
	case done <- 0:
	case <-c.ctx.ctx.Done():
	}
	select {
	case done <- 0:
	case <-c.ctx.ctx.Done():
	}
}

func (c *sConn) Close() error {
	return c.local.Close()
}
