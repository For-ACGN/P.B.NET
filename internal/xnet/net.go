package xnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"project/internal/nettool"
	"project/internal/xnet/light"
	"project/internal/xnet/quic"
	"project/internal/xnet/xtls"
)

// supported modes
const (
	ModeQUIC  = "quic"
	ModeLight = "light"
	ModeTLS   = "tls"
	ModeTCP   = "tcp"
	ModePipe  = "pipe"
)

var defaultNetwork = map[string]string{
	ModeQUIC:  "udp",
	ModeLight: "tcp",
	ModeTLS:   "tcp",
	ModeTCP:   "tcp",
	ModePipe:  "pipe",
}

// errors about check network
var (
	ErrEmptyMode    = fmt.Errorf("empty mode")
	ErrEmptyNetwork = fmt.Errorf("empty network")
)

// CheckModeNetwork is used to check if the mode and network matched.
func CheckModeNetwork(mode string, network string) error {
	if mode == "" {
		return ErrEmptyMode
	}
	if network == "" {
		return ErrEmptyNetwork
	}
	switch mode {
	case ModeQUIC:
		switch network {
		case "udp", "udp4", "udp6":
			return nil
		}
	case ModeLight:
		switch network {
		case "tcp", "tcp4", "tcp6":
			return nil
		}
	case ModeTLS:
		switch network {
		case "tcp", "tcp4", "tcp6":
			return nil
		}
	case ModeTCP:
		switch network {
		case "tcp", "tcp4", "tcp6":
			return nil
		}
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
	return fmt.Errorf("mismatched mode and network: %s %s", mode, network)
}

// Listener contains a net.Listener and listener's mode.
type Listener struct {
	net.Listener
	mode string
	now  func() time.Time
}

// AcceptEx is used to accept *Conn, role will use it.
func (l *Listener) AcceptEx() (*Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(conn, l.mode, l.now()), nil
}

// Mode is used to get the listener mode.
func (l *Listener) Mode() string {
	return l.mode
}

// String is used to return listener information.
// tls (tcp 127.0.0.1:443)
// quic (udp 127.0.0.1:443)
func (l *Listener) String() string {
	addr := l.Listener.Addr()
	return fmt.Sprintf("%s (%s %s)", l.mode, addr.Network(), addr)
}

// Options contains options about all modes.
type Options struct {
	TLSConfig   *tls.Config         // tls, quic need it
	Password    []byte              // kcp need it
	Salt        []byte              // kcp need it
	Timeout     time.Duration       // handshake timeout
	DialContext nettool.DialContext // for proxy
	Now         func() time.Time    // get connect time
}

// Listen is used to listen a listener.
func Listen(mode, network, address string, opts *Options) (*Listener, error) {
	err := CheckModeNetwork(mode, network)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(Options)
	}
	var listener net.Listener
	switch mode {
	case ModeQUIC:
		listener, err = quic.Listen(network, address, opts.TLSConfig, opts.Timeout)
	case ModeLight:
		listener, err = light.Listen(network, address, opts.Timeout)
	case ModeTLS:
		listener, err = xtls.Listen(network, address, opts.TLSConfig)
	case ModeTCP:
		listener, err = net.Listen(network, address)
	}
	if err != nil {
		return nil, err
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	return &Listener{Listener: listener, mode: mode, now: now}, nil
}

// Dial is used to dial context with context.Background().
func Dial(mode, network, address string, opts *Options) (*Conn, error) {
	return DialContext(context.Background(), mode, network, address, opts)
}

// DialContext is used to dial with context.
func DialContext(ctx context.Context, mode, network, address string, opts *Options) (*Conn, error) {
	err := CheckModeNetwork(mode, network)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(Options)
	}
	var conn net.Conn
	switch mode {
	case ModeQUIC:
		conn, err = quic.DialContext(ctx, network, address, opts.TLSConfig, opts.Timeout)
	case ModeLight:
		conn, err = light.DialContext(ctx, network, address, opts.Timeout, opts.DialContext)
	case ModeTLS:
		conn, err = xtls.DialContext(ctx, network, address, opts.TLSConfig, opts.Timeout, opts.DialContext)
	case ModeTCP:
		conn, err = (&net.Dialer{Timeout: opts.Timeout}).DialContext(ctx, network, address)
	}
	if err != nil {
		return nil, err
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	return NewConn(conn, mode, now()), nil
}
