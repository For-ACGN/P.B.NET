package netmon

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/compare"
	"project/internal/logger"
	"project/internal/module/netstat"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

const (
	defaultRefreshInterval = 500 * time.Millisecond
	minimumRefreshInterval = 100 * time.Millisecond
)

// events about monitor.
const (
	_ uint8 = iota
	EventConnCreated
	EventConnClosed
)

// EventHandler is used to handle appeared event, data type can be []interface{} that include
// []*netstat.TCP4Conn, []*netstat.TCP6Conn, []*netstat.UDP4Conn, []*netstat.UDP6Conn
type EventHandler func(ctx context.Context, event uint8, data interface{})

// Options contains options about network monitor.
type Options struct {
	Interval time.Duration
	Netstat  *netstat.Options
}

// Monitor is used tp monitor network status about current system.
type Monitor struct {
	logger  logger.Logger
	handler EventHandler

	pauser      *pauser.Pauser
	netstat     netstat.Netstat
	interval    time.Duration
	intervalRWM sync.RWMutex

	// about check network status
	tcp4Conns []*netstat.TCP4Conn
	tcp6Conns []*netstat.TCP6Conn
	udp4Conns []*netstat.UDP4Conn
	udp6Conns []*netstat.UDP6Conn
	connsRWM  sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New is used to create a network status monitor.
func New(logger logger.Logger, handler EventHandler, opts *Options) (*Monitor, error) {
	if opts == nil {
		opts = new(Options)
	}
	interval := opts.Interval
	if interval < defaultRefreshInterval {
		interval = defaultRefreshInterval
	}
	ns, err := netstat.New(opts.Netstat)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create netstat module")
	}
	ctx, cancel := context.WithCancel(context.Background())
	monitor := Monitor{
		logger:   logger,
		pauser:   pauser.New(ctx),
		netstat:  ns,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
	// refresh before refreshLoop, and not set eventHandler.
	err = monitor.Refresh()
	if err != nil {
		_ = ns.Close()
		return nil, err
	}
	// not trigger eventHandler before first refresh.
	monitor.handler = handler
	monitor.wg.Add(1)
	go monitor.refreshLoop()
	return &monitor, nil
}

// GetInterval is used to get refresh interval.
func (mon *Monitor) GetInterval() time.Duration {
	mon.intervalRWM.RLock()
	defer mon.intervalRWM.RUnlock()
	return mon.interval
}

// SetInterval is used to set refresh interval, if set zero, it will pause auto refresh.
func (mon *Monitor) SetInterval(interval time.Duration) {
	if interval < minimumRefreshInterval {
		interval = minimumRefreshInterval
	}
	mon.intervalRWM.Lock()
	defer mon.intervalRWM.Unlock()
	mon.interval = interval
}

func (mon *Monitor) log(lv logger.Level, log ...interface{}) {
	mon.logger.Println(lv, "network monitor", log...)
}

func (mon *Monitor) refreshLoop() {
	defer mon.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			mon.log(logger.Fatal, xpanic.Print(r, "Monitor.refreshLoop"))
			// restart
			time.Sleep(time.Second)
			mon.wg.Add(1)
			go mon.refreshLoop()
		}
	}()
	timer := time.NewTimer(mon.GetInterval())
	defer timer.Stop()
	for {
		mon.pauser.Paused()
		select {
		case <-timer.C:
			err := mon.Refresh()
			if err != nil {
				mon.log(logger.Error, "failed to refresh:", err)
				return
			}
		case <-mon.ctx.Done():
			return
		}
		timer.Reset(mon.GetInterval())
	}
}

// Refresh is used to refresh current network status at once.
func (mon *Monitor) Refresh() error {
	tcp4Conns, err := mon.netstat.GetTCP4Conns()
	if err != nil {
		return err
	}
	tcp6Conns, err := mon.netstat.GetTCP6Conns()
	if err != nil {
		return err
	}
	udp4Conns, err := mon.netstat.GetUDP4Conns()
	if err != nil {
		return err
	}
	udp6Conns, err := mon.netstat.GetUDP6Conns()
	if err != nil {
		return err
	}
	ds := &dataSource{
		tcp4Conns: tcp4Conns,
		tcp6Conns: tcp6Conns,
		udp4Conns: udp4Conns,
		udp6Conns: udp6Conns,
	}
	if mon.handler != nil {
		result := mon.compare(ds)
		mon.refresh(ds)
		mon.notice(result)
		return nil
	}
	mon.refresh(ds)
	return nil
}

// for compare package

type tcp4Conns []*netstat.TCP4Conn

func (conns tcp4Conns) Len() int {
	return len(conns)
}

func (conns tcp4Conns) ID(i int) string {
	return conns[i].ID()
}

type tcp6Conns []*netstat.TCP6Conn

func (conns tcp6Conns) Len() int {
	return len(conns)
}

func (conns tcp6Conns) ID(i int) string {
	return conns[i].ID()
}

type udp4Conns []*netstat.UDP4Conn

