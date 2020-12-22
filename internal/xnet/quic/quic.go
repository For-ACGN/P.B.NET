package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
)

const (
	defaultTimeout   = 30 * time.Second // dial and accept
	defaultNextProto = "h3-32"          // HTTP/3
)

// ErrConnClosed is an error about closed.
var ErrConnClosed = errors.New("connection closed")

// Conn implement net.Conn
type Conn struct {
	// must close rawConn manually to prevent goroutine leak
	// in package github.com/lucas-clemente/quic-go
	// go m.listen() in newPacketHandlerMap()
	rawConn net.PacketConn

	session quic.Session
	stream  quic.Stream

	// must use extra Mutex because SendStream
	// is not safe for use by multiple goroutines
	//
	// stream.Close() must not be called concurrently with Write()
	sendMu sync.Mutex

	// only server connection need it
	timeout    time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	acceptErr  error
	acceptOnce sync.Once
}

func (c *Conn) acceptStream() error {
	c.acceptOnce.Do(func() {
		if c.stream == nil {
			defer c.cancel()
			c.stream, c.acceptErr = c.session.AcceptStream(c.ctx)
			if c.acceptErr != nil {
				return
			}
			// read data for prevent block
			_ = c.stream.SetReadDeadline(time.Now().Add(c.timeout))
			_, c.acceptErr = c.stream.Read(make([]byte, 1))
		}
	})
	return c.acceptErr
}

// Read reads data from the connection.
func (c *Conn) Read(b []byte) (n int, err error) {
	err = c.acceptStream()
	if err != nil {
		return
	}
	return c.stream.Read(b)
}

// Write writes data to the connection.
func (c *Conn) Write(b []byte) (n int, err error) {
	err = c.acceptStream()
	if err != nil {
		return
	}
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	return c.stream.Write(b)
}

// Close is used to close connection.
func (c *Conn) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.acceptOnce.Do(func() {
		c.acceptErr = ErrConnClosed
	})
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.stream != nil {
		_ = c.stream.Close()
	}
	err := c.session.CloseWithError(0, "no error")
	if c.rawConn != nil {
		_ = c.rawConn.Close()
	}
	return err
}

// LocalAddr is used to get local address.
func (c *Conn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

// RemoteAddr is used to get remote address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

// SetDeadline is used to set read and write deadline.
func (c *Conn) SetDeadline(t time.Time) error {
	err := c.SetReadDeadline(t)
	if err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline is used to set read deadline.
func (c *Conn) SetReadDeadline(t time.Time) error {
	err := c.acceptStream()
	if err != nil {
		return err
	}
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline is used to set write deadline.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	err := c.acceptStream()
	if err != nil {
		return err
	}
	return c.stream.SetWriteDeadline(t)
}

type listener struct {
	rawConn net.PacketConn // see Conn
	quic.Listener
	timeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

func (l *listener) Accept() (net.Conn, error) {
	session, err := l.Listener.Accept(l.ctx)
	if err != nil {
		return nil, err
	}
	conn := Conn{
		session: session,
		timeout: l.timeout,
	}
	conn.ctx, conn.cancel = context.WithTimeout(l.ctx, l.timeout)
	return &conn, nil
}

func (l *listener) Close() error {
	l.cancel()
	err := l.Listener.Close()
	_ = l.rawConn.Close()
	return err
}

// Listen is used to create a listener.
func Listen(network, address string, config *tls.Config, timeout time.Duration) (net.Listener, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP(network, addr)
	if err != nil {
		return nil, err
	}
	if timeout < 1 {
		timeout = defaultTimeout
	}
	quicCfg := quic.Config{
		HandshakeIdleTimeout: timeout,
		MaxIdleTimeout:       timeout,
		KeepAlive:            true,
	}
	if len(config.NextProtos) == 0 {
		c := config.Clone()
		c.NextProtos = []string{defaultNextProto}
		config = c
	}
	quicListener, err := quic.Listen(conn, config, &quicCfg)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	listener := listener{
		rawConn:  conn,
		Listener: quicListener,
		timeout:  timeout,
	}
	listener.ctx, listener.cancel = context.WithCancel(context.Background())
	return &listener, nil
}

// Dial is used to dial a connection with context.Background().
func Dial(network, address string, config *tls.Config, timeout time.Duration) (*Conn, error) {
	return DialContext(context.Background(), network, address, config, timeout)
}

// DialContext is used to dial a connection with context.
func DialContext(
	ctx context.Context,
	network string,
	address string,
	config *tls.Config,
	timeout time.Duration,
) (*Conn, error) {
	rAddr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP(network, nil)
	if err != nil {
		return nil, err
	}
	var ok bool
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	if config == nil {
		config = new(tls.Config)
	}
	// set server name
	if config.ServerName == "" {
		colonPos := strings.LastIndex(address, ":")
		if colonPos == -1 {
			return nil, errors.New("missing port in address")
		}
		hostname := address[:colonPos]
		c := config.Clone()
		c.ServerName = hostname
		config = c
	}
	// set next protocols
	if len(config.NextProtos) == 0 {
		c := config.Clone()
		c.NextProtos = []string{defaultNextProto}
		config = c
	}
	// set quic configuration
	if timeout < 1 {
		timeout = defaultTimeout
	}
	quicCfg := quic.Config{
		HandshakeIdleTimeout: timeout,
		MaxIdleTimeout:       5 * timeout,
		KeepAlive:            true,
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	session, err := quic.DialContext(ctx, conn, rAddr, address, config, &quicCfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			_ = session.CloseWithError(0, "no error")
		}
	}()
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			_ = stream.Close()
		}
	}()
	// write data for prevent block
	_ = stream.SetWriteDeadline(time.Now().Add(timeout))
	_, err = stream.Write([]byte{0})
	if err != nil {
		return nil, err
	}
	ok = true
	return &Conn{rawConn: conn, session: session, stream: stream}, nil
}
