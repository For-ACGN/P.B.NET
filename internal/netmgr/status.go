package netmgr

import (
	"fmt"
	"time"

	"project/internal/convert"
	"project/internal/logger"
)

// ConnStatus contains connection status.
type ConnStatus struct {
	LocalNetwork  string
	LocalAddress  string
	RemoteNetwork string
	RemoteAddress string
	Sent          uint64
	Received      uint64
	EstablishedAt time.Time
}

// String is used to get connection status.
//
// local:  tcp 127.0.0.1:123
// remote: tcp 127.0.0.1:124
// sent:   123 Byte,
// recv:   1.101 KB
// est at: 2018-11-27 00:00:00
func (status *ConnStatus) String() string {
	const format = "" +
		"local:  %s %s\n" +
		"remote: %s %s\n" +
		"sent:   %s\n" +
		"recv:   %s\n" +
		"est at: %s"
	return fmt.Sprintf(format,
		status.LocalNetwork, status.LocalAddress,
		status.RemoteNetwork, status.RemoteAddress,
		convert.StorageUnit(status.Sent),
		convert.StorageUnit(status.Received),
		status.EstablishedAt.Format(logger.TimeLayout),
	)
}
