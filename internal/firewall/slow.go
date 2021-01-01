package firewall

import (
	"context"
	"net"

	"project/internal/random"
)

// SlowConn is used to sleep random time when call Read and Write.
// It used to defend scanner, make scanner scan slow.
type SlowConn struct {
	net.Conn
	fixed   uint
	random  uint
	sleeper *random.Sleeper
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewSlowConn is used to create a new slow connection.
func NewSlowConn(conn net.Conn, fixed, rand uint) *SlowConn {
	sc := SlowConn{
		Conn:    conn,
		fixed:   fixed,
		random:  rand,
		sleeper: random.NewSleeper(),
	}
	sc.ctx, sc.cancel = context.WithCancel(context.Background())
	return &sc
}

// Read is used to sleep random time and read data from connection.
func (sc *SlowConn) Read(b []byte) (int, error) {
	select {
	case <-sc.sleeper.Sleep(sc.fixed, sc.random):
		return sc.Conn.Read(b)
	case <-sc.ctx.Done():
		return 0, sc.ctx.Err()
	}
}

// Write is used to sleep random time and write data to connection.
func (sc *SlowConn) Write(b []byte) (int, error) {
	select {
	case <-sc.sleeper.Sleep(sc.fixed, sc.random):
		return sc.Conn.Write(b)
	case <-sc.ctx.Done():
		return 0, sc.ctx.Err()
	}
}

// Close is used to close connection.
func (sc *SlowConn) Close() error {
	sc.cancel()
	sc.sleeper.Stop()
	return sc.Conn.Close()
}

// SlowListener is used to create a slow listener to defend
// scanner, make scanner scan slow.
type SlowListener struct {
}

// NewSlowListener is used to create a slow listener.
func NewSlowListener() {

}
