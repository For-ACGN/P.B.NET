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

// Manager is the network manager, it used to store status about listeners
// and connections. It can close the listeners and connections by guid,
// limit established connections about each listener, it also can set the
// send and receive limit rate about each connection.
type Manager struct {
	now func() time.Time

	// for generate key about listener and connection
	guid *guid.Generator

	// default configuration
	readLimitRate    uint64 // connection
	writeLimitRate   uint64 // connection
	maxConns         uint64 // listener
	defaultConfigRWM sync.RWMutex

	listeners    map[guid.GUID]*Listener
	listenersRWM sync.RWMutex
	conns        map[guid.GUID]*Conn
	connsRWM     sync.RWMutex

	inShutdown int32
}

// New is used to create a new network manager.
func New(now func() time.Time) *Manager {
	if now == nil {
		now = time.Now
	}
	return &Manager{
		now:       now,
		guid:      guid.NewGenerator(512, now),
		listeners: make(map[guid.GUID]*Listener, 8),
		conns:     make(map[guid.GUID]*Conn, 1024),
	}
}

// TrackListener is used to wrap a net.Listener to a limit listener.
func (mgr *Manager) TrackListener(listener net.Listener) *Listener {
	ml := mgr.newListener(listener)
	mgr.addListener(ml)
	return ml
}

// TrackConn is used to wrap a net.Conn to a limit rate connection.
func (mgr *Manager) TrackConn(conn net.Conn) *Conn {
	mc := mgr.newConn(conn, nil)
	mgr.addConn(mc)
	return mc
}

// Listeners is used to get all tracked listeners.
func (mgr *Manager) Listeners() map[guid.GUID]*Listener {
	mgr.listenersRWM.RLock()
	defer mgr.listenersRWM.RUnlock()
	listeners := make(map[guid.GUID]*Listener, len(mgr.listeners))
	for key, listener := range mgr.listeners {
		listeners[key] = listener
	}
	return listeners
}

// Conns is used to get all tracked connections.
func (mgr *Manager) Conns() map[guid.GUID]*Conn {
	mgr.connsRWM.RLock()
	defer mgr.connsRWM.RUnlock()
	conns := make(map[guid.GUID]*Conn, len(mgr.conns))
	for key, conn := range mgr.conns {
		conns[key] = conn
	}
	return conns
}

// GetListener is used to get listener by guid.
func (mgr *Manager) GetListener(guid *guid.GUID) (*Listener, error) {
	mgr.listenersRWM.RLock()
	defer mgr.listenersRWM.RUnlock()
	if listener, ok := mgr.listeners[*guid]; ok {
		return listener, nil
	}
	return nil, errors.Errorf("listener: %s is not exist", guid)
}

// GetConn is used to get connection by guid.
func (mgr *Manager) GetConn(guid *guid.GUID) (*Conn, error) {
	mgr.connsRWM.RLock()
	defer mgr.connsRWM.RUnlock()
	if conn, ok := mgr.conns[*guid]; ok {
		return conn, nil
	}
	return nil, errors.Errorf("connection: %s is not exist", guid)
}

// KillListener is used to kill listener by guid.
func (mgr *Manager) KillListener(guid *guid.GUID) error {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return err
	}
	return listener.Close()
}

// KillConn is used to kill connection by guid.
func (mgr *Manager) KillConn(guid *guid.GUID) error {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return err
	}
	return conn.Close()
}

// GetListenerMaxConns is used to get the default maximum number of
// connections that each listener can established, zero value means no limit.
func (mgr *Manager) GetListenerMaxConns() uint64 {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.maxConns
}

// SetListenerMaxConns is used to set the default maximum number of
// connections that each listener can established, zero value means no limit.
func (mgr *Manager) SetListenerMaxConns(n uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.maxConns = n
}

// GetConnLimitRate is used to get the default read and write limit
// rate of connection, zero value means no limit.
func (mgr *Manager) GetConnLimitRate() (read, write uint64) {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.readLimitRate, mgr.writeLimitRate
}

// SetConnLimitRate is used to set the default read and write limit
// rate of connection, zero value means no limit.
func (mgr *Manager) SetConnLimitRate(read, write uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.readLimitRate = read
	mgr.writeLimitRate = write
}

// GetConnReadLimitRate is used to get the default read limit rate of
// connection, zero value means no limit.
func (mgr *Manager) GetConnReadLimitRate() uint64 {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.readLimitRate
}

// SetConnReadLimitRate is used to set the default read limit rate of
// connection, zero value means no limit.
func (mgr *Manager) SetConnReadLimitRate(n uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.readLimitRate = n
}

// GetConnWriteLimitRate is used to get the default write limit rate of
// connection, zero value means no limit.
func (mgr *Manager) GetConnWriteLimitRate() uint64 {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.writeLimitRate
}

// SetConnWriteLimitRate is used to set the default write limit rate of
// connection, zero value means no limit.
func (mgr *Manager) SetConnWriteLimitRate(n uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.writeLimitRate = n
}

func (mgr *Manager) shuttingDown() bool {
	return atomic.LoadInt32(&mgr.inShutdown) != 0
}

func (mgr *Manager) addListener(listener *Listener) {
	key := *listener.guid
	mgr.listenersRWM.Lock()
	defer mgr.listenersRWM.Unlock()
	if mgr.shuttingDown() {
		_ = listener.Listener.Close()
		return
	}
	mgr.listeners[key] = listener
}

func (mgr *Manager) addConn(conn *Conn) {
	key := *conn.guid
	mgr.connsRWM.Lock()
	defer mgr.connsRWM.Unlock()
	if mgr.shuttingDown() {
		_ = conn.Conn.Close()
		return
	}
	mgr.conns[key] = conn
}

func (mgr *Manager) deleteListener(listener *Listener) {
	key := *listener.guid
	mgr.listenersRWM.Lock()
	defer mgr.listenersRWM.Unlock()
	delete(mgr.listeners, key)
}

func (mgr *Manager) deleteConn(conn *Conn) {
	key := *conn.guid
	mgr.connsRWM.Lock()
	defer mgr.connsRWM.Unlock()
	delete(mgr.conns, key)
}

// Close is used to close network manager, it will close all tracked
// listeners and connections, if appear error, it will return the
// first error about close listener or connection.
func (mgr *Manager) Close() error {
	atomic.StoreInt32(&mgr.inShutdown, 1)
	var err error
	// close all listeners
	mgr.listenersRWM.Lock()
	defer mgr.listenersRWM.Unlock()
	for key, listener := range mgr.listeners {
		e := listener.Listener.Close()
		if e != nil && !nettool.IsNetClosingError(e) && err == nil {
			err = e
		}
		delete(mgr.listeners, key)
	}
	// close all connections
	mgr.connsRWM.Lock()
	defer mgr.connsRWM.Unlock()
	for key, conn := range mgr.conns {
		e := conn.Conn.Close()
		if e != nil && !nettool.IsNetClosingError(e) && err == nil {
			err = e
		}
		delete(mgr.conns, key)
	}
	mgr.guid.Close()
	return err
}
