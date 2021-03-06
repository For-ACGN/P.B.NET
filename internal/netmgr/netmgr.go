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

// Interface is used to wrap Manager and not export functions about Track and Close.
type Interface interface {
	// GetListener is used to get listener by guid.
	GetListener(guid *guid.GUID) (*Listener, error)

	// GetConn is used to get connection by guid.
	GetConn(guid *guid.GUID) (*Conn, error)

	// GetListenerMaxConnsByGUID is used to get listener maximum number of
	// the established connection by guid, zero value means no limit.
	GetListenerMaxConnsByGUID(guid *guid.GUID) (uint64, error)

	// SetListenerMaxConnsByGUID is used to set listener maximum number of
	// the established connection by guid, zero value means no limit.
	SetListenerMaxConnsByGUID(guid *guid.GUID, n uint64) error

	// GetListenerEstConnsNumByGUID is used to get listener the number of
	// the established connection by guid.
	GetListenerEstConnsNumByGUID(guid *guid.GUID) (uint64, error)

	// GetConnLimitRateByGUID is used to get connection read and write limit
	// rate by guid, zero value means no limit.
	GetConnLimitRateByGUID(guid *guid.GUID) (uint64, uint64, error)

	// SetConnLimitRateByGUID is used to set connection read and write limit
	// rate by guid, zero value means no limit.
	SetConnLimitRateByGUID(guid *guid.GUID, read, write uint64) error

	// GetConnReadLimitRateByGUID is used to get connection read limit rate
	// by guid, zero value means no limit.
	GetConnReadLimitRateByGUID(guid *guid.GUID) (uint64, error)

	// SetConnReadLimitRateByGUID is used to set connection read limit rate
	// by guid, zero value means no limit.
	SetConnReadLimitRateByGUID(guid *guid.GUID, n uint64) error

	// GetConnWriteLimitRateByGUID is used to get connection write limit rate
	// by guid, zero value means no limit.
	GetConnWriteLimitRateByGUID(guid *guid.GUID) (uint64, error)

	// SetConnWriteLimitRateByGUID is used to set connection write limit rate
	// by guid, zero value means no limit.
	SetConnWriteLimitRateByGUID(guid *guid.GUID, n uint64) error

	// GetListenerStatusByGUID is used to get listener status by guid.
	GetListenerStatusByGUID(guid *guid.GUID) (*ListenerStatus, error)

	// GetConnStatusByGUID is used to get connection status by guid.
	GetConnStatusByGUID(guid *guid.GUID) (*ConnStatus, error)

	// CloseListener is used to close listener by guid.
	CloseListener(guid *guid.GUID) error

	// CloseConn is used to close connection by guid.
	CloseConn(guid *guid.GUID) error

	// GetListenersNum is used to get the number of the listeners.
	GetListenersNum() int

	// GetConnsNum is used to get the number of the connections.
	GetConnsNum() int

	// Listeners is used to get all tracked listeners.
	Listeners() map[guid.GUID]*Listener

	// Conns is used to get all tracked connections.
	Conns() map[guid.GUID]*Conn

	// GetAllListenersStatus is used to get status about all listeners.
	GetAllListenersStatus() map[guid.GUID]*ListenerStatus

	// GetAllConnsStatus is used to get status about all connections.
	GetAllConnsStatus() map[guid.GUID]*ConnStatus

	// CloseAllListeners is used to close all listeners.
	CloseAllListeners() error

	// CloseAllConns is used to close all connections.
	CloseAllConns() error

	// GetListenerMaxConns is used to get the default maximum number of
	// connections that each listener can established, zero value means no limit.
	GetListenerMaxConns() uint64

	// SetListenerMaxConns is used to set the default maximum number of
	// connections that each listener can established, zero value means
	// no limit. Only subsequent new tracked listeners will be affected.
	SetListenerMaxConns(n uint64)

	// GetConnLimitRate is used to get the default read and write limit
	// rate of connection, zero value means no limit.
	GetConnLimitRate() (read, write uint64)

	// SetConnLimitRate is used to set the default read and write limit
	// rate of connection, zero value means no limit. Only subsequent
	// new tracked connections will be affected.
	SetConnLimitRate(read, write uint64)

	// GetConnReadLimitRate is used to get the default read limit rate of
	// connection, zero value means no limit.
	GetConnReadLimitRate() uint64

	// SetConnReadLimitRate is used to set the default read limit rate of
	// connection, zero value means no limit. Only subsequent new tracked
	// connections will be affected.
	SetConnReadLimitRate(n uint64)

	// GetConnWriteLimitRate is used to get the default write limit rate of
	// connection, zero value means no limit.
	GetConnWriteLimitRate() uint64

	// SetConnWriteLimitRate is used to set the default write limit rate of
	// connection, zero value means no limit. Only subsequent new tracked
	// connections will be affected.
	SetConnWriteLimitRate(n uint64)
}

