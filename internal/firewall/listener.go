package firewall

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	// ListenerModeDefault allow all connections, only use limit connections.
	ListenerModeDefault ListenerMode = iota

	// ListenerModeAllow only allow connection with remote address in allow list.
	ListenerModeAllow

	// ListenerModeBlock is block connection with remote address in block list.
	ListenerModeBlock
)

const (
	defaultMaxConnPerAddr = 500
	defaultMaxConnTotal   = 10000
)

// ListenerMode is the firewall listener mode.
type ListenerMode int

func (l ListenerMode) String() string {
	switch l {
	case ListenerModeDefault:
		return "default"
	case ListenerModeAllow:
		return "allow"
	case ListenerModeBlock:
		return "block"
	default:
		return fmt.Sprintf("invalid firewall listener mode: %d", l)
	}
}

type maxConnError struct {
	total bool
	addr  string
}

func (m *maxConnError) Error() string {
	if m.total {
		return "listener accepted too many connections"
	}
	return fmt.Sprintf("listener accepted too many connections about %s", m.addr)
}

func (m *maxConnError) Timeout() bool {
	return false
}

func (m *maxConnError) Temporary() bool {
	return true
}

// Conn is the connection that Listener accepted.
type Conn struct {
	l *Listener
	net.Conn
}

func (conn *Conn) Close() error {

	return conn.Conn.Close()
}

// Listener is used to limit address and the number of connections of each address.
type Listener struct {
	listener    net.Listener
	mode        ListenerMode
	onBlockConn func(conn net.Conn)

	// key = remote address
	allowList    map[string]struct{}
	allowListRWM sync.RWMutex
	blockList    map[string]struct{}
	blockListRWM sync.RWMutex

	// store raw connections that can kill it ,key = remote address
	conns    map[string]map[*Conn]struct{}
	connsRWM sync.RWMutex

	maxConnPerAddr atomic.Value
	maxConnTotal   atomic.Value
}

// ListenerOptions contains options about Listener.
type ListenerOptions struct {
	Mode           int `json:"mode"`
	MaxConnPerAddr int `json:"max_conn_per_addr"`
	MaxConnTotal   int `json:"max_conn_total"`

	// OnBlockConn is used to let listener user handle blocked connection
	// with another handler, for example: allowed connection will reach
	// common server but blocked connection will reach the fake http server
	// that always return 404 page or 503 page ...(you can think it).
	OnBlockConn func(conn net.Conn) `json:"-"`
}

// NewListener is used to create a new firewall listener.
func NewListener(listener net.Listener, opts *ListenerOptions) (*Listener, error) {
	if opts == nil {
		opts = new(ListenerOptions)
	}
	lm := ListenerMode(opts.Mode)
	l := Listener{
		listener:    listener,
		mode:        lm,
		onBlockConn: opts.OnBlockConn,
	}
	if l.onBlockConn == nil {
		l.onBlockConn = func(conn net.Conn) {
			_ = conn.Close()
		}
	}
	switch lm {
	case ListenerModeDefault:
	case ListenerModeAllow:
		l.allowList = make(map[string]struct{}, 1)
	case ListenerModeBlock:
		l.blockList = make(map[string]struct{}, 1)
	default:
		return nil, errors.New(lm.String())
	}
	l.conns = make(map[string]map[*Conn]struct{}, 1)
	maxConnPerAddr := opts.MaxConnPerAddr
	if maxConnPerAddr < 1 {
		maxConnPerAddr = defaultMaxConnPerAddr
	}
	maxConnTotal := opts.MaxConnTotal
	if maxConnTotal < 1 {
		maxConnTotal = defaultMaxConnTotal
	}
	l.maxConnPerAddr.Store(maxConnPerAddr)
	l.maxConnTotal.Store(maxConnTotal)
	return &l, nil
}

// Accept is used to wait for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, addr, err := l.accept()
		if err != nil {
			return nil, err
		}
		if conn == nil {
			continue
		}
		c := &Conn{
			l:    l,
			Conn: conn,
		}
		l.trackConn(c, addr, true)
		return c, nil
	}
}

func (l *Listener) accept() (net.Conn, string, error) {
	// check the number of total connections
	if l.GetConnsNumTotal() >= l.GetMaxConnTotal() {
		return nil, "", errors.WithStack(&maxConnError{total: true})
	}
	// accept connection
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, "", err
	}
	addr := conn.RemoteAddr().String()
	// check the number of connection about this address
	if l.GetConnsNumWithAddr(addr) >= l.GetMaxConnPerAddr() {
		_ = conn.Close()
		return nil, "", errors.WithStack(&maxConnError{addr: addr})
	}
	// check remote address is allowed
	switch l.mode {
	case ListenerModeDefault:
		return conn, addr, nil
	case ListenerModeAllow:
		if l.isAllowedAddr(addr) {
			return conn, addr, nil
		}
		l.onBlockConn(conn)
	case ListenerModeBlock:
		if !l.isBlockedAddr(addr) {
			return conn, addr, nil
		}
		l.onBlockConn(conn)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.mode))
	}
	return nil, "", nil
}

