package watchdog

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"project/internal/logger"
)

// WatchDog is used to watch a loop in worker goroutine is blocked.
type WatchDog struct {
	logger  logger.Logger
	tag     string
	onBlock func(id int)

	nextID int32

	watchers    map[int]chan struct{}
	watchersRWM sync.RWMutex

	// not notice again
	blockedID    map[int]struct{}
	blockedIDRWM sync.RWMutex

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// New is used to create a new watch dog.
func New(logger logger.Logger, tag string, onBlock func(id int)) (*WatchDog, error) {
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	wd := WatchDog{
		logger:     logger,
		tag:        tag,
		onBlock:    onBlock,
		nextID:     -1,
		watchers:   make(map[int]chan struct{}),
		blockedID:  make(map[int]struct{}),
		stopSignal: make(chan struct{}),
	}
	return &wd, nil
}

// Watch is used to create a new watcher.
func (wd *WatchDog) Watch() (int, <-chan struct{}) {
	id := int(atomic.AddInt32(&wd.nextID, 1))
	watcher := make(chan struct{}, 1)
	wd.watchersRWM.Lock()
	defer wd.watchersRWM.Unlock()
	wd.watchers[id] = watcher
	return id, watcher
}

// Start is used to start watch loop.
func (wd *WatchDog) Start() {
	wd.wg.Add(1)
	go wd.watchLoop()
}

func (wd *WatchDog) watchLoop() {
	defer wd.wg.Done()

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
