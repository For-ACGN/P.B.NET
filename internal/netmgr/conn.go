package netmgr

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"project/internal/guid"
	"project/internal/nettool"
)

// Conn is a net.Conn wrapper that spawned by Listener.AcceptEx or Manager.TrackConn.
type Conn struct {
	ctx *Manager

	net.Conn
	release     func()
	now         func() time.Time
	guid        *guid.GUID
	established time.Time

	// limit read and write rate
	readLimiter    *rate.Limiter
	writeLimiter   *rate.Limiter
	readLimitRate  uint64
	writeLimitRate uint64

	// read and written are imprecise, only record data length
	// by Read and Write, the underlying data will not record
	// like TCP, IP, Ethernet.
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
	readLimitRate, writeLimitRate := mgr.GetConnLimitRate()
	readLimit := calcLimitRate(readLimitRate)
	writeLimit := calcLimitRate(writeLimitRate)
	readLimiter := rate.NewLimiter(readLimit, int(readLimit))
	writeLimiter := rate.NewLimiter(writeLimit, int(writeLimit))
	c := &Conn{
		ctx:            mgr,
		Conn:           conn,
		release:        release,
		now:            mgr.now,
		guid:           mgr.guid.Get(),
		established:    mgr.now(),
		readLimiter:    readLimiter,
		writeLimiter:   writeLimiter,
		readLimitRate:  readLimitRate,
		writeLimitRate: writeLimitRate,
	}
	c.context, c.cancel = context.WithCancel(context.Background())
	return c
}

// Read is used to read data from the connection.
func (c *Conn) Read(b []byte) (int, error) {
	err := c.readLimiter.WaitN(c.context, len(b))
	if err != nil {
		return 0, net.ErrClosed
	}
	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.read += uint64(n)
	c.lastRead = c.now()
	return n, nil
}

// Write is used to write data from the connection.
func (c *Conn) Write(b []byte) (int, error) {
	err := c.writeLimiter.WaitN(c.context, len(b))
	if err != nil {
		return 0, net.ErrClosed
	}
	n, err := c.Conn.Write(b)
	if err != nil {
		return n, err
	}
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.written += uint64(n)
	c.lastWrite = c.now()
	return n, nil
}

// GUID is used to get the guid of the connection.
func (c *Conn) GUID() guid.GUID {
	return *c.guid
}

// GetLimitRate is used to get read and write limit rate,
// zero value means no limit.
func (c *Conn) GetLimitRate() (read, write uint64) {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.readLimitRate, c.writeLimitRate
}

// SetLimitRate is used to set read and write limit rate,
// zero value means no limit.
func (c *Conn) SetLimitRate(read, write uint64) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.readLimitRate = read
	c.writeLimitRate = write
}

// GetReadLimitRate is used to get read limit rate,
// zero value means no limit.
func (c *Conn) GetReadLimitRate() uint64 {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.readLimitRate
}

// SetReadLimitRate is used to set read limit rate,
// zero value means no limit.
func (c *Conn) SetReadLimitRate(n uint64) {
	limit := calcLimitRate(n)
	c.readLimiter.SetLimit(limit)
	c.readLimiter.SetBurst(int(limit))
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.readLimitRate = n
}

// GetWriteLimitRate is used to get write limit rate,
// zero value means no limit.
func (c *Conn) GetWriteLimitRate() uint64 {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.writeLimitRate
}

// SetWriteLimitRate is used to set write limit rate,
// zero value means no limit.
func (c *Conn) SetWriteLimitRate(n uint64) {
	limit := calcLimitRate(n)
	c.writeLimiter.SetLimit(limit)
	c.writeLimiter.SetBurst(int(limit))
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.writeLimitRate = n
}

// Status is used to get status about connection.
// Local and remote address maybe changed, such as QUIC.
func (c *Conn) Status() *ConnStatus {
	localAddr := c.LocalAddr()
	remoteAddr := c.RemoteAddr()
	cs := ConnStatus{
		LocalNetwork:  localAddr.Network(),
		LocalAddress:  localAddr.String(),
		RemoteNetwork: remoteAddr.Network(),
		RemoteAddress: remoteAddr.String(),
		Established:   c.established,
	}
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	cs.ReadLimitRate = c.readLimitRate
	cs.WriteLimitRate = c.writeLimitRate
	cs.Read = c.read
	cs.Written = c.written
	cs.LastRead = c.lastRead
	cs.LastWrite = c.lastWrite
	return &cs
}

// Close is used to close connection.
func (c *Conn) Close() error {
	c.cancel()
	err := c.Conn.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		return err
	}
	c.closeOnce.Do(func() {
		if c.release != nil {
			c.release()
		}
		c.ctx.deleteConn(c)
	})
	return nil
}

func calcLimitRate(n uint64) rate.Limit {
	limit := rate.Limit(n)
	if limit == 0 {
		limit = rate.Inf
	}
	return limit
}
