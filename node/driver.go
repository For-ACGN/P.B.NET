package node

import (
	"context"
	"sync"
	"time"

	"project/internal/bootstrap"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/random"
	"project/internal/xpanic"
)

type driver struct {
	ctx *Node

	// store node listeners
	nodeListeners      map[guid.GUID]map[uint64]*bootstrap.Listener
	nodeListenersIndex uint64
	nodeListenersRWM   sync.RWMutex

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newDriver(ctx *Node, config *Config) (*driver, error) {
	// cfg := config.Driver

	driver := driver{
		ctx:           ctx,
		nodeListeners: make(map[guid.GUID]map[uint64]*bootstrap.Listener),
	}
	driver.context, driver.cancel = context.WithCancel(context.Background())
	return &driver, nil
}

func (driver *driver) Drive() {
	driver.wg.Add(1)
	go driver.clientWatcher()
}

func (driver *driver) Close() {
	driver.cancel()
	driver.wg.Wait()
	driver.ctx = nil
}

func (driver *driver) log(lv logger.Level, log ...interface{}) {
	driver.ctx.logger.Println(lv, "driver", log...)
}

// clientWatcher is used to check Beacon is connected enough Nodes.
func (driver *driver) clientWatcher() {
	defer func() {
		if r := recover(); r != nil {
			driver.log(logger.Fatal, xpanic.Print(r, "driver.clientWatcher"))
			// restart queryLoop
			time.Sleep(time.Second)
			go driver.clientWatcher()
		} else {
			driver.wg.Done()
		}
	}()
	sleeper := random.NewSleeper()
	defer sleeper.Stop()
	for {
		select {
		case <-sleeper.SleepSecond(5, 10):
			driver.watchClient()
		case <-driver.context.Done():
			return
		}
	}
}

func (driver *driver) watchClient() {

}
