package light

import (
	"context"
	"net"
	"time"

	"project/internal/options"
)

func Server(ctx context.Context, conn net.Conn, timeout time.Duration) *Conn {
	return &Conn{ctx: ctx, Conn: conn, handshakeTimeout: timeout}
}

func Client(ctx context.Context, conn net.Conn, timeout time.Duration) *Conn {
	return &Conn{ctx: ctx, Conn: conn, handshakeTimeout: timeout, isClient: true}
}

type listener struct {
	net.Listener
	timeout time.Duration // handshake timeout
	ctx     context.Context
	cancel  context.CancelFunc
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return Server(l.ctx, conn, l.timeout), nil
}

func (l *listener) Close() error {
	l.cancel()
	return l.Listener.Close()
}

func Listen(network, address string, timeout time.Duration) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(l, timeout), nil
}

func NewListener(inner net.Listener, timeout time.Duration) net.Listener {
	l := listener{
		Listener: inner,
		timeout:  timeout,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	return &l
}

func Dial(
	ctx context.Context,
	network string,
	address string,
	timeout time.Duration,
	dial func(context.Context, string, string) (net.Conn, error),
) (*Conn, error) {
	if timeout < 1 {
		timeout = options.DefaultDialTimeout
	}
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if dial == nil {
		dial = new(net.Dialer).DialContext
	}
	conn, err := dial(dialCtx, network, address)
	if err != nil {
		return nil, err
	}
	return Client(ctx, conn, timeout), nil
}
