package firewall

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
)

type channelListenerAddr struct {
	str string
}

func (*channelListenerAddr) Network() string {
	return "channel listener"
}

func (cla *channelListenerAddr) String() string {
	return cla.str
}

// ChannelListener is a special listener that accept connection
// from a net.Conn channel, use SendConn to add connection and
// wait ChannelListener call Accept.
type ChannelListener struct {
	connCh chan net.Conn
	addr   net.Addr
	ctx    context.Context
	cancel context.CancelFunc
}

// NewChannelListener is used to create a channel listener.
func NewChannelListener() *ChannelListener {
	cl := ChannelListener{
		connCh: make(chan net.Conn, 16),
	}
	cl.addr = &channelListenerAddr{
		str: fmt.Sprintf("channel listener pointer: 0x%X", &cl.connCh),
	}
	cl.ctx, cl.cancel = context.WithCancel(context.Background())
	return &cl
}

// SendConn is used to send a connection to listener connection channel.
func (cl *ChannelListener) SendConn(conn net.Conn) {
	select {
	case cl.connCh <- conn:
	case <-cl.ctx.Done():
	}
}

// Accept is used to receive a connection from listener connection channel.
func (cl *ChannelListener) Accept() (net.Conn, error) {
	select {
	case conn := <-cl.connCh:
		if conn != nil {
			return conn, nil
		}
		return nil, errors.New("accept nil net.Conn")
	case <-cl.ctx.Done():
		return nil, cl.ctx.Err()
	}
}

// Addr is used to get the pointer of the connection channel.
func (cl *ChannelListener) Addr() net.Addr {
	return cl.addr
}

// Close is used to close channel listener.
func (cl *ChannelListener) Close() error {
	cl.cancel()
	return nil
}
