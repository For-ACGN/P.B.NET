package netstat

import (
	"encoding/binary"
	"net"
	"unsafe"
)

// Netstat is used to get current connections about TCP and UDP.
type Netstat interface {
	GetTCP4Conns() ([]*TCP4Conn, error)
	GetTCP6Conns() ([]*TCP6Conn, error)
	GetUDP4Conns() ([]*UDP4Conn, error)
	GetUDP6Conns() ([]*UDP6Conn, error)
	Close() error
}

// TCP4Conn contains information about TCP Over IPv4 connection.
type TCP4Conn struct {
	LocalAddr  net.IP
	LocalPort  uint16
	RemoteAddr net.IP
	RemotePort uint16
	State      uint8
	PID        int64
	Process    string
}

// ID is used to identified this connection.
func (conn *TCP4Conn) ID() string {
	b := make([]byte, net.IPv4len+2+net.IPv4len+2)
	copy(b[:net.IPv4len], conn.LocalAddr)
	binary.BigEndian.PutUint16(b[net.IPv4len:], conn.LocalPort)
	copy(b[net.IPv4len+2:], conn.RemoteAddr)
	binary.BigEndian.PutUint16(b[net.IPv4len+2+net.IPv4len:], conn.RemotePort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// TCP6Conn contains information about TCP Over IPv6 connection.
type TCP6Conn struct {
	LocalAddr     net.IP
	LocalScopeID  uint32
	LocalPort     uint16
	RemoteAddr    net.IP
	RemoteScopeID uint32
	RemotePort    uint16
	State         uint8
	PID           int64
	Process       string
}

// ID is used to identified this connection.
func (conn *TCP6Conn) ID() string {
	b := make([]byte, net.IPv6len+4+2+net.IPv6len+4+2)
	copy(b[:net.IPv6len], conn.LocalAddr)
	binary.BigEndian.PutUint32(b[net.IPv6len:], conn.LocalScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4:], conn.LocalPort)
	copy(b[net.IPv6len+4+2:], conn.RemoteAddr)
	binary.BigEndian.PutUint32(b[net.IPv6len+4+2+net.IPv6len:], conn.RemoteScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4+2+net.IPv6len+4:], conn.RemotePort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// UDP4Conn contains information about UDP Over IPv4 connection.
type UDP4Conn struct {
	LocalAddr net.IP
	LocalPort uint16
	PID       int64
	Process   string
}

// ID is used to identified this connection.
func (conn *UDP4Conn) ID() string {
	b := make([]byte, net.IPv4len+2)
	copy(b[:net.IPv4len], conn.LocalAddr)
	binary.BigEndian.PutUint16(b[net.IPv4len:], conn.LocalPort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// UDP6Conn contains information about UDP Over IPv6 connection.
type UDP6Conn struct {
	LocalAddr    net.IP
	LocalScopeID uint32
	LocalPort    uint16
	PID          int64
	Process      string
}

// ID is used to identified this connection.
func (conn *UDP6Conn) ID() string {
	b := make([]byte, net.IPv6len+4+2)
	copy(b[:net.IPv6len], conn.LocalAddr)
	binary.BigEndian.PutUint32(b[net.IPv6len:], conn.LocalScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4:], conn.LocalPort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}
