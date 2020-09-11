package api

import (
	"net"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"

	"project/internal/convert"
)

// references:
//
// DLL information, get exported functions list:
// http://xpdll.nirsoft.net/iphlpapi_dll.html GetExtendedTcpTable and GetExtendedUdpTable
//
// parameters about exported function:
// https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedtcptable
// https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedudptable

var (
	modIphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

	procGetExtendedTCPTable = modIphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUDPTable = modIphlpapi.NewProc("GetExtendedUdpTable")
)

// about TCP connection state, reference MIB_TCP_STATE
const (
	_ uint8 = iota
	TCPStateClosed
	TCPStateListening
	TCPStateSYNSent
	TCPStateSYNReceived
	TCPStateEstablished
	TCPStateFinWait1
	TCPStateFinWait2
	TCPStateCloseWait
	TCPStateClosing
	TCPStateLastAck
	TCPStateTimeWait
	TCPStateDeleteTCB
)

var tcpConnStates = map[uint8]string{
	TCPStateClosed:      "Closed",
	TCPStateListening:   "Listening",
	TCPStateSYNSent:     "SYN_Sent",
	TCPStateSYNReceived: "SYN_Received",
	TCPStateEstablished: "Established",
	TCPStateFinWait1:    "Fin_Wait1",
	TCPStateFinWait2:    "Fin_Wait2",
	TCPStateCloseWait:   "Close_Wait",
	TCPStateClosing:     "Closing",
	TCPStateLastAck:     "Last_Ack",
	TCPStateTimeWait:    "Time_Wait",
	TCPStateDeleteTCB:   "Delete_TCB",
}

// GetTCPConnState is used to convert state to string.
func GetTCPConnState(state uint8) string {
	return tcpConnStates[state]
}

// TCP4Conn contains information about TCP-over-IPv4 connection.
type TCP4Conn struct {
	LocalAddr  net.IP
	LocalPort  uint16
	RemoteAddr net.IP
	RemotePort uint16
	State      uint8
	PID        int64
	CreateTime time.Time
	ModuleInfo [16]int64 // 16 is TCP IP_OWNING_MODULE_SIZE
	Process    string    // process name
}

// TCP6Conn contains information about TCP-over-IPv6 connection.
type TCP6Conn struct {
	LocalAddr     net.IP
	LocalScopeID  uint32
	LocalPort     uint16
	RemoteAddr    net.IP
	RemoteScopeID uint32
	RemotePort    uint16
	State         uint8
	PID           int64
	CreateTime    time.Time
	ModuleInfo    [16]int64 // 16 is TCP IP_OWNING_MODULE_SIZE
	Process       string
}

// TCP table class
const (
	TCPTableBasicListener uint32 = iota
	TCPTableBasicConnections
	TCPTableBasicAll
	TCPTableOwnerPIDListener
	TCPTableOwnerPIDConnections
	TCPTableOwnerPIDAll
	TCPTableOwnerModuleListener
	TCPTableOwnerModuleConnections
	TCPTableOwnerModuleAll
)

// #nosec
func getTCPTable(ulAf, class uint32) ([]byte, error) {
	const maxAttemptTimes = 1024
	var (
		buffer   []byte
		tcpTable *byte
		dwSize   uint32
	)
	for i := 0; i < maxAttemptTimes; i++ {
		ret, _, _ := procGetExtendedTCPTable.Call(
			uintptr(unsafe.Pointer(tcpTable)), uintptr(unsafe.Pointer(&dwSize)),
			uintptr(uint32(1)), uintptr(ulAf), uintptr(class), uintptr(uint32(0)),
		)
		if ret != windows.NO_ERROR {
			if windows.Errno(ret) == windows.ERROR_INSUFFICIENT_BUFFER {
				buffer = make([]byte, dwSize)
				tcpTable = &buffer[0]
				continue
			}
			return nil, errors.WithStack(windows.Errno(ret))
		}
		return buffer, nil
	}
	return nil, errors.New("reach maximum attempt times")
}

// GetTCP4Conns is used to get TCP-over-IPv4 connections.
func GetTCP4Conns(class uint32) ([]*TCP4Conn, error) {
	buffer, err := getTCPTable(windows.AF_INET, class)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get tcp table")
	}
	var conns []*TCP4Conn
	switch {
	case class < 3:
		conns = parseTCP4TableBasic(buffer)
	case class < 6:
		conns = parseTCP4TableOwnerPID(buffer)
	case class < 9:
		conns, err = parseTCP4TableOwnerModule(buffer)
		if err != nil {
			return nil, err
		}
	default:
		panic("api/network: unreachable code")
	}
	return conns, nil
}

type tcp4TableBasic struct {
	n     uint32
	table [AnySize]tcp4RowBasic
}

