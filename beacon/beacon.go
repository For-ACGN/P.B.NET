package beacon

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
)

// Beacon send messages to Controller.
type Beacon struct {
	logger     *gLogger    // global logger
	global     *global     // certificate, proxy, dns, time syncer, and ...
	syncer     *syncer     // sync network guid
	clientMgr  *clientMgr  // clients manager
	register   *register   // about register to Controller
	sender     *sender     // send message to controller
	messageMgr *messageMgr // message manager
	handler    *handler    // handle message from controller
	worker     *worker     // do work
	driver     *driver     // control all modules
	Test       *Test       // internal test module

	once sync.Once
	wait chan struct{}
	exit chan error
}

// New is used to create a Beacon from configuration.
func New(cfg *Config) (*Beacon, error) {
	beacon := new(Beacon)
	// logger
	lg, err := newLogger(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize logger")
	}
	beacon.logger = lg
	// global
	global, err := newGlobal(beacon.logger, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize global")
	}
	beacon.global = global
	// syncer
	syncer, err := newSyncer(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize syncer")
	}
	beacon.syncer = syncer
	// client manager
	clientMgr, err := newClientManager(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize client manager")
	}
	beacon.clientMgr = clientMgr
	// register
	register, err := newRegister(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize register")
	}
	beacon.register = register
	// sender
	sender, err := newSender(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize sender")
	}
	beacon.sender = sender
	// message manager
	beacon.messageMgr = newMessageManager(beacon, cfg)
	// handler
	beacon.handler = newHandler(beacon)
	// worker
	worker, err := newWorker(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize worker")
	}
	beacon.worker = worker
	// driver
	driver, err := newDriver(beacon, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize worker")
	}
	beacon.driver = driver
	// test
	beacon.Test = newTest(cfg)
	// wait and exit
	beacon.wait = make(chan struct{})
	beacon.exit = make(chan error, 1)
	return beacon, nil
}

// HijackLogWriter is used to hijack all packages that call functions like log.Println().
func (beacon *Beacon) HijackLogWriter() {
	logger.HijackLogWriter(logger.Error, "pkg", beacon.logger)
}

func (beacon *Beacon) fatal(err error, msg string) error {
	err = errors.WithMessage(err, msg)
	beacon.logger.Println(logger.Fatal, "main", err)
	beacon.Exit(nil)
	close(beacon.wait)
	return err
}

// Main is used to run Beacon, it will block until exit or return error.
func (beacon *Beacon) Main() error {
	const src = "main"
	// start log sender
	beacon.logger.StartSender()
	// synchronize time
	if beacon.Test.options.SkipSynchronizeTime {
		beacon.global.TimeSyncer.StartWalker()
	} else {
		err := beacon.global.TimeSyncer.Start()
		if err != nil {
			return beacon.fatal(err, "failed to synchronize time")
		}
	}
	now := beacon.global.Now().Local()
	beacon.global.SetStartupTime(now)
	nowStr := now.Format(logger.TimeLayout)
	beacon.logger.Println(logger.Info, src, "time:", nowStr)
	// start register
	err := beacon.register.Register()
	if err != nil {
		return beacon.fatal(err, "failed to register")
	}
	// driver
	beacon.driver.Drive()
	beacon.logger.Print(logger.Info, src, "running")
	close(beacon.wait)
	return <-beacon.exit
}

// Wait is used to wait for Main().
func (beacon *Beacon) Wait() {
	<-beacon.wait
}

// Exit is used to exit with an error.
func (beacon *Beacon) Exit(err error) {
	const src = "exit"
	beacon.once.Do(func() {
		beacon.logger.CloseSender()
		beacon.driver.Close()
		beacon.logger.Print(logger.Info, src, "driver is stopped")
		beacon.handler.Cancel()
		beacon.worker.Close()
		beacon.logger.Print(logger.Info, src, "worker is stopped")
		beacon.handler.Close()
		beacon.logger.Print(logger.Info, src, "handler is stopped")
		beacon.messageMgr.Close()
		beacon.logger.Print(logger.Info, src, "message manager is stopped")
		beacon.sender.Close()
		beacon.logger.Print(logger.Info, src, "sender is stopped")
		beacon.register.Close()
		beacon.logger.Print(logger.Info, src, "register is closed")
		beacon.clientMgr.Close()
		beacon.logger.Print(logger.Info, src, "client manager is closed")
		beacon.syncer.Close()
		beacon.logger.Print(logger.Info, src, "syncer is stopped")
		beacon.global.Close()
		beacon.logger.Print(logger.Info, src, "global is closed")
		beacon.logger.Print(logger.Info, src, "beacon is stopped")
		beacon.logger.Close()
		beacon.exit <- err
		close(beacon.exit)
	})
}

// GUID is used to get Beacon GUID.
func (beacon *Beacon) GUID() *guid.GUID {
	return beacon.global.GUID()
}

// Synchronize is used to connect a Node and start to synchronize.
func (beacon *Beacon) Synchronize(
	ctx context.Context,
	guid *guid.GUID,
	listener *bootstrap.Listener,
) error {
	return beacon.sender.Synchronize(ctx, guid, listener)
}

// Disconnect is used to disconnect Node.
func (beacon *Beacon) Disconnect(guid *guid.GUID) error {
	return beacon.sender.Disconnect(guid)
}

// Send is used to send message to Controller.
func (beacon *Beacon) Send(
	ctx context.Context,
	command []byte,
	message []byte,
	deflate bool,
) error {
	return beacon.sender.Send(ctx, command, message, deflate)
}

// SendRT is used to send message to Controller and get response.
func (beacon *Beacon) SendRT(
	ctx context.Context,
	command []byte,
	message messages.RoundTripper,
	deflate bool,
	timeout time.Duration,
) (interface{}, error) {
	return beacon.messageMgr.Send(ctx, command, message, deflate, timeout)
}

// Query is used to query message from Controller.
func (beacon *Beacon) Query() error {
	return beacon.sender.Query()
}

// NodeListeners is used to get all Node listeners.
func (beacon *Beacon) NodeListeners() map[guid.GUID]map[uint64]*bootstrap.Listener {
	return beacon.driver.NodeListeners()
}
