package netmgr

import (
	"net"
	"sync"
	"time"
)

// reference:
// Function LimitListener in golang.org/x/net/netutil

// Listener is a net.Listener wrapper that spawned by Manager.
type Listener struct {
	ctx *Manager

	net.Listener
	listened time.Time

	// about status
	estConns   uint64
	maxConns   uint64
	lastAccept time.Time
	statusRWM  sync.RWMutex

	// limit established connection
	sem    chan struct{}
	semRWM sync.RWMutex

	stopSignal chan struct{}
	closeOnce  sync.Once
}

func newListener(mgr *Manager, l net.Listener, max uint64) *Listener {
	now := mgr.now()
	return &Listener{
		ctx:        mgr,
		Listener:   l,
		listened:   now,
		maxConns:   max,
		lastAccept: now,
		sem:        make(chan struct{}, max),
		stopSignal: make(chan struct{}),
	}
}

// Accept is used to accept next connection and wrap it to net.Conn.
func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptEx()
}

// AcceptEx is used to accept next connection and wrap it to *Conn.
func (l *Listener) AcceptEx() (*Conn, error) {

	return nil, nil
}

func (l *Listener) acquire() bool {
	select {
	case l.sem <- struct{}{}:
		return true
	case <-l.stopSignal:
		return false
	}
}

// GetMaxConns is used to get the maximum number of the established connection.
func (l *Listener) GetMaxConns() {

}

// SetMaxConns is used to set the maximum number of the established connection.
func (l *Listener) SetMaxConns() {

}

// GetEstConnsNum is used to get the number of the established connection.
func (l *Listener) GetEstConnsNum() {

}

// GetLastAcceptTime is used to ge the time of last accepted connection.
func (l *Listener) GetLastAcceptTime() {

}

// Status is used to get status about listener.
func (l *Listener) Status() *ListenerStatus {
	addr := l.Listener.Addr()
	ls := ListenerStatus{
		Network:  addr.Network(),
		Address:  addr.String(),
		Listened: l.listened,
	}
	l.statusRWM.RLock()
	defer l.statusRWM.RUnlock()
	ls.EstConns = l.estConns
	ls.MaxConns = l.maxConns
	ls.LastAccept = l.lastAccept
	return &ls
}