// Manager is the network manager, it used to store status about listeners
// and connections. It can close the listeners and connections by guid,
// limit established connections about each listener, it also can set the
// send and receive limit rate about each connection.
type Manager struct {
	now func() time.Time

	// for generate key about listener and connection
	guidGen *guid.Generator

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
		guidGen:   guid.NewGenerator(512, now),
		listeners: make(map[guid.GUID]*Listener, 8),
		conns:     make(map[guid.GUID]*Conn, 64),
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

// GetListenerMaxConnsByGUID is used to get listener maximum number of
// the established connection by guid, zero value means no limit.
func (mgr *Manager) GetListenerMaxConnsByGUID(guid *guid.GUID) (uint64, error) {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return 0, err
	}
	return listener.GetMaxConns(), nil
}

// SetListenerMaxConnsByGUID is used to set listener maximum number of
// the established connection by guid, zero value means no limit.
func (mgr *Manager) SetListenerMaxConnsByGUID(guid *guid.GUID, n uint64) error {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return err
	}
	listener.SetMaxConns(n)
	return nil
}

// GetListenerEstConnsNumByGUID is used to get listener the number of
// the established connection by guid.
func (mgr *Manager) GetListenerEstConnsNumByGUID(guid *guid.GUID) (uint64, error) {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return 0, err
	}
	return listener.GetEstConnsNum(), nil
}

// GetConnLimitRateByGUID is used to get connection read and write limit
// rate by guid, zero value means no limit.
func (mgr *Manager) GetConnLimitRateByGUID(guid *guid.GUID) (uint64, uint64, error) {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return 0, 0, err
	}
	read, write := conn.GetLimitRate()
	return read, write, nil
}

// SetConnLimitRateByGUID is used to set connection read and write limit
// rate by guid, zero value means no limit.
func (mgr *Manager) SetConnLimitRateByGUID(guid *guid.GUID, read, write uint64) error {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return err
	}
	conn.SetLimitRate(read, write)
	return nil
}

// GetConnReadLimitRateByGUID is used to get connection read limit rate
// by guid, zero value means no limit.
func (mgr *Manager) GetConnReadLimitRateByGUID(guid *guid.GUID) (uint64, error) {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return 0, err
	}
	return conn.GetReadLimitRate(), nil
}

// SetConnReadLimitRateByGUID is used to set connection read limit rate
// by guid, zero value means no limit.
func (mgr *Manager) SetConnReadLimitRateByGUID(guid *guid.GUID, n uint64) error {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return err
	}
	conn.SetReadLimitRate(n)
	return nil
}

// GetConnWriteLimitRateByGUID is used to get connection write limit rate
// by guid, zero value means no limit.
func (mgr *Manager) GetConnWriteLimitRateByGUID(guid *guid.GUID) (uint64, error) {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return 0, err
	}
	return conn.GetWriteLimitRate(), nil
}

// SetConnWriteLimitRateByGUID is used to set connection write limit rate
// by guid, zero value means no limit.
func (mgr *Manager) SetConnWriteLimitRateByGUID(guid *guid.GUID, n uint64) error {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return err
	}
	conn.SetWriteLimitRate(n)
	return nil
}

