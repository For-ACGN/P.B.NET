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
	read    uint64
	written uint64
	rwm     sync.RWMutex

	established time.Time
}

func (c *Conn) Close() error {

	return c.Conn.Close()
}

// Status is used to get status about connection.
// LocalAddress maybe changed, such as QUIC.
func (c *Conn) Status() *ConnStatus {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return &ConnStatus{
		LocalNetwork:  c.LocalAddr().Network(),
		LocalAddress:  c.LocalAddr().String(),
		RemoteNetwork: c.RemoteAddr().Network(),
		RemoteAddress: c.RemoteAddr().String(),
		Read:          c.read,
		Written:       c.written,
		Established:   c.established,
	}
}
