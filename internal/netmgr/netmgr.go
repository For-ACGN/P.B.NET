package netmgr

import (
	"sync"
	"sync/atomic"
	"time"

	"project/internal/guid"
)

const defaultListenerMaxConns = 4096

// Manager is the network manager, it used to store status about listeners
// and connections. It can close the listeners and connections by guid,
// limit established connections about each listener, it also can set the
// upload and download rate.
type Manager struct {
	now func() time.Time

	// for generate listener and connection key
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

// NewManager is used to create a new network manager.
func NewManager(now func() time.Time) *Manager {
	return &Manager{
		now:       now,
		guid:      guid.NewGenerator(4096, now),
		maxConns:  defaultListenerMaxConns,
		listeners: make(map[guid.GUID]*Listener, 8),
		conns:     make(map[guid.GUID]*Conn, 1024),
	}
}

// TraceListener is used to wrap a net.Listener to a limit listener.
func (mgr *Manager) TraceListener() {

}

// TraceConn is used to wrap a net.Conn to a limit rate connection.
func (mgr *Manager) TraceConn() {

}

// KillListener is used to kill listener by guid.
func (mgr *Manager) KillListener() error {

	return nil
}

// KillConn is used to kill connection by guid.
func (mgr *Manager) KillConn() error {

	return nil
}

// GetListenerMaxConns is used to get the default maximum number of
// connections that each listener can established.
func (mgr *Manager) GetListenerMaxConns() uint64 {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.maxConns
}

// GetConnLimitRate is used to get the default read and write limit
// rate of connection, zero means no limit.
func (mgr *Manager) GetConnLimitRate() (read, write uint64) {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.readLimitRate, mgr.writeLimitRate
}

// SetListenerMaxConns is used to set the default maximum number of
// connections that each listener can established.
func (mgr *Manager) SetListenerMaxConns(n uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.maxConns = n
}

// SetConnLimitRate is used to set the default read and write limit
// rate of connection, zero means no limit.
func (mgr *Manager) SetConnLimitRate(read, write uint64) {
	mgr.defaultConfigRWM.Lock()
	defer mgr.defaultConfigRWM.Unlock()
	mgr.readLimitRate = read
	mgr.writeLimitRate = write
}

func (mgr *Manager) shuttingDown() bool {
	return atomic.LoadInt32(&mgr.inShutdown) != 0
}

func (mgr *Manager) addListener(listener *Listener) {
	if mgr.shuttingDown() {
		_ = listener.Listener.Close()
		return
	}
	key := *listener.guid
	mgr.listenersRWM.Lock()
	defer mgr.listenersRWM.Unlock()
	mgr.listeners[key] = listener
}

func (mgr *Manager) addConn(conn *Conn) {
	if mgr.shuttingDown() {
		_ = conn.Conn.Close()
		return
	}
	key := *conn.guid
	mgr.connsRWM.Lock()
	defer mgr.connsRWM.Unlock()
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

// Close is used to close network manager, it will close all traced
// listeners and connections, if appear error, it will return the
// first error about close listener or connection.
func (mgr *Manager) Close() error {
	atomic.StoreInt32(&mgr.inShutdown, 1)

	mgr.guid.Close()
	return nil
}