// GetListenerStatusByGUID is used to get listener status by guid.
func (mgr *Manager) GetListenerStatusByGUID(guid *guid.GUID) (*ListenerStatus, error) {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return nil, err
	}
	return listener.Status(), nil
}

// GetConnStatusByGUID is used to get connection status by guid.
func (mgr *Manager) GetConnStatusByGUID(guid *guid.GUID) (*ConnStatus, error) {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return nil, err
	}
	return conn.Status(), nil
}

// CloseListener is used to close listener by guid.
func (mgr *Manager) CloseListener(guid *guid.GUID) error {
	listener, err := mgr.GetListener(guid)
	if err != nil {
		return err
	}
	return listener.Close()
}

// CloseConn is used to close connection by guid.
func (mgr *Manager) CloseConn(guid *guid.GUID) error {
	conn, err := mgr.GetConn(guid)
	if err != nil {
		return err
	}
	return conn.Close()
}

// GetListenersNum is used to get the number of the listeners.
func (mgr *Manager) GetListenersNum() int {
	mgr.listenersRWM.RLock()
	defer mgr.listenersRWM.RUnlock()
	return len(mgr.listeners)
}

// GetConnsNum is used to get the number of the connections.
func (mgr *Manager) GetConnsNum() int {
	mgr.connsRWM.RLock()
	defer mgr.connsRWM.RUnlock()
	return len(mgr.conns)
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

// GetAllListenersStatus is used to get status about all listeners.
func (mgr *Manager) GetAllListenersStatus() map[guid.GUID]*ListenerStatus {
	listeners := mgr.Listeners()
	status := make(map[guid.GUID]*ListenerStatus, len(listeners))
	for key, listener := range listeners {
		status[key] = listener.Status()
	}
	return status
}

// GetAllConnsStatus is used to get status about all connections.
func (mgr *Manager) GetAllConnsStatus() map[guid.GUID]*ConnStatus {
	conns := mgr.Conns()
	status := make(map[guid.GUID]*ConnStatus, len(conns))
	for key, conn := range conns {
		status[key] = conn.Status()
	}
	return status
}

// CloseAllListeners is used to close all listeners.
func (mgr *Manager) CloseAllListeners() error {
	var err error
	for _, listener := range mgr.Listeners() {
		e := listener.Close()
		if e != nil && !nettool.IsNetClosedError(e) && err == nil {
			err = e
		}
	}
	return err
}

// CloseAllConns is used to close all connections.
func (mgr *Manager) CloseAllConns() error {
	var err error
	for _, conn := range mgr.Conns() {
		e := conn.Close()
		if e != nil && !nettool.IsNetClosedError(e) && err == nil {
			err = e
		}
	}
	return err
}

// GetListenerMaxConns is used to get the default maximum number of
// connections that each listener can established, zero value means no limit.
func (mgr *Manager) GetListenerMaxConns() uint64 {
	mgr.defaultConfigRWM.RLock()
	defer mgr.defaultConfigRWM.RUnlock()
	return mgr.maxConns
}

// SetListenerMaxConns is used to set the default maximum number of
// connections that each listener can established, zero value means
// no limit. Only subsequent new tracked listeners will be affected.
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
// rate of connection, zero value means no limit. Only subsequent
// new tracked connections will be affected.
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
// connection, zero value means no limit. Only subsequent new tracked
// connections will be affected.
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
// connection, zero value means no limit. Only subsequent new tracked
// connections will be affected.
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
		if e != nil && !nettool.IsNetClosedError(e) && err == nil {
			err = e
		}
		delete(mgr.listeners, key)
	}
	// close all connections
	mgr.connsRWM.Lock()
	defer mgr.connsRWM.Unlock()
	for key, conn := range mgr.conns {
		e := conn.Conn.Close()
		if e != nil && !nettool.IsNetClosedError(e) && err == nil {
			err = e
		}
		delete(mgr.conns, key)
	}
	mgr.guidGen.Stop()
	return err
}
