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

// Callback is used to notice when receiver is blocked.
type Callback func(ctx context.Context, id int32)

// Receiver is a receiver that spawned by WatchDog.
type Receiver struct {
	ctx    *WatchDog
	id     int32
	Signal <-chan struct{}
}

// Receive is used to receive signal.
func (r *Receiver) Receive() {
	select {
	case <-r.Signal:
	default:
	}
}

// Stop is used to stop this receiver.
func (r *Receiver) Stop() {
	r.ctx.deleteReceiver(r.id)
	r.ctx.deleteBlockedID(r.id)
	r.ctx = nil
}

// WatchDog is used to watch a loop in worker goroutine is blocked.
type WatchDog struct {
	logger  logger.Logger
	onBlock Callback

	logSrc string
	nextID *int32
	period atomic.Value

	receivers    map[int32]chan struct{}
	receiversRWM sync.RWMutex

	// not notice blocked receiver twice
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
		receivers: make(map[int32]chan struct{}),
		blockedID: make(map[int32]struct{}),
	}
	*wd.nextID = -1
	wd.period.Store(defaultPeriod)
	wd.ctx, wd.cancel = context.WithCancel(context.Background())
	return &wd
}

// Receiver is used to create a new receiver.
func (wd *WatchDog) Receiver() *Receiver {
	id := atomic.AddInt32(wd.nextID, 1)
	ch := make(chan struct{}, 1)
	receiver := Receiver{
		ctx:    wd,
		id:     id,
		Signal: ch,
	}
	wd.receiversRWM.Lock()
	defer wd.receiversRWM.Unlock()
	wd.receivers[id] = ch
	return &receiver
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
	if period < 50*time.Millisecond || period > 3*time.Minute {
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
			ticker.Reset(period)
		}
	}
}

func (wd *WatchDog) kick() {
	for id, receiver := range wd.getReceivers() {
		// kicking the dog
		select {
		case receiver <- struct{}{}:
			if wd.isBlocked(id) {
				wd.deleteBlockedID(id)
				wd.logf(logger.Info, "receiver [%d] is running", id)
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
			wd.logf(logger.Warning, "receiver [%d] is blocked", id)
			if wd.onBlock == nil {
				return
			}
			wd.onBlock(wd.ctx, id)
		}(id)
	}
}

func (wd *WatchDog) getReceivers() map[int32]chan struct{} {
	wd.receiversRWM.RLock()
	defer wd.receiversRWM.RUnlock()
	receivers := make(map[int32]chan struct{}, len(wd.receivers))
	for id, receiver := range wd.receivers {
		receivers[id] = receiver
	}
	return receivers
}

func (wd *WatchDog) deleteReceiver(id int32) {
	wd.receiversRWM.Lock()
	defer wd.receiversRWM.Unlock()
	delete(wd.receivers, id)
}

// ReceiversNum is used to get the number of receiver.
func (wd *WatchDog) ReceiversNum() int {
	wd.receiversRWM.RLock()
	defer wd.receiversRWM.RUnlock()
	return len(wd.receivers)
}

// isBlocked is used to prevent notice blocked receiver twice.
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

// BlockedID is used to get blocked receivers id list.
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