type tcp4RowBasic struct {
	state      uint32
	localAddr  uint32
	localPort  [4]byte
	remoteAddr uint32
	remotePort [4]byte
}

func parseTCP4TableBasic(buffer []byte) []*TCP4Conn {
	table := (*tcp4TableBasic)(unsafe.Pointer(&buffer[0]))
	var rows []tcp4RowBasic
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rows))
	sh.Data = uintptr(unsafe.Pointer(&table.table))
	sh.Len = int(table.n)
	sh.Cap = int(table.n)
	l := len(rows)
	conns := make([]*TCP4Conn, l)
	for i := 0; i < l; i++ {
		conn := TCP4Conn{
			LocalAddr:  convert.LEUint32ToBytes(rows[i].localAddr),
			RemoteAddr: convert.LEUint32ToBytes(rows[i].remoteAddr),
			State:      uint8(rows[i].state),
		}
		conn.LocalPort = convert.BEBytesToUint16(rows[i].localPort[:2])
		conn.RemotePort = convert.BEBytesToUint16(rows[i].remotePort[:2])
		conns[i] = &conn
	}
	runtime.KeepAlive(table)
	return conns
}

type tcp4TableOwnerPID struct {
	n     uint32
	table [AnySize]tcp4RowOwnerPID
}

type tcp4RowOwnerPID struct {
	state      uint32
	localAddr  uint32
	localPort  [4]byte
	remoteAddr uint32
	remotePort [4]byte
	pid        uint32
}

func parseTCP4TableOwnerPID(buffer []byte) []*TCP4Conn {
	table := (*tcp4TableOwnerPID)(unsafe.Pointer(&buffer[0]))
	var rows []tcp4RowOwnerPID
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rows))
	sh.Data = uintptr(unsafe.Pointer(&table.table))
	sh.Len = int(table.n)
	sh.Cap = int(table.n)
	l := len(rows)
	conns := make([]*TCP4Conn, l)
	for i := 0; i < l; i++ {
		conn := TCP4Conn{
			LocalAddr:  convert.LEUint32ToBytes(rows[i].localAddr),
			RemoteAddr: convert.LEUint32ToBytes(rows[i].remoteAddr),
			State:      uint8(rows[i].state),
			PID:        int64(rows[i].pid),
		}
		conn.LocalPort = convert.BEBytesToUint16(rows[i].localPort[:2])
		conn.RemotePort = convert.BEBytesToUint16(rows[i].remotePort[:2])
		conns[i] = &conn
	}
	runtime.KeepAlive(table)
	return conns
}

type tcp4TableOwnerModule struct {
	n     uint32
	table [AnySize]tcp4RowOwnerModule
}

type tcp4RowOwnerModule struct {
	state      uint32
	localAddr  uint32
	localPort  [4]byte
	remoteAddr uint32
	remotePort [4]byte
	pid        uint32
	createTime FileTime
	moduleInfo [16]int64 // 16 is TCP IP_OWNING_MODULE_SIZE
}

func parseTCP4TableOwnerModule(buffer []byte) ([]*TCP4Conn, error) {
	// create process list map
	processes, err := GetProcessList()
	if err != nil {
		return nil, err
	}
	pm := make(map[uint32]string, len(processes))
	for i := 0; i < len(processes); i++ {
		pm[processes[i].PID] = processes[i].Name
	}
	table := (*tcp4TableOwnerModule)(unsafe.Pointer(&buffer[0]))
	var rows []tcp4RowOwnerModule
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rows))
	sh.Data = uintptr(unsafe.Pointer(&table.table))
	sh.Len = int(table.n)
	sh.Cap = int(table.n)
	l := len(rows)
	conns := make([]*TCP4Conn, l)
	for i := 0; i < l; i++ {
		conn := TCP4Conn{
			LocalAddr:  convert.LEUint32ToBytes(rows[i].localAddr),
			RemoteAddr: convert.LEUint32ToBytes(rows[i].remoteAddr),
			State:      uint8(rows[i].state),
			PID:        int64(rows[i].pid),
			CreateTime: rows[i].createTime.Time(),
			ModuleInfo: rows[i].moduleInfo,
			Process:    pm[rows[i].pid],
		}
		conn.LocalPort = convert.BEBytesToUint16(rows[i].localPort[:2])
		conn.RemotePort = convert.BEBytesToUint16(rows[i].remotePort[:2])
		conns[i] = &conn
	}
	runtime.KeepAlive(table)
	return conns, nil
}

// GetTCP6Conns is used to get TCP-over-IPv6 connections.
// can't use basic class.
func GetTCP6Conns(class uint32) ([]*TCP6Conn, error) {
	buffer, err := getTCPTable(windows.AF_INET6, class)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get tcp table")
	}
	var conns []*TCP6Conn
	switch {
	case class > 2 && class < 6:
		conns = parseTCP6TableOwnerPID(buffer)
	case class < 9:
		conns, err = parseTCP6TableOwnerModule(buffer)
		if err != nil {
			return nil, err
		}
	default:
		panic("api/network: unreachable code")
	}
	return conns, nil
}

