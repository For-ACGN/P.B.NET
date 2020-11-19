package netstat

import (
	"encoding/binary"
	"fmt"
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
	LocalIP    net.IP
	LocalPort  uint16
	RemoteIP   net.IP
	RemotePort uint16
	State      uint8
	PID        int64
	Process    string
}

// ID is used to identified this connection.
func (conn *TCP4Conn) ID() string {
	b := make([]byte, net.IPv4len+2+net.IPv4len+2)
	copy(b[:net.IPv4len], conn.LocalIP)
	binary.BigEndian.PutUint16(b[net.IPv4len:], conn.LocalPort)
	copy(b[net.IPv4len+2:], conn.RemoteIP)
	binary.BigEndian.PutUint16(b[net.IPv4len+2+net.IPv4len:], conn.RemotePort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// LocalAddr is used to get the local address about this connection.
func (conn *TCP4Conn) LocalAddr() string {
	return fmt.Sprintf("%s:%d", conn.LocalIP, conn.LocalPort)
}

// RemoteAddr is used to get the remote address about this connection.
func (conn *TCP4Conn) RemoteAddr() string {
	return fmt.Sprintf("%s:%d", conn.RemoteIP, conn.RemotePort)
}

// TCP6Conn contains information about TCP Over IPv6 connection.
type TCP6Conn struct {
	LocalIP       net.IP
	LocalScopeID  uint32
	LocalPort     uint16
	RemoteIP      net.IP
	RemoteScopeID uint32
	RemotePort    uint16
	State         uint8
	PID           int64
	Process       string
}

// ID is used to identified this connection.
func (conn *TCP6Conn) ID() string {
	b := make([]byte, net.IPv6len+4+2+net.IPv6len+4+2)
	copy(b[:net.IPv6len], conn.LocalIP)
	binary.BigEndian.PutUint32(b[net.IPv6len:], conn.LocalScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4:], conn.LocalPort)
	copy(b[net.IPv6len+4+2:], conn.RemoteIP)
	binary.BigEndian.PutUint32(b[net.IPv6len+4+2+net.IPv6len:], conn.RemoteScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4+2+net.IPv6len+4:], conn.RemotePort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// LocalAddr is used to get the local address about this connection.
func (conn *TCP6Conn) LocalAddr() string {
	return fmt.Sprintf("[%s%%%d]:%d", conn.LocalIP, conn.LocalScopeID, conn.LocalPort)
}

// RemoteAddr is used to get the remote address about this connection.
func (conn *TCP6Conn) RemoteAddr() string {
	return fmt.Sprintf("[%s%%%d]:%d", conn.RemoteIP, conn.RemoteScopeID, conn.RemotePort)
}

// UDP4Conn contains information about UDP Over IPv4 connection.
type UDP4Conn struct {
	LocalIP   net.IP
	LocalPort uint16
	PID       int64
	Process   string
}

// ID is used to identified this connection.
func (conn *UDP4Conn) ID() string {
	b := make([]byte, net.IPv4len+2)
	copy(b[:net.IPv4len], conn.LocalIP)
	binary.BigEndian.PutUint16(b[net.IPv4len:], conn.LocalPort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// Addr is used to get the local address about this connection.
func (conn *UDP4Conn) Addr() string {
	return fmt.Sprintf("%s:%d", conn.LocalIP, conn.LocalPort)
}

// UDP6Conn contains information about UDP Over IPv6 connection.
type UDP6Conn struct {
	LocalIP      net.IP
	LocalScopeID uint32
	LocalPort    uint16
	PID          int64
	Process      string
}

// ID is used to identified this connection.
func (conn *UDP6Conn) ID() string {
	b := make([]byte, net.IPv6len+4+2)
	copy(b[:net.IPv6len], conn.LocalIP)
	binary.BigEndian.PutUint32(b[net.IPv6len:], conn.LocalScopeID)
	binary.BigEndian.PutUint16(b[net.IPv6len+4:], conn.LocalPort)
	return *(*string)(unsafe.Pointer(&b)) // #nosec
}

// Addr is used to get the local address about this connection.
func (conn *UDP6Conn) Addr() string {
	return fmt.Sprintf("[%s%%%d]:%d", conn.LocalIP, conn.LocalScopeID, conn.LocalPort)
}
