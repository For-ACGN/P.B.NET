package netmgr

import (
	"sync"
	"time"

	"project/internal/guid"
)

// Manager is the network manager, it used to store status about listeners
// and connections. It can close the listeners and connections by guid,
// limit established connections about each listener, it also can set the
// upload and download rate.
type Manager struct {
	now func() time.Time

	readLimitRate  uint64
	writeLimitRate uint64
	limitRateRWM   sync.RWMutex

	guid         *guid.Generator
	listeners    map[guid.GUID]*Listener
	listenersRWM sync.RWMutex
	conns        map[guid.GUID]*Conn
	connsRWM     sync.RWMutex
}

// NewManager is used to create a new network manager.
func NewManager() *Manager {
	return nil
}

// GetLimitRate is used to get the connection read and write limit rate.
func (mgr *Manager) GetLimitRate() (read, write uint64) {
	mgr.limitRateRWM.RLock()
	defer mgr.limitRateRWM.RUnlock()
	return mgr.readLimitRate, mgr.writeLimitRate
}

// TraceListener is used to wrap a net.Listener to a limited listener.
func (mgr *Manager) TraceListener() {

}

// TraceConn is used to wrap a net.Conn to a rate limited connection.
func (mgr *Manager) TraceConn() {

}

func (mgr *Manager) addListener(listener *Listener) {

}

func (mgr *Manager) addConn(conn *Conn) {

}

func (mgr *Manager) CloseListener() error {

	return nil
}

func (mgr *Manager) CloseConn() error {
	return nil
}

func (mgr *Manager) deleteListener(guid guid.GUID) error {
	return nil
}

func (mgr *Manager) deleteConn(guid guid.GUID) error {
	return nil
}