func convertUint32sToIPv6(addr [4]uint32) net.IP {
	ip := make([]byte, 0, net.IPv6len)
	for i := 0; i < 4; i++ {
		ip = append(ip, convert.LEUint32ToBytes(addr[i])...)
	}
	return ip
}

type tcp6TableOwnerPID struct {
	n     uint32
	table [AnySize]tcp6RowOwnerPID
}

type tcp6RowOwnerPID struct {
	localAddr     [4]uint32
	localScopeID  uint32
	localPort     [4]byte
	remoteAddr    [4]uint32
	remoteScopeID uint32
	remotePort    [4]byte
	state         uint32
	pid           uint32
}

func parseTCP6TableOwnerPID(buffer []byte) []*TCP6Conn {
	table := (*tcp6TableOwnerPID)(unsafe.Pointer(&buffer[0]))
	var rows []tcp6RowOwnerPID
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rows))
	sh.Data = uintptr(unsafe.Pointer(&table.table))
	sh.Len = int(table.n)
	sh.Cap = int(table.n)
	l := len(rows)
	conns := make([]*TCP6Conn, l)
	for i := 0; i < l; i++ {
		conn := TCP6Conn{
			LocalAddr:     convertUint32sToIPv6(rows[i].localAddr),
			LocalScopeID:  rows[i].localScopeID,
			RemoteAddr:    convertUint32sToIPv6(rows[i].remoteAddr),
			RemoteScopeID: rows[i].remoteScopeID,
			State:         uint8(rows[i].state),
			PID:           int64(rows[i].pid),
		}
		conn.LocalPort = convert.BEBytesToUint16(rows[i].localPort[:2])
		conn.RemotePort = convert.BEBytesToUint16(rows[i].remotePort[:2])
		conns[i] = &conn
	}
	runtime.KeepAlive(table)
	return conns
}

type tcp6TableOwnerModule struct {
	n     uint32
	table [AnySize]tcp6RowOwnerModule
}

type tcp6RowOwnerModule struct {
	localAddr     [4]uint32
	localScopeID  uint32
	localPort     [4]byte
	remoteAddr    [4]uint32
	remoteScopeID uint32
	remotePort    [4]byte
	state         uint32
	pid           uint32
	createTime    FileTime
	moduleInfo    [16]int64 // 16 is TCP IP_OWNING_MODULE_SIZE
}

func parseTCP6TableOwnerModule(buffer []byte) ([]*TCP6Conn, error) {
	// create process list map
	processes, err := GetProcessList()
	if err != nil {
		return nil, err
	}
	pm := make(map[uint32]string, len(processes))
	for i := 0; i < len(processes); i++ {
		pm[processes[i].PID] = processes[i].Name
	}
	table := (*tcp6TableOwnerModule)(unsafe.Pointer(&buffer[0]))
	var rows []tcp6RowOwnerModule
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&rows))
	sh.Data = uintptr(unsafe.Pointer(&table.table))
	sh.Len = int(table.n)
	sh.Cap = int(table.n)
	l := len(rows)
	conns := make([]*TCP6Conn, l)
	for i := 0; i < l; i++ {
		conn := TCP6Conn{
			LocalAddr:     convertUint32sToIPv6(rows[i].localAddr),
			LocalScopeID:  rows[i].localScopeID,
			RemoteAddr:    convertUint32sToIPv6(rows[i].remoteAddr),
			RemoteScopeID: rows[i].remoteScopeID,
			State:         uint8(rows[i].state),
			PID:           int64(rows[i].pid),
			CreateTime:    rows[i].createTime.Time(),
			ModuleInfo:    rows[i].moduleInfo,
			Process:       pm[rows[i].pid],
		}
		conn.LocalPort = convert.BEBytesToUint16(rows[i].localPort[:2])
		conn.RemotePort = convert.BEBytesToUint16(rows[i].remotePort[:2])
		conns[i] = &conn
	}
	runtime.KeepAlive(table)
	return conns, nil
}

// UDP table class
const (
	UDPTableBasic uint32 = iota
	UDPTableOwnerPID
	UDPTableOwnerModule
)

// UDP over IPv4
type udp4Table struct {
	n     uint32
	table [AnySize]udp4Row
}

// UDP over IPv4 connection
type udp4Row struct {
	localAddr uint32
	localPort uint32
	pid       uint32
}

// UDP over IPv6
type udp6Table struct {
	n     uint32
	table [AnySize]udp6Row
}

// UDP over IPv6 connection
type udp6Row struct {
	localAddr    [4]uint32
	localScopeID uint32
	localPort    uint32
	pid          uint32
}
