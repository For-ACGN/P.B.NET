package netmgr

import (
	"net"
	"sync"
	"time"
)

// Conn is a net.Conn wrapper that spawn by Listener.
type Conn struct {
	net.Conn

	// imprecise
	sent     uint64
	received uint64
	rwm      sync.RWMutex

	estAt time.Time
}

// Status is used to get connection status.
// address maybe changed, such as QUIC.
func (c *Conn) Status() *ConnStatus {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return &ConnStatus{
		LocalNetwork:  c.LocalAddr().Network(),
		LocalAddress:  c.LocalAddr().String(),
		RemoteNetwork: c.RemoteAddr().Network(),
		RemoteAddress: c.RemoteAddr().String(),
		Sent:          c.sent,
		Received:      c.received,
		EstablishedAt: c.estAt,
	}
}
