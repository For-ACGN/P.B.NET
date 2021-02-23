package netmgr

import (
	"net"
	"sync"
	"time"

	"project/internal/guid"
)

// reference:
// Function LimitListener in golang.org/x/net/netutil

// Listener is a net.Listener wrapper that spawned by Manager.
type Listener struct {
	ctx *Manager

	net.Listener
	now      func() time.Time
	guid     *guid.GUID
	listened time.Time

	estConns   uint64
	maxConns   uint64
	lastAccept time.Time
	semaphore  chan struct{}
	rwm        sync.RWMutex

	stopSignal chan struct{}
	closeOnce  sync.Once
}

func (mgr *Manager) newListener(listener net.Listener) *Listener {
	now := mgr.now()
	maxConns := mgr.GetListenerMaxConns()
	return &Listener{
		ctx:        mgr,
		Listener:   listener,
		now:        mgr.now,
		guid:       mgr.guid.Get(),
		listened:   now,
		maxConns:   maxConns,
		lastAccept: now,
		semaphore:  make(chan struct{}, maxConns),
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

// Close is used to close the listener.
func (l *Listener) Close() error {
	return l.Listener.Close()
}

// GUID is used to get the guid of the connection.
func (l *Listener) GUID() guid.GUID {
	return *l.guid
}

func (l *Listener) acquire() bool {
	select {
	case l.semaphore <- struct{}{}:
		return true
	case <-l.stopSignal:
		return false
	}
}

func (l *Listener) release() {

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

// Status is used to get status about listener.
func (l *Listener) Status() *ListenerStatus {
	addr := l.Listener.Addr()
	ls := ListenerStatus{
		Network:  addr.Network(),
		Address:  addr.String(),
		Listened: l.listened,
	}
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	ls.EstConns = l.estConns
	ls.MaxConns = l.maxConns
	ls.LastAccept = l.lastAccept
	return &ls
}
