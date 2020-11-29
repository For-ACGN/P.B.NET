package psmon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/compare"
	"project/internal/logger"
	"project/internal/module/process"
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
	EventProcessCreated
	EventProcessTerminated
)

// ErrMonitorClosed is an error that monitor is closed.
var ErrMonitorClosed = fmt.Errorf("monitor is closed")

// EventHandler is used to handle events, data type can be []*Process.
type EventHandler func(ctx context.Context, event uint8, data interface{})

// Options contains options about process monitor.
type Options struct {
	Interval time.Duration
	Process  *process.Options
}

// Monitor is used tp monitor process..
type Monitor struct {
	logger  logger.Logger
	handler EventHandler

	pauser *pauser.Pauser

	process  process.Process
	interval time.Duration
	closed   bool
	rwm      sync.RWMutex

	// for compare difference
	processes    []*process.PsInfo
	processesRWM sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New is used to create a process monitor.
func New(lg logger.Logger, handler EventHandler, opts *Options) (*Monitor, error) {
	if opts == nil {
		opts = new(Options)
	}
	interval := opts.Interval
	if interval < minimumRefreshInterval {
		interval = minimumRefreshInterval
	}
	monitor := Monitor{
		logger:   lg,
		interval: interval,
	}
	ps, err := process.New(opts.Process)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create process module")
	}
	var ok bool
	defer func() {
		if ok {
			return
		}
		err := ps.Close()
		if err != nil {
			monitor.log(logger.Error, "failed to close process module:", err)
		}
	}()
	monitor.process = ps
	// refresh before refreshLoop, and not set eventHandler.
	err = monitor.Refresh()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get the initial process data")
	}
	// not trigger eventHandler before first refresh.
	monitor.handler = handler
	// set context
	monitor.ctx, monitor.cancel = context.WithCancel(context.Background())
	// refreshLoop will block until call Start.
	monitor.pauser = pauser.New(monitor.ctx)
	monitor.pauser.Pause()
	monitor.wg.Add(1)
	go monitor.refreshLoop()
	ok = true
	return &monitor, nil
}

// GetInterval is used to get refresh interval.
func (mon *Monitor) GetInterval() time.Duration {
	mon.rwm.RLock()
	defer mon.rwm.RUnlock()
	return mon.interval
}

// SetInterval is used to set refresh interval, if set zero, it will pause auto refresh.
func (mon *Monitor) SetInterval(interval time.Duration) {
	if interval < minimumRefreshInterval {
		interval = minimumRefreshInterval
	}
	mon.rwm.Lock()
	defer mon.rwm.Unlock()
	mon.interval = interval
}

// SetOptions is used to update process module options.
func (mon *Monitor) SetOptions(opts *process.Options) error {
	mon.rwm.Lock()
	defer mon.rwm.Unlock()
	if mon.closed {
		return ErrMonitorClosed
	}
	ps, err := process.New(opts)
	if err != nil {
		return err
	}
	var ok bool
	defer func() {
		if ok {
			return
		}
		err = ps.Close()
		if err != nil {
			mon.log(logger.Error, "failed to close process module:", err)
		}
	}()
	err = mon.process.Close()
	if err != nil {
		return err
	}
	mon.process = ps
	ok = true
	return nil
}

func (mon *Monitor) log(lv logger.Level, log ...interface{}) {
	mon.logger.Println(lv, "process monitor", log...)
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
				if err != ErrMonitorClosed {
					mon.log(logger.Error, "failed to refresh:", err)
				}
				return
			}
		case <-mon.ctx.Done():
			return
		}
		timer.Reset(mon.GetInterval())
	}
}

// Refresh is used to refresh current system status at once.
func (mon *Monitor) Refresh() error {
	mon.rwm.RLock()
	defer mon.rwm.RUnlock()
	if mon.closed {
		return ErrMonitorClosed
	}
	processes, err := mon.process.List()
	if err != nil {
		return errors.WithMessage(err, "failed to get process list")
	}
	ds := &dataSource{
		processes: processes,
	}
	if mon.handler != nil {
		mon.compare(ds)
		return nil
	}
	mon.processesRWM.Lock()
	defer mon.processesRWM.Unlock()
	mon.refresh(ds)
	return nil
}

// for compare package
type processes []*process.PsInfo

func (ps processes) Len() int {
	return len(ps)
}

func (ps processes) ID(i int) string {
	return ps[i].ID()
}

type dataSource struct {
	processes []*process.PsInfo
}

type compareResult struct {
	createdProcesses  []*process.PsInfo
	terminatedProcess []*process.PsInfo
}

// compare is used to compare between stored in monitor.
func (mon *Monitor) compare(ds *dataSource) {
	var (
		createdProcesses  []*process.PsInfo
		terminatedProcess []*process.PsInfo
	)
	defer func() {
		mon.notice(&compareResult{
			createdProcesses:  createdProcesses,
			terminatedProcess: terminatedProcess,
		})
	}()
	mon.processesRWM.Lock()
	defer mon.processesRWM.Unlock()
	defer mon.refresh(ds)

	added, deleted := compare.UniqueSlice(processes(ds.processes), processes(mon.processes))
	for i := 0; i < len(added); i++ {
		createdProcesses = append(createdProcesses, ds.processes[added[i]].Clone())
	}
	for i := 0; i < len(deleted); i++ {
		terminatedProcess = append(terminatedProcess, mon.processes[deleted[i]].Clone())
	}
}

func (mon *Monitor) refresh(ds *dataSource) {
	mon.processes = ds.processes
}

func (mon *Monitor) notice(result *compareResult) {
	if len(result.createdProcesses) != 0 {
		mon.handler(mon.ctx, EventProcessCreated, result.createdProcesses)
	}
	if len(result.terminatedProcess) != 0 {
		mon.handler(mon.ctx, EventProcessTerminated, result.terminatedProcess)
	}
}

// GetProcesses is used to get processes that stored in monitor.
func (mon *Monitor) GetProcesses() []*process.PsInfo {
	mon.processesRWM.RLock()
	defer mon.processesRWM.RUnlock()
	l := len(mon.processes)
	processes := make([]*process.PsInfo, l)
	for i := 0; i < l; i++ {
		processes[i] = mon.processes[i].Clone()
	}
	return processes
}

// Start is used to start monitor.
func (mon *Monitor) Start() {
	mon.pauser.Continue()
}

// Pause is used to pause refresh automatically.
func (mon *Monitor) Pause() {
	mon.pauser.Pause()
}

// Continue is used to continue refresh automatically.
func (mon *Monitor) Continue() {
	mon.pauser.Continue()
}

// Close is used to close process monitor.
func (mon *Monitor) Close() error {
	mon.cancel()
	mon.wg.Wait()
	mon.rwm.Lock()
	defer mon.rwm.Unlock()
	err := mon.process.Close()
	if err != nil {
		return err
	}
	mon.closed = true
	return nil
}
