package netmgr

import (
	"fmt"
	"time"

	"project/internal/convert"
	"project/internal/logger"
)

// ListenerStatus contains status about listener.
type ListenerStatus struct {
	Network    string    `json:"network"`
	Address    string    `json:"address"`
	EstConns   uint64    `json:"est_conns"`
	MaxConns   uint64    `json:"max_conns"`
	Listened   time.Time `json:"listened"`
	LastAccept time.Time `json:"last_accept"`
}

// String is used to get status about listener.
// Output:
// -------------------------listener status------------------------
// address:     tcp 127.0.0.1:1234
// connections: 123/10000 (est/max)
// listened:    2018-11-27 00:00:00 +08:00
// last accept: 2018-11-27 00:01:00 +08:00
// ----------------------------------------------------------------
func (ls *ListenerStatus) String() string {
	const format = `
------------------------listener status-------------------------
address:     %s %s
connections: %d/%d (est/max)
listened:    %s
last accept: %s
----------------------------------------------------------------`
	return fmt.Sprintf(format[1:],
		ls.Network, ls.Address,
		ls.EstConns, ls.MaxConns,
		ls.Listened.Format(logger.TimeLayout),
		ls.LastAccept.Format(logger.TimeLayout),
	)
}

// ConnStatus contains status about connection.
type ConnStatus struct {
	LocalNetwork  string    `json:"local_network"`
	LocalAddress  string    `json:"local_address"`
	RemoteNetwork string    `json:"remote_network"`
	RemoteAddress string    `json:"remote_address"`
	Sent          uint64    `json:"sent"`
	Received      uint64    `json:"received"`
	Established   time.Time `json:"established"`
	LastRead      time.Time `json:"last_read"`
	LastWrite     time.Time `json:"last_write"`
}

// String is used to get status about connection.
// Output:
// -----------------------connection status------------------------
// local addr:  tcp 127.0.0.1:1234
// remote addr: tcp 127.0.0.1:5678
// traffic:     123 Byte/1.101 KB (sent/recv)
// established: 2018-11-27 00:00:00 +08:00
// last read:   2018-11-27 00:01:00 +08:00
// last write:  2018-11-27 00:02:00 +08:00
// ----------------------------------------------------------------
func (cs *ConnStatus) String() string {
	const format = `
-----------------------connection status------------------------
local addr:  %s %s
remote addr: %s %s
traffic:     %s/%s (sent/recv)
established: %s
last read:   %s
last write:  %s
----------------------------------------------------------------
`
	return fmt.Sprintf(format[1:],
		cs.LocalNetwork, cs.LocalAddress,
		cs.RemoteNetwork, cs.RemoteAddress,
		convert.StorageUnit(cs.Sent),
		convert.StorageUnit(cs.Received),
		cs.Established.Format(logger.TimeLayout),
		cs.LastRead.Format(logger.TimeLayout),
		cs.LastWrite.Format(logger.TimeLayout),
	)
}
