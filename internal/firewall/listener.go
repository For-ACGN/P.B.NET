package firewall

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	// FilterModeDefault allow all connections, only use limit number.
	FilterModeDefault FilterMode = iota

	// FilterModeAllow only allow connection with host in allow list.
	FilterModeAllow

	// FilterModeBlock is block connection with host in block list.
	FilterModeBlock
)

const (
	defaultMaxConnsPerHost = 500
	defaultMaxConnsTotal   = 10000
)

// FilterMode is the firewall listener filter mode.
type FilterMode int

func (fm FilterMode) String() string {
	switch fm {
	case FilterModeDefault:
		return "default"
	case FilterModeAllow:
		return "allow"
	case FilterModeBlock:
		return "block"
	default:
		return fmt.Sprintf("invalid firewall listener filter mode: %d", fm)
	}
}

type maxConnsError struct {
	total bool
	host  string
}

func (e *maxConnsError) Error() string {
	if e.total {
		return "listener accepted too many connections"
	}
	return fmt.Sprintf("listener accepted too many connections about %s", e.host)
}

func (*maxConnsError) Timeout() bool { return false }

func (*maxConnsError) Temporary() bool { return true }

// Conn is the connection that Listener accepted.
type Conn struct {
	net.Conn
	fl   *Listener
	host string
}

// Close is used to close connection and delete connection in listener.
func (conn *Conn) Close() error {
	conn.fl.trackConn(conn, conn.host, false)
	return conn.Conn.Close()
}

// Listener is used to limit host and the number of connections of each host.
type Listener struct {
	listener      net.Listener
	filterMode    FilterMode
	onBlockedConn func(conn net.Conn)

	// key = host
	allowList    map[string]struct{}
	allowListRWM sync.RWMutex
	blockList    map[string]struct{}
	blockListRWM sync.RWMutex

	// store raw connections that can kill it, key = host
	conns    map[string]map[*Conn]struct{}
	connsRWM sync.RWMutex

	maxConnsPerHost atomic.Value
	maxConnsTotal   atomic.Value
}

// ListenerOptions contains options about Listener.
type ListenerOptions struct {
	FilterMode      FilterMode `json:"filter_mode"`
	MaxConnsPerHost int        `json:"max_conns_per_host"`
	MaxConnsTotal   int        `json:"max_conns_total"`

	// OnBlockedConn is used to let listener user handle blocked connection
	// with another handler, for example: allowed connection will reach
	// common server but blocked connection will reach the fake http server
	// that always return 404 page or 503 page ...(you can think it).
	OnBlockedConn func(conn net.Conn) `json:"-"`
}

// NewListener is used to create a new firewall listener.
func NewListener(listener net.Listener, opts *ListenerOptions) (*Listener, error) {
	if opts == nil {
		opts = new(ListenerOptions)
	}
	l := Listener{
		listener:      listener,
		filterMode:    opts.FilterMode,
		onBlockedConn: opts.OnBlockedConn,
	}
	switch opts.FilterMode {
	case FilterModeDefault:
	case FilterModeAllow:
		l.allowList = make(map[string]struct{}, 1)
	case FilterModeBlock:
		l.blockList = make(map[string]struct{}, 1)
	default:
		return nil, errors.New(opts.FilterMode.String())
	}
	if l.onBlockedConn == nil {
		l.onBlockedConn = func(conn net.Conn) {
			_ = conn.Close()
		}
	}
	l.conns = make(map[string]map[*Conn]struct{}, 1)
	maxConnsPerHost := opts.MaxConnsPerHost
	if maxConnsPerHost < 1 {
		maxConnsPerHost = defaultMaxConnsPerHost
	}
	maxConnsTotal := opts.MaxConnsTotal
	if maxConnsTotal < 1 {
		maxConnsTotal = defaultMaxConnsTotal
	}
	l.maxConnsPerHost.Store(maxConnsPerHost)
	l.maxConnsTotal.Store(maxConnsTotal)
	return &l, nil
}

// Accept is used to wait for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, host, err := l.accept()
		if err != nil {
			return nil, err
		}
		if conn == nil { // blocked conn
			continue
		}
		c := &Conn{
			fl:   l,
			host: host,
			Conn: conn,
		}
		l.trackConn(c, host, true)
		return c, nil
	}
}

func (l *Listener) accept() (net.Conn, string, error) {
	// check the number of total connections
	if l.GetConnsNumTotal() >= l.GetMaxConnsTotal() {
		return nil, "", &maxConnsError{total: true}
	}
	// accept connection
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, "", err
	}
	var ok bool
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, "", err
	}
	// check the number of connection about this host
	if l.GetConnsNumWithHost(host) >= l.GetMaxConnsPerHost() {
		return nil, "", &maxConnsError{host: host}
	}
	// check host is allowed
	switch l.filterMode {
	case FilterModeDefault:
		ok = true
		return conn, host, nil
	case FilterModeAllow:
		if l.isAllowedHost(host) {
			ok = true
			return conn, host, nil
		}
		l.onBlockedConn(conn)
	case FilterModeBlock:
		if !l.isBlockedHost(host) {
			ok = true
			return conn, host, nil
		}
		l.onBlockedConn(conn)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.filterMode))
	}
	return nil, "", nil
}