func (conns udp4Conns) Len() int {
	return len(conns)
}

func (conns udp4Conns) ID(i int) string {
	return conns[i].ID()
}

type udp6Conns []*netstat.UDP6Conn

func (conns udp6Conns) Len() int {
	return len(conns)
}

func (conns udp6Conns) ID(i int) string {
	return conns[i].ID()
}

type dataSource struct {
	tcp4Conns []*netstat.TCP4Conn
	tcp6Conns []*netstat.TCP6Conn
	udp4Conns []*netstat.UDP4Conn
	udp6Conns []*netstat.UDP6Conn
}

type compareResult struct {
	createdConns []interface{}
	closedConns  []interface{}
}

// compare is used to compare between stored in monitor.
func (mon *Monitor) compare(ds *dataSource) *compareResult {
	var (
		createdConns []interface{}
		closedConns  []interface{}
	)
	mon.connsRWM.RLock()
	defer mon.connsRWM.RUnlock()
	// TCP4
	added, deleted := compare.UniqueSlice(tcp4Conns(ds.tcp4Conns), tcp4Conns(mon.tcp4Conns))
	for i := 0; i < len(added); i++ {
		createdConns = append(createdConns, ds.tcp4Conns[added[i]])
	}
	for i := 0; i < len(deleted); i++ {
		closedConns = append(closedConns, mon.tcp4Conns[deleted[i]])
	}
	// TCP6
	added, deleted = compare.UniqueSlice(tcp6Conns(ds.tcp6Conns), tcp6Conns(mon.tcp6Conns))
	for i := 0; i < len(added); i++ {
		createdConns = append(createdConns, ds.tcp6Conns[added[i]])
	}
	for i := 0; i < len(deleted); i++ {
		closedConns = append(closedConns, mon.tcp6Conns[deleted[i]])
	}
	// UDP4
	added, deleted = compare.UniqueSlice(udp4Conns(ds.udp4Conns), udp4Conns(mon.udp4Conns))
	for i := 0; i < len(added); i++ {
		createdConns = append(createdConns, ds.udp4Conns[added[i]])
	}
	for i := 0; i < len(deleted); i++ {
		closedConns = append(closedConns, mon.udp4Conns[deleted[i]])
	}
	// UDP6
	added, deleted = compare.UniqueSlice(udp6Conns(ds.udp6Conns), udp6Conns(mon.udp6Conns))
	for i := 0; i < len(added); i++ {
		createdConns = append(createdConns, ds.udp6Conns[added[i]])
	}
	for i := 0; i < len(deleted); i++ {
		closedConns = append(closedConns, mon.udp6Conns[deleted[i]])
	}
	return &compareResult{
		createdConns: createdConns,
		closedConns:  closedConns,
	}
}

func (mon *Monitor) refresh(ds *dataSource) {
	mon.connsRWM.Lock()
	defer mon.connsRWM.Unlock()
	mon.tcp4Conns = ds.tcp4Conns
	mon.tcp6Conns = ds.tcp6Conns
	mon.udp4Conns = ds.udp4Conns
	mon.udp6Conns = ds.udp6Conns
}

func (mon *Monitor) notice(result *compareResult) {
	if len(result.createdConns) != 0 {
		mon.handler(mon.ctx, EventConnCreated, result.createdConns)
	}
	if len(result.closedConns) != 0 {
		mon.handler(mon.ctx, EventConnClosed, result.closedConns)
	}
}

// GetTCP4Conns is used to get tcp4 connections that stored in monitor.
func (mon *Monitor) GetTCP4Conns() []*netstat.TCP4Conn {
	mon.connsRWM.RLock()
	defer mon.connsRWM.RUnlock()
	return mon.tcp4Conns
}

// GetTCP6Conns is used to get tcp6 connections that stored in monitor.
func (mon *Monitor) GetTCP6Conns() []*netstat.TCP6Conn {
	mon.connsRWM.RLock()
	defer mon.connsRWM.RUnlock()
	return mon.tcp6Conns
}

// GetUDP4Conns is used to get udp4 connections that stored in monitor.
func (mon *Monitor) GetUDP4Conns() []*netstat.UDP4Conn {
	mon.connsRWM.RLock()
	defer mon.connsRWM.RUnlock()
	return mon.udp4Conns
}

// GetUDP6Conns is used to get udp6 connections that stored in monitor.
func (mon *Monitor) GetUDP6Conns() []*netstat.UDP6Conn {
	mon.connsRWM.RLock()
	defer mon.connsRWM.RUnlock()
	return mon.udp6Conns
}

// Pause is used to pause auto refresh.
func (mon *Monitor) Pause() {
	mon.pauser.Pause()
}

// Continue is used to continue auto refresh.
func (mon *Monitor) Continue() {
	mon.pauser.Continue()
}

// Close is used to close network status monitor.
func (mon *Monitor) Close() error {
	mon.cancel()
	mon.wg.Wait()
	return mon.netstat.Close()
}
