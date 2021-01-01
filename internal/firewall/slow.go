package firewall

import (
	"context"
	"net"
	"sync"

	"project/internal/random"
)

// SlowConn is used to sleep random time when call Read and Write.
// It used to defend scanner, make scanner scan slow, for trigger
// Nmap's --host-timeout arguments, it will drop all result of host.
type SlowConn struct {
	net.Conn

	fixed        uint
	random       uint
	readSleeper  *random.Sleeper
	writeSleeper *random.Sleeper
	readMu       sync.Mutex
	writeMu      sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSlowConn is used to create a new slow connection, time unit is millisecond.
func NewSlowConn(conn net.Conn, fixed, rand uint) *SlowConn {
	sc := SlowConn{
		Conn:         conn,
		fixed:        fixed,
		random:       rand,
		readSleeper:  random.NewSleeper(),
		writeSleeper: random.NewSleeper(),
	}
	sc.ctx, sc.cancel = context.WithCancel(context.Background())
	return &sc
}

// Read is used to sleep random millisecond and read data from connection.
func (sc *SlowConn) Read(b []byte) (int, error) {
	sc.readMu.Lock()
	defer sc.readMu.Unlock()
	select {
	case <-sc.readSleeper.SleepMillisecond(sc.fixed, sc.random):
		return sc.Conn.Read(b)
	case <-sc.ctx.Done():
		return 0, sc.ctx.Err()
	}
}

// Write is used to sleep random millisecond and write data to connection.
func (sc *SlowConn) Write(b []byte) (int, error) {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()
	select {
	case <-sc.writeSleeper.SleepMillisecond(sc.fixed, sc.random):
		return sc.Conn.Write(b)
	case <-sc.ctx.Done():
		return 0, sc.ctx.Err()
	}
}

// Close is used to close connection, it will cancel context for interrupt
// sleep in Read or Write, it will also stop sleeper.
func (sc *SlowConn) Close() error {
	sc.cancel()
	err := sc.Conn.Close()
	sc.readMu.Lock()
	defer sc.readMu.Unlock()
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()
	sc.readSleeper.Stop()
	sc.writeSleeper.Stop()
	return err
}

// SlowListener is a listener wrapper, it will wrap accepted connection to SlowConn.
type SlowListener struct {
	net.Listener
	fixed  uint
	random uint
}

// NewSlowListener is used to create a slow listener.
func NewSlowListener(listener net.Listener, fixed, random uint) *SlowListener {
	return &SlowListener{Listener: listener, fixed: fixed, random: random}
}

// Accept is used to accept connection and wrap SlowConn.
func (sl *SlowListener) Accept() (net.Conn, error) {
	conn, err := sl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewSlowConn(conn, sl.fixed, sl.random), nil
}