func (l *Listener) trackConn(conn *Conn, host string, add bool) {
	l.connsRWM.Lock()
	defer l.connsRWM.Unlock()
	if add {
		conns := l.conns[host]
		if conns == nil {
			conns = make(map[*Conn]struct{})
			l.conns[host] = conns
		}
		conns[conn] = struct{}{}
	} else {
		delete(l.conns[host], conn)
	}
}

// GetMaxConnsPerHost is used to get the maximum connection per-host.
func (l *Listener) GetMaxConnsPerHost() int {
	return l.maxConnsPerHost.Load().(int)
}

// GetMaxConnsTotal is used to get the maximum connection total.
func (l *Listener) GetMaxConnsTotal() int {
	return l.maxConnsTotal.Load().(int)
}

// SetMaxConnsPerHost is used to set the maximum connection per-host.
func (l *Listener) SetMaxConnsPerHost(v int) {
	if v < 1 {
		v = defaultMaxConnsPerHost
	}
	l.maxConnsPerHost.Store(v)
}

// SetMaxConnsTotal is used to set the maximum connection total.
func (l *Listener) SetMaxConnsTotal(v int) {
	if v < 1 {
		v = defaultMaxConnsTotal
	}
	l.maxConnsTotal.Store(v)
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

// GetConnsNumWithHost is used to get the number of the connections about host.
func (l *Listener) GetConnsNumWithHost(host string) int {
	l.connsRWM.RLock()
	defer l.connsRWM.RUnlock()
	return len(l.conns[host])
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

func (l *Listener) isAllowedHost(host string) bool {
	l.allowListRWM.RLock()
	defer l.allowListRWM.RUnlock()
	_, ok := l.allowList[host]
	return ok
}

func (l *Listener) isBlockedHost(host string) bool {
	l.blockListRWM.RLock()
	defer l.blockListRWM.RUnlock()
	_, ok := l.blockList[host]
	return ok
}

// IsAllowedHost is used to check this host is allowed.
func (l *Listener) IsAllowedHost(host string) bool {
	switch l.filterMode {
	case FilterModeDefault:
		return true
	case FilterModeAllow:
		return l.isAllowedHost(host)
	case FilterModeBlock:
		return !l.isBlockedHost(host)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.filterMode))
	}
}

// IsBlockedHost is used to check this host is blocked.
func (l *Listener) IsBlockedHost(host string) bool {
	switch l.filterMode {
	case FilterModeDefault:
		return false
	case FilterModeAllow:
		return !l.isAllowedHost(host)
	case FilterModeBlock:
		return l.isBlockedHost(host)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.filterMode))
	}
}

// AllowList is used to get allow host list.
func (l *Listener) AllowList() []string {
	if l.filterMode != FilterModeAllow {
		return nil
	}
	l.allowListRWM.RLock()
	defer l.allowListRWM.RUnlock()
	list := make([]string, 0, len(l.allowList))
	for host := range l.allowList {
		list = append(list, host)
	}
	return list
}

// BlockList is used to get block host list.
func (l *Listener) BlockList() []string {
	if l.filterMode != FilterModeBlock {
		return nil
	}
	l.blockListRWM.RLock()
	defer l.blockListRWM.RUnlock()
	list := make([]string, 0, len(l.blockList))
	for host := range l.blockList {
		list = append(list, host)
	}
	return list
}

// AddAllowedHost is used to add host to allow list.
func (l *Listener) AddAllowedHost(host string) {
	if l.filterMode != FilterModeAllow {
		return
	}
	l.allowListRWM.Lock()
	defer l.allowListRWM.Unlock()
	l.allowList[host] = struct{}{}
}

// AddBlockedHost is used to add host to block list.
func (l *Listener) AddBlockedHost(host string) {
	if l.filterMode != FilterModeBlock {
		return
	}
	l.blockListRWM.Lock()
	defer l.blockListRWM.Unlock()
	l.blockList[host] = struct{}{}
	// close all connections about this host
	l.connsRWM.RLock()
	defer l.connsRWM.RUnlock()
	for conn := range l.conns[host] {
		_ = conn.Conn.Close()
	}
}

// DeleteAllowedHost is used to delete host from allow list.
func (l *Listener) DeleteAllowedHost(host string) {
	if l.filterMode != FilterModeAllow {
		return
	}
	l.allowListRWM.Lock()
	defer l.allowListRWM.Unlock()
	delete(l.allowList, host)
}

// DeleteBlockedHost is used to delete host from block list.
func (l *Listener) DeleteBlockedHost(host string) {
	if l.filterMode != FilterModeBlock {
		return
	}
	l.blockListRWM.Lock()
	defer l.blockListRWM.Unlock()
	delete(l.blockList, host)
}

// FilterMode is used to get the firewall listener filter mode.
func (l *Listener) FilterMode() FilterMode {
	return l.filterMode
}

// Addr is used to get the listener's network host.
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close is used to close the listener.
func (l *Listener) Close() error {
	return l.listener.Close()
}
