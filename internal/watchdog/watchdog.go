package watchdog

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/xpanic"
)

const defaultWatchInterval = 10 * time.Second

// Callback is used to notice when watcher is blocked.
type Callback func(id int32)

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

	logSrc   string
	interval atomic.Value
	nextID   *int32

	watchers    map[int32]chan struct{}
	watchersRWM sync.RWMutex

	// not notice blocked watcher twice
	blockedID    map[int32]struct{}
	blockedIDRWM sync.RWMutex

	// about control
	stopSignal chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

// New is used to create a new watch dog.
func New(logger logger.Logger, tag string, onBlock Callback) (*WatchDog, error) {
	if tag == "" {
		return nil, errors.New("empty watch dog tag")
	}
	wd := WatchDog{
		logger:     logger,
		onBlock:    onBlock,
		logSrc:     "watchdog-" + tag,
		nextID:     new(int32),
		watchers:   make(map[int32]chan struct{}),
		blockedID:  make(map[int32]struct{}),
		stopSignal: make(chan struct{}),
	}
	*wd.nextID = -1
	wd.interval.Store(defaultWatchInterval)
	return &wd, nil
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
	go wd.watchLoop()
}

// GetWatchInterval is used to get watch interval.
func (wd *WatchDog) GetWatchInterval() time.Duration {
	return wd.interval.Load().(time.Duration)
}

// SetWatchInterval is used to set watch interval.
func (wd *WatchDog) SetWatchInterval(interval time.Duration) {
	if interval < 10*time.Millisecond || interval > 3*time.Minute {
		interval = defaultWatchInterval
	}
	wd.interval.Store(interval)
}

func (wd *WatchDog) logf(lv logger.Level, format string, log ...interface{}) {
	wd.logger.Printf(lv, wd.logSrc, format, log...)
}

func (wd *WatchDog) log(lv logger.Level, log ...interface{}) {
	wd.logger.Println(lv, wd.logSrc, log...)
}

func (wd *WatchDog) watchLoop() {
	defer wd.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			wd.log(logger.Fatal, xpanic.Print(r, "WatchDog.watchLoop"))
			// restart
			time.Sleep(time.Second)
			wd.wg.Add(1)
			go wd.watchLoop()
		}
	}()
	timer := time.NewTimer(wd.GetWatchInterval())
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			wd.watch()
		case <-wd.stopSignal:
			return
		}
		timer.Reset(wd.GetWatchInterval())
	}
}

func (wd *WatchDog) watch() {
	for id, watcher := range wd.getWatchers() {
		select {
		case watcher <- struct{}{}:
			if wd.isBlocked(id) {
				wd.deleteBlockedID(id)
			}
			continue
		default:
		}
		if wd.isBlocked(id) {
			return
		}
		wd.addBlockedID(id)
		go func(id int32, onBlock Callback) {
			if r := recover(); r != nil {
				wd.log(logger.Fatal, xpanic.Printf(r, "WatchDog.watch"))
			}
			wd.logf(logger.Warning, "watcher [%d] is blocked", id)
			onBlock(id)
		}(id, wd.onBlock)
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
		close(wd.stopSignal)
		wd.wg.Wait()
		wd.onBlock = nil
	})
}
