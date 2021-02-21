package netmgr

import (
	"net"
)

// Listener is a net.Listener wrapper that spawn by Manager.
type Listener struct {
	net.Listener
}
