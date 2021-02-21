package netmgr

import (
	"sync"

	"project/internal/guid"
)

// Manager is the network manager, it used to store status about listeners
// and connections. It can close the listeners and connections by guid,
// limit established connections about each listener, it also can set the
// upload and download rate.
type Manager struct {
	listeners    map[guid.GUID]*Listener
	listenersRWM sync.RWMutex
	conns        map[guid.GUID]*Conn
	connsRWM     sync.RWMutex
}

// NewManager is used to create a new network manager.
func NewManager() {

}

// WrapListener is used to wrap a net.Listener to a limited listener.
func (mgr *Manager) WrapListener() {

}

// WrapConn is used to wrap a net.Conn to a rate limited connection.
func (mgr *Manager) WrapConn() {

}

func (mgr *Manager) CloseListener() error {

	return nil
}

func (mgr *Manager) CloseConn() error {
	return nil
}
