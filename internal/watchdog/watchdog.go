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

// WatchDog is used to watch a loop in worker goroutine is blocked.
type WatchDog struct {
	logger  logger.Logger
	onBlock func(id int)

	logSrc   string
	interval atomic.Value
	nextID   int32

	watchers    map[int]chan struct{}
	watchersRWM sync.RWMutex

	// not notice blocked watcher twice
	blockedID    map[int]struct{}
	blockedIDRWM sync.RWMutex

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// New is used to create a new watch dog.
func New(lg logger.Logger, tag string, onBlock func(id int)) (*WatchDog, error) {
	if tag == "" {
		return nil, errors.New("empty watch dog tag")
	}
	wd := WatchDog{
		logger:     lg,
		onBlock:    onBlock,
		logSrc:     "watchdog-" + tag,
		nextID:     -1,
		watchers:   make(map[int]chan struct{}),
		blockedID:  make(map[int]struct{}),
		stopSignal: make(chan struct{}),
	}
	wd.interval.Store(defaultWatchInterval)
	return &wd, nil
}

// Watcher is used to create a new watcher.
func (wd *WatchDog) Watcher() (int, <-chan struct{}) {
	id := int(atomic.AddInt32(&wd.nextID, 1))
	watcher := make(chan struct{}, 1)
	wd.watchersRWM.Lock()
	defer wd.watchersRWM.Unlock()
	wd.watchers[id] = watcher
	return id, watcher
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

// Start is used to start watch loop.
func (wd *WatchDog) Start() {
	wd.wg.Add(1)
	go wd.watchLoop()
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
		blocked := wd.isBlocked(id)
		select {
		case watcher <- struct{}{}:
			if blocked {
				wd.deleteBlockedID(id)
			}
			continue
		default:
		}
		if blocked {
			return
		}
		wd.addBlockedID(id)
		go func(id int) {
			if r := recover(); r != nil {
				wd.log(logger.Fatal, xpanic.Printf(r, "WatchDog.watch"))
			}
			wd.logf(logger.Warning, "watcher [%d] is blocked", id)
			wd.onBlock(id)
		}(id)
	}
}

func (wd *WatchDog) getWatchers() map[int]chan struct{} {
	wd.watchersRWM.RLock()
	defer wd.watchersRWM.RUnlock()
	watchers := make(map[int]chan struct{}, len(wd.watchers))
	for id, watcher := range wd.watchers {
		watchers[id] = watcher
	}
	return watchers
}

// isBlocked is used to prevent notice blocked watcher twice.
func (wd *WatchDog) isBlocked(id int) bool {
	wd.blockedIDRWM.RLock()
	defer wd.blockedIDRWM.RUnlock()
	_, ok := wd.blockedID[id]
	return ok
}

func (wd *WatchDog) addBlockedID(id int) {
	wd.blockedIDRWM.Lock()
	defer wd.blockedIDRWM.Unlock()
	wd.blockedID[id] = struct{}{}
}

func (wd *WatchDog) deleteBlockedID(id int) {
	wd.blockedIDRWM.Lock()
	defer wd.blockedIDRWM.Unlock()
	delete(wd.blockedID, id)
}

// BlockedID is used to get blocked watcher id list.
func (wd *WatchDog) BlockedID() []int {
	wd.blockedIDRWM.RLock()
	defer wd.blockedIDRWM.RUnlock()
	list := make([]int, 0, len(wd.blockedID))
	for id := range wd.blockedID {
		list = append(list, id)
	}
	return list
}

// Close is used to close watch dog.
func (wd *WatchDog) Close() {
	wd.closeOnce.Do(func() {
		close(wd.stopSignal)
		wd.wg.Wait()
		wd.onBlock = nil
	})
}
