package netmgr

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/guid"
	"project/internal/nettool"
)

// Listener is a net.Listener wrapper that spawned by Manager.TrackListener.
type Listener struct {
	ctx *Manager

	net.Listener
	now      func() time.Time
	guid     *guid.GUID
	listened time.Time

	estConns   uint64
	maxConns   uint64
	lastAccept time.Time
	rwm        *sync.RWMutex
	cond       *sync.Cond

	inShutdown int32
	closeOnce  sync.Once
}

func (mgr *Manager) newListener(listener net.Listener) *Listener {
	rwm := new(sync.RWMutex)
	return &Listener{
		ctx:      mgr,
		Listener: listener,
		now:      mgr.now,
		guid:     mgr.guid.Get(),
		listened: mgr.now(),
		maxConns: mgr.GetListenerMaxConns(),
		rwm:      rwm,
		cond:     sync.NewCond(rwm),
	}
}

// Accept is used to accept next connection and wrap it to net.Conn.
func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptEx()
}

// AcceptEx is used to accept next connection and wrap it to *Conn.
func (l *Listener) AcceptEx() (*Conn, error) {
	if !l.require() {
		return nil, errors.New("listener is closed")
	}
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	// track connection
	c := l.ctx.newConn(conn, l.release)
	l.ctx.addConn(c)
	// update counter
	l.rwm.Lock()
	defer l.rwm.Unlock()
	l.estConns++
	l.lastAccept = l.now()
	return c, nil
}

func (l *Listener) shuttingDown() bool {
	return atomic.LoadInt32(&l.inShutdown) != 0
}

func (l *Listener) require() bool {
	if l.shuttingDown() {
		return false
	}
	l.rwm.Lock()
	defer l.rwm.Unlock()
	if l.maxConns == 0 {
		return true
	}
	for l.estConns >= l.maxConns {
		if l.shuttingDown() {
			return false
		}
		l.cond.Wait()
	}
	return true
}

func (l *Listener) release() {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	l.estConns--
	l.cond.Signal()
}

// signal is used to signal l.cond if l.require() is blocked
func (l *Listener) signal() {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	l.cond.Signal()
}

// GUID is used to get the guid of the connection.
func (l *Listener) GUID() guid.GUID {
	return *l.guid
}

// GetMaxConns is used to get the maximum number of the established
// connection, zero value means no limit.
func (l *Listener) GetMaxConns() uint64 {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	return l.maxConns
}

// SetMaxConns is used to set the maximum number of the established
// connection, zero value means no limit.
func (l *Listener) SetMaxConns(n uint64) {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	l.maxConns = n
}

// GetEstConnsNum is used to get the number of the established connection.
func (l *Listener) GetEstConnsNum() uint64 {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	return l.estConns
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

// Close is used to close the listener.
func (l *Listener) Close() error {
	atomic.StoreInt32(&l.inShutdown, 1)
	l.signal()
	err := l.Listener.Close()
	if err != nil && !nettool.IsNetClosedError(err) {
		return err
	}
	l.closeOnce.Do(func() {
		l.ctx.deleteListener(l)
	})
	return nil
}