func (l *Listener) trackConn(conn *Conn, addr string, add bool) {
	l.connsRWM.Lock()
	defer l.connsRWM.Unlock()
	if add {
		conns := l.conns[addr]
		if conns == nil {
			conns = make(map[*Conn]struct{})
			l.conns[addr] = conns
		}
		conns[conn] = struct{}{}
	} else {
		delete(l.conns[addr], conn)
	}
}

// GetMaxConnPerAddr is used to get the maximum connection per-address.
func (l *Listener) GetMaxConnPerAddr() int {
	return l.maxConnPerAddr.Load().(int)
}

// GetMaxConnTotal is used to get the maximum connection total.
func (l *Listener) GetMaxConnTotal() int {
	return l.maxConnTotal.Load().(int)
}

// SetMaxConnPerAddr is used to set the maximum connection per-address.
func (l *Listener) SetMaxConnPerAddr(v int) {
	if v < 1 {
		v = defaultMaxConnPerAddr
	}
	l.maxConnPerAddr.Store(v)
}

// SetMaxConnTotal is used to set the maximum connection total.
func (l *Listener) SetMaxConnTotal(v int) {
	if v < 1 {
		v = defaultMaxConnTotal
	}
	l.maxConnTotal.Store(v)
}

// GetConnsNumTotal is used to get the number of the connections.
func (l *Listener) GetConnsNumTotal() int {
	l.connsRWM.RLock()
	defer l.connsRWM.RUnlock()
	var num int
	for _, conns := range l.conns {
		num += len(conns)
	}
	return num
}

// GetConnsNumWithAddr is used to get the number of the connections about address.
func (l *Listener) GetConnsNumWithAddr(addr string) int {
	l.connsRWM.RLock()
	defer l.connsRWM.RUnlock()
	return len(l.conns[addr])
}

// GetConns is used to get the all connections.
func (l *Listener) GetConns() []*Conn {
	var cs []*Conn
	l.connsRWM.RLock()
	defer l.connsRWM.RUnlock()
	for _, conns := range l.conns {
		for conn := range conns {
			cs = append(cs, conn)
		}
	}
	return cs
}

func (l *Listener) isAllowedAddr(addr string) bool {
	l.allowListRWM.RLock()
	defer l.allowListRWM.RUnlock()
	_, ok := l.allowList[addr]
	return ok
}

func (l *Listener) isBlockedAddr(addr string) bool {
	l.blockListRWM.RLock()
	defer l.blockListRWM.RUnlock()
	_, ok := l.blockList[addr]
	return ok
}

// IsAllowedAddr is used to check this address is allowed.
func (l *Listener) IsAllowedAddr(addr string) bool {
	switch l.mode {
	case ListenerModeDefault:
		return true
	case ListenerModeAllow:
		return l.isAllowedAddr(addr)
	case ListenerModeBlock:
		return !l.isBlockedAddr(addr)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.mode))
	}
}

// IsBlockedAddr is used to check this address is blocked.
func (l *Listener) IsBlockedAddr(addr string) bool {
	switch l.mode {
	case ListenerModeDefault:
		return false
	case ListenerModeAllow:
		return !l.isAllowedAddr(addr)
	case ListenerModeBlock:
		return l.isBlockedAddr(addr)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.mode))
	}
}

// AllowList is used to get allow address list.
func (l *Listener) AllowList() []string {
	if l.mode != ListenerModeAllow {
		return nil
	}
	l.allowListRWM.RLock()
	defer l.allowListRWM.RUnlock()
	list := make([]string, 0, len(l.allowList))
	for addr := range l.allowList {
		list = append(list, addr)
	}
	return list
}

// BlockList is used to get block address list.
func (l *Listener) BlockList() []string {
	if l.mode != ListenerModeBlock {
		return nil
	}
	l.blockListRWM.RLock()
	defer l.blockListRWM.RUnlock()
	list := make([]string, 0, len(l.blockList))
	for addr := range l.blockList {
		list = append(list, addr)
	}
	return list
}

// AddAllowedAddress is used to add address to allow list.
func (l *Listener) AddAllowedAddress(addr string) {
	if l.mode != ListenerModeAllow {
		return
	}
	l.allowListRWM.Lock()
	defer l.allowListRWM.Unlock()
	l.allowList[addr] = struct{}{}
}

// AddBlockedAddr is used to add address to block list.
func (l *Listener) AddBlockedAddr(addr string) {
	if l.mode != ListenerModeBlock {
		return
	}
	l.blockListRWM.Lock()
	defer l.blockListRWM.Unlock()
	l.blockList[addr] = struct{}{}
}

// DeleteAllowedAddr is used to delete address from allow list.
func (l *Listener) DeleteAllowedAddr(addr string) {
	if l.mode != ListenerModeAllow {
		return
	}
	l.allowListRWM.Lock()
	defer l.allowListRWM.Unlock()
	delete(l.allowList, addr)
}

// DeleteBlockedAddr is used to delete address from block list.
func (l *Listener) DeleteBlockedAddr(addr string) {
	if l.mode != ListenerModeBlock {
		return
	}
	l.blockListRWM.Lock()
	defer l.blockListRWM.Unlock()
	delete(l.blockList, addr)
}

// Addr is used to get the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close is used to close the listener.
func (l *Listener) Close() error {
	return l.listener.Close()
}
