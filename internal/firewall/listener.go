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

// Listener is used to limit IP address and the number of connections of each IP address.
type Listener struct {
	listener    net.Listener
	mode        ListenerMode
	onBlockConn func(conn net.Conn)

	// key = remote address
	allowList    map[string]struct{}
	allowListRWM sync.RWMutex
	blockList    map[string]struct{}
	blockListRWM sync.RWMutex

	// key = remote address
	conns    map[string]map[*net.Conn]struct{}
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

// NewListener is used to create a new limit listener.
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
	l.conns = make(map[string]map[*net.Conn]struct{}, 1)
	maxConnPerAddr := opts.MaxConnPerAddr
	if maxConnPerAddr < 1 {
		maxConnPerAddr = 500
	}
	maxConnTotal := opts.MaxConnTotal
	if maxConnTotal < 1 {
		maxConnTotal = 10000
	}
	l.maxConnPerAddr.Store(maxConnPerAddr)
	l.maxConnTotal.Store(maxConnTotal)
	return &l, nil
}

// Accept is used to wait for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, err := l.accept()
		if err != nil {
			return nil, err
		}
		if conn != nil {
			return conn, nil
		}
	}
}

func (l *Listener) accept() (net.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	// check the number of connections

	// check remote address is allowed
	switch l.mode {
	case ListenerModeDefault:
		return conn, nil
	case ListenerModeAllow:
		if l.isAllowed(conn.RemoteAddr().String()) {
			return conn, nil
		}
		l.onBlockConn(conn)
	case ListenerModeBlock:
		if !l.isBlocked(conn.RemoteAddr().String()) {
			return conn, nil
		}
		l.onBlockConn(conn)
	default:
		panic(fmt.Sprintf("firewall listener internal error: %s", l.mode))
	}
	return nil, nil
}

func (l *Listener) isAllowed(addr string) bool {
	l.allowListRWM.RLock()
	defer l.allowListRWM.RUnlock()
	_, ok := l.allowList[addr]
	return ok
}

func (l *Listener) isBlocked(addr string) bool {
	l.blockListRWM.RLock()
	defer l.blockListRWM.RUnlock()
	_, ok := l.blockList[addr]
	return ok
}
