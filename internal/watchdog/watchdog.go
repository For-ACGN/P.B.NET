package watchdog

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"project/internal/logger"
	"project/internal/xpanic"
)

const defaultPeriod = 10 * time.Second

// Callback is used to notice when watcher is blocked.
type Callback func(ctx context.Context, id int32)

// Watcher is a watcher that spawned by WatchDog.
type Watcher struct {
	ctx    *WatchDog
	id     int32
	Signal <-chan struct{}
}

// Receive is used to receive signal first.
func (watcher *Watcher) Receive() {
	select {
	case <-watcher.Signal:
	default:
	}
}

// Stop is used to stop this watcher.
func (watcher *Watcher) Stop() {
	watcher.ctx.deleteWatcher(watcher.id)
	watcher.ctx.deleteBlockedID(watcher.id)
	watcher.ctx = nil
}

// WatchDog is used to watch a loop in worker goroutine is blocked.
type WatchDog struct {
	logger  logger.Logger
	onBlock Callback

	logSrc string
	nextID *int32
	period atomic.Value

	watchers    map[int32]chan struct{}
	watchersRWM sync.RWMutex

	// not notice blocked watcher twice
	blockedID    map[int32]struct{}
	blockedIDRWM sync.RWMutex

	// about control
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// New is used to create a new watch dog.
func New(logger logger.Logger, tag string, onBlock Callback) *WatchDog {
	logSrc := "watchdog"
	if tag != "" {
		logSrc += "-" + tag
	}
	wd := WatchDog{
		logger:    logger,
		onBlock:   onBlock,
		logSrc:    logSrc,
		nextID:    new(int32),
		watchers:  make(map[int32]chan struct{}),
		blockedID: make(map[int32]struct{}),
	}
	*wd.nextID = -1
	wd.period.Store(defaultPeriod)
	wd.ctx, wd.cancel = context.WithCancel(context.Background())
	return &wd
}

// NewWatcher is used to create a new watcher.
func (wd *WatchDog) NewWatcher() *Watcher {
	id := atomic.AddInt32(wd.nextID, 1)
	watcher := make(chan struct{}, 1)
	wd.watchersRWM.Lock()
	defer wd.watchersRWM.Unlock()
	wd.watchers[id] = watcher
	return &Watcher{
		ctx:    wd,
		id:     id,
		Signal: watcher,
	}
}

// Start is used to start watch dog.
func (wd *WatchDog) Start() {
	wd.wg.Add(1)
	go wd.kickLoop()
}

// GetPeriod is used to get watch dog period.
func (wd *WatchDog) GetPeriod() time.Duration {
	return wd.period.Load().(time.Duration)
}

// SetPeriod is used to set watch dog period.
func (wd *WatchDog) SetPeriod(period time.Duration) {
	if period < 10*time.Millisecond || period > 3*time.Minute {
		period = defaultPeriod
	}
	wd.period.Store(period)
}

func (wd *WatchDog) logf(lv logger.Level, format string, log ...interface{}) {
	wd.logger.Printf(lv, wd.logSrc, format, log...)
}

func (wd *WatchDog) log(lv logger.Level, log ...interface{}) {
	wd.logger.Println(lv, wd.logSrc, log...)
}

func (wd *WatchDog) kickLoop() {
	defer wd.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			wd.log(logger.Fatal, xpanic.Print(r, "WatchDog.kickLoop"))
			// restart
			time.Sleep(time.Second)
			wd.wg.Add(1)
			go wd.kickLoop()
		}
	}()
	period := wd.GetPeriod()
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			wd.kick()
		case <-wd.ctx.Done():
			return
		}
		// update kick period
		p := wd.GetPeriod()
		if p != period {
			period = p
		}
		ticker.Reset(period)
	}
}

func (wd *WatchDog) kick() {
	for id, watcher := range wd.getWatchers() {
		// kicking
		select {
		case watcher <- struct{}{}:
			if wd.isBlocked(id) {
				wd.deleteBlockedID(id)
				wd.logf(logger.Info, "watcher [%d] is running", id)
			}
			continue
		default:
		}
		// check it is already blocked
		if wd.isBlocked(id) {
			return
		}
		wd.addBlockedID(id)
		// start a new goroutine to call onBlock
		wd.wg.Add(1)
		go func(id int32) {
			defer wd.wg.Done()
			if r := recover(); r != nil {
				wd.log(logger.Fatal, xpanic.Printf(r, "WatchDog.onBlock"))
			}
			wd.logf(logger.Warning, "watcher [%d] is blocked", id)
			if wd.onBlock == nil {
				return
			}
			wd.onBlock(wd.ctx, id)
		}(id)
	}
}

func (wd *WatchDog) getWatchers() map[int32]chan struct{} {
	wd.watchersRWM.RLock()
	defer wd.watchersRWM.RUnlock()
	watchers := make(map[int32]chan struct{}, len(wd.watchers))
	for id, watcher := range wd.watchers {
		watchers[id] = watcher
	}
	return watchers
}

func (wd *WatchDog) deleteWatcher(id int32) {
	wd.watchersRWM.Lock()
	defer wd.watchersRWM.Unlock()
	delete(wd.watchers, id)
}

// WatchersNum is used to get the number of watcher.
func (wd *WatchDog) WatchersNum() int {
	wd.watchersRWM.RLock()
	defer wd.watchersRWM.RUnlock()
	return len(wd.watchers)
}

// isBlocked is used to prevent notice blocked watcher twice.
func (wd *WatchDog) isBlocked(id int32) bool {
	wd.blockedIDRWM.RLock()
	defer wd.blockedIDRWM.RUnlock()
	_, ok := wd.blockedID[id]
	return ok
}

func (wd *WatchDog) addBlockedID(id int32) {
	wd.blockedIDRWM.Lock()
	defer wd.blockedIDRWM.Unlock()
	wd.blockedID[id] = struct{}{}
}

func (wd *WatchDog) deleteBlockedID(id int32) {
	wd.blockedIDRWM.Lock()
	defer wd.blockedIDRWM.Unlock()
	delete(wd.blockedID, id)
}

// BlockedID is used to get blocked watcher id list.
func (wd *WatchDog) BlockedID() []int32 {
	wd.blockedIDRWM.RLock()
	defer wd.blockedIDRWM.RUnlock()
	list := make([]int32, 0, len(wd.blockedID))
	for id := range wd.blockedID {
		list = append(list, id)
	}
	return list
}

// Stop is used to close watch dog.
func (wd *WatchDog) Stop() {
	wd.stopOnce.Do(func() {
		wd.cancel()
		wd.wg.Wait()
		wd.onBlock = nil
	})
}
