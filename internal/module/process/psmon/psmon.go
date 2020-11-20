package psmon

import (
	"context"
	"sync"
	"time"

	"project/internal/compare"
	"project/internal/logger"
	"project/internal/task/pauser"
	"project/internal/xpanic"
)

const (
	defaultRefreshInterval = 500 * time.Millisecond
	minimumRefreshInterval = 100 * time.Millisecond
)

// about events
const (
	_ uint8 = iota
	EventProcessCreated
	EventProcessTerminated
)

// EventHandler is used to handle appeared event.
// data type can be []*Process.
type EventHandler func(ctx context.Context, event uint8, data interface{})

// Monitor is used tp monitor system status about current system.
type Monitor struct {
	logger  logger.Logger
	handler EventHandler

	pauser   *pauser.Pauser
	tasklist TaskList
	interval time.Duration
	rwm      sync.RWMutex

	// about check system status
	processes []*Process
	statusRWM sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMonitor is used to create a system status monitor.
func NewMonitor(logger logger.Logger, handler EventHandler, opts *Options) (*Monitor, error) {
	tasklist, err := NewTaskList(opts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	monitor := Monitor{
		logger:   logger,
		pauser:   pauser.New(ctx),
		tasklist: tasklist,
		interval: defaultRefreshInterval,
		ctx:      ctx,
		cancel:   cancel,
	}
	// refresh before refreshLoop, and not set eventHandler.
	err = monitor.Refresh()
	if err != nil {
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

func (mon *Monitor) log(lv logger.Level, log ...interface{}) {
	mon.logger.Println(lv, "system monitor", log...)
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

// Refresh is used to refresh current system status at once.
func (mon *Monitor) Refresh() error {
	processes, err := mon.tasklist.GetProcesses()
	if err != nil {
		return err
	}
	ds := &dataSource{
		processes: processes,
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

type dataSource struct {
	processes []*Process
}

type compareResult struct {
	createdProcesses  []*Process
	terminatedProcess []*Process
}

// compare is used to compare between stored in monitor.
func (mon *Monitor) compare(ds *dataSource) *compareResult {
	var (
		createdProcesses  []*Process
		terminatedProcess []*Process
	)
	mon.statusRWM.RLock()
	defer mon.statusRWM.RUnlock()
	added, deleted := compare.UniqueSlice(processes(ds.processes), processes(mon.processes))
	for i := 0; i < len(added); i++ {
		createdProcesses = append(createdProcesses, ds.processes[added[i]])
	}
	for i := 0; i < len(deleted); i++ {
		terminatedProcess = append(terminatedProcess, mon.processes[deleted[i]])
	}
	return &compareResult{
		createdProcesses:  createdProcesses,
		terminatedProcess: terminatedProcess,
	}
}

func (mon *Monitor) refresh(ds *dataSource) {
	mon.statusRWM.Lock()
	defer mon.statusRWM.Unlock()
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
func (mon *Monitor) GetProcesses() []*Process {
	mon.statusRWM.RLock()
	defer mon.statusRWM.RUnlock()
	return mon.processes
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
	return mon.tasklist.Close()
}
