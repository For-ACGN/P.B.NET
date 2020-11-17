package controller

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/logger"
	"project/internal/xpanic"
)

type boot struct {
	ctx *Ctrl

	// key = mBoot.Tag
	clients    map[string]*bootClient
	clientsRWM sync.RWMutex

	closing int32
}

func newBoot(ctx *Ctrl) *boot {
	return &boot{
		ctx:     ctx,
		clients: make(map[string]*bootClient),
	}
}

func (boot *boot) Add(m *mBoot) error {
	if atomic.LoadInt32(&boot.closing) != 0 {
		return errors.New("boot is closed")
	}
	boot.clientsRWM.Lock()
	defer boot.clientsRWM.Unlock()
	// check exist
	if _, ok := boot.clients[m.Tag]; ok {
		return errors.Errorf("boot %s is running", m.Tag)
	}
	// load bootstrap
	bc := bootClient{
		ctx:      boot,
		tag:      m.Tag,
		interval: time.Duration(m.Interval) * time.Second,
		logSrc:   "boot-" + m.Tag,
	}
	bc.context, bc.cancel = context.WithCancel(context.Background())
	b, err := bootstrap.Load(bc.context, m.Mode, []byte(m.Config),
		boot.ctx.global.CertPool,
		boot.ctx.global.ProxyPool,
		boot.ctx.global.DNSClient,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to load bootstrap %s", m.Tag)
	}
	bc.boot = b
	boot.clients[m.Tag] = &bc
	bc.Boot()
	return nil
}

func (boot *boot) Delete(tag string) error {
	if client, ok := boot.Clients()[tag]; ok {
		client.Stop()
		return nil
	}
	return errors.Errorf("boot: %s is not exist", tag)
}

func (boot *boot) Clients() map[string]*bootClient {
	boot.clientsRWM.RLock()
	defer boot.clientsRWM.RUnlock()
	clients := make(map[string]*bootClient, len(boot.clients))
	for key, client := range boot.clients {
		clients[key] = client
	}
	return clients
}

func (boot *boot) Close() {
	atomic.StoreInt32(&boot.closing, 1)
	for {
		// stop all boot client
		for _, client := range boot.Clients() {
			client.Stop()
		}
		// wait close
		time.Sleep(10 * time.Millisecond)
		if len(boot.Clients()) == 0 {
			break
		}
	}
	boot.ctx = nil
}

type bootClient struct {
	ctx *boot

	tag      string
	interval time.Duration
	logSrc   string
	boot     bootstrap.Bootstrap

	context   context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func (bc *bootClient) Boot() {
	bc.wg.Add(1)
	go bc.bootLoop()
}

func (bc *bootClient) Stop() {
	bc.closeOnce.Do(func() {
		bc.cancel()
		bc.wg.Wait()
		bc.boot = nil
	})
}

func (bc *bootClient) logf(lv logger.Level, format string, log ...interface{}) {
	bc.ctx.ctx.logger.Printf(lv, bc.logSrc, format, log...)
}

func (bc *bootClient) log(lv logger.Level, log ...interface{}) {
	bc.ctx.ctx.logger.Println(lv, bc.logSrc, log...)
}

func (bc *bootClient) bootLoop() {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "bootClient.bootLoop")
			bc.log(logger.Fatal, err)
		}
		// delete boot client
		bc.ctx.clientsRWM.Lock()
		defer bc.ctx.clientsRWM.Unlock()
		delete(bc.ctx.clients, bc.tag)
		bc.logf(logger.Info, "boot %s stopped", bc.tag)
		bc.wg.Done()
	}()
	if bc.resolve() {
		return
	}
	ticker := time.NewTicker(bc.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if bc.resolve() {
				return
			}
		case <-bc.context.Done():
			return
		}
	}
}

func (bc *bootClient) resolve() bool {
	var err error
	defer func() {
		if err != nil {
			bc.log(logger.Warning, err)
		}
	}()
	nodes, err := bc.boot.Resolve()
	if err != nil {
		return false
	}
	// add to sender

	nodes[0] = nil
	return true
}
