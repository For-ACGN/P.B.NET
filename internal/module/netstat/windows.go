// +build windows

package netstat

import (
	"project/internal/module/windows/api"
)

// Options contain options about table class.
type Options struct {
	TCPTableClass uint32
	UDPTableClass uint32
}

type netstat struct {
	tcpTableClass uint32
	udpTableClass uint32
}

// NewNetstat is used to create a netstat with TCP and UDP table class.
func NewNetstat(opts *Options) (Netstat, error) {
	if opts == nil {
		opts = &Options{
			TCPTableClass: api.TCPTableOwnerModuleAll,
			UDPTableClass: api.UDPTableOwnerModule,
		}
	}
	return &netstat{
		tcpTableClass: opts.TCPTableClass,
		udpTableClass: opts.UDPTableClass,
	}, nil
}

func (n *netstat) GetTCP4Conns() ([]*TCP4Conn, error) {
	conns, err := api.GetTCP4Conns(n.tcpTableClass)
	if err != nil {
		return nil, err
	}
	l := len(conns)
	cs := make([]*TCP4Conn, l)
	for i := 0; i < l; i++ {
		cs[i] = &TCP4Conn{
			LocalAddr:  conns[i].LocalAddr,
			LocalPort:  conns[i].LocalPort,
			RemoteAddr: conns[i].RemoteAddr,
			RemotePort: conns[i].RemotePort,
			State:      conns[i].State,
			PID:        conns[i].PID,
			Process:    conns[i].Process,
		}
	}
	return cs, nil
}

func (n *netstat) GetTCP6Conns() ([]*TCP6Conn, error) {
	conns, err := api.GetTCP6Conns(n.tcpTableClass)
	if err != nil {
		return nil, err
	}
	l := len(conns)
	cs := make([]*TCP6Conn, l)
	for i := 0; i < l; i++ {
		cs[i] = &TCP6Conn{
			LocalAddr:     conns[i].LocalAddr,
			LocalScopeID:  conns[i].LocalScopeID,
			LocalPort:     conns[i].LocalPort,
			RemoteAddr:    conns[i].RemoteAddr,
			RemoteScopeID: conns[i].RemoteScopeID,
			RemotePort:    conns[i].RemotePort,
			State:         conns[i].State,
			PID:           conns[i].PID,
			Process:       conns[i].Process,
		}
	}
	return cs, nil
}

func (n *netstat) GetUDP4Conns() ([]*UDP4Conn, error) {
	conns, err := api.GetUDP4Conns(n.udpTableClass)
	if err != nil {
		return nil, err
	}
	l := len(conns)
	cs := make([]*UDP4Conn, l)
	for i := 0; i < l; i++ {
		cs[i] = &UDP4Conn{
			LocalAddr: conns[i].LocalAddr,
			LocalPort: conns[i].LocalPort,
			PID:       conns[i].PID,
			Process:   conns[i].Process,
		}
	}
	return cs, nil
}

func (n *netstat) GetUDP6Conns() ([]*UDP6Conn, error) {
	conns, err := api.GetUDP6Conns(n.udpTableClass)
	if err != nil {
		return nil, err
	}
	l := len(conns)
	cs := make([]*UDP6Conn, l)
	for i := 0; i < l; i++ {
		cs[i] = &UDP6Conn{
			LocalAddr:    conns[i].LocalAddr,
			LocalScopeID: conns[i].LocalScopeID,
			LocalPort:    conns[i].LocalPort,
			PID:          conns[i].PID,
			Process:      conns[i].Process,
		}
	}
	return cs, nil
}

func (n *netstat) Close() error {
	return nil
}

// GetTCPConnState is used to convert state to string.
func GetTCPConnState(state uint8) string {
	return api.GetTCPConnState(state)
}
