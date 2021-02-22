package netmgr

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Conn is a net.Conn wrapper that spawned by Listener.
type Conn struct {
	ctx *Manager

	net.Conn
	established time.Time
	release     func()

	// limit read and write rate
	readLimiter  *rate.Limiter
	writeLimiter *rate.Limiter

	// imprecise, only record data length by Read and Write
	// underlying data will not record like TCP, IP, Ethernet.
	read      uint64
	written   uint64
	lastRead  time.Time
	lastWrite time.Time
	rwm       sync.RWMutex

	context   context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func (mgr *Manager) newConn(conn net.Conn, release func()) *Conn {
	now := mgr.now()
	readLimitRate, writeLimitRate := mgr.GetLimitRate()
	readLimit := rate.Limit(readLimitRate)
	if readLimit == 0 {
		readLimit = rate.Inf
	}
	writeLimit := rate.Limit(writeLimitRate)
	if writeLimit == 0 {
		writeLimit = rate.Inf
	}
	c := Conn{
		ctx:          mgr,
		Conn:         conn,
		established:  now,
		release:      release,
		readLimiter:  rate.NewLimiter(readLimit, int(readLimitRate)),
		writeLimiter: rate.NewLimiter(writeLimit, int(writeLimitRate)),
		lastRead:     now,
		lastWrite:    now,
	}
	c.context, c.cancel = context.WithCancel(context.Background())
	return &c
}

// Read is used to read data from the connection.
func (c *Conn) Read(b []byte) (int, error) {

	return c.Conn.Read(b)
}

// Write is used to write data from the connection.
func (c *Conn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

// Close is used to close connection.
func (c *Conn) Close() error {
	c.closeOnce.Do(c.release)
	return c.Conn.Close()
}

// Status is used to get status about connection.
// LocalAddress maybe changed, such as QUIC.
func (c *Conn) Status() *ConnStatus {
	cs := ConnStatus{
		Established: c.established,
	}
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	cs.LocalNetwork = c.LocalAddr().Network()
	cs.LocalAddress = c.LocalAddr().String()
	cs.RemoteNetwork = c.RemoteAddr().Network()
	cs.RemoteAddress = c.RemoteAddr().String()
	cs.Read = c.read
	cs.Written = c.written
	return &cs
}
