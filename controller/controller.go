package controller

import (
	"bytes"
	"context"
	"io/ioutil"
	"sync"
	"time"

	"github.com/axgle/mahonia"
	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/crypto/cert"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
)

// Ctrl is controller.
// broadcast messages to Nodes, send messages to Nodes or Beacons.
type Ctrl struct {
	logger     *gLogger    // global logger
	global     *global     // certificate, proxy, dns, time syncer, and ...
	database   *database   // database
	syncer     *syncer     // receive message
	clientMgr  *clientMgr  // client manager
	sender     *sender     // broadcast and send message
	messageMgr *messageMgr // message manager
	actionMgr  *actionMgr  // action manager
	handler    *handler    // handle message from Node or Beacon
	worker     *worker     // do work
	boot       *boot       // auto discover bootstrap node listeners
	web        *web        // web server
	Test       *Test       // test module

	once sync.Once
	wait chan struct{}
	exit chan error
}

// New is used to create controller from configuration.
func New(cfg *Config) (*Ctrl, error) {
	ctrl := new(Ctrl)
	// logger
	lg, err := newLogger(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize logger")
	}
	ctrl.logger = lg
	// global
	global, err := newGlobal(ctrl.logger, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize global")
	}
	ctrl.global = global
	// database
	database, err := newDatabase(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize database")
	}
	ctrl.database = database
	// syncer
	syncer, err := newSyncer(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize syncer")
	}
	ctrl.syncer = syncer
	// client manager
	clientMgr, err := newClientManager(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize client manager")
	}
	ctrl.clientMgr = clientMgr
	// sender
	sender, err := newSender(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize sender")
	}
	ctrl.sender = sender
	// message manager
	ctrl.messageMgr = newMessageManager(ctrl, cfg)
	// action manager
	ctrl.actionMgr = newActionManager(ctrl, cfg)
	// handler
	ctrl.handler = newHandler(ctrl)
	// worker
	worker, err := newWorker(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize worker")
	}
	ctrl.worker = worker
	// boot
	ctrl.boot = newBoot(ctrl)
	// http server
	web, err := newWeb(ctrl, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize web server")
	}
	ctrl.web = web
	// test
	ctrl.Test = newTest(ctrl, cfg)
	// wait and exit
	ctrl.wait = make(chan struct{}, 2)
	ctrl.exit = make(chan error, 1)
	return ctrl, nil
}

// HijackLogWriter is used to hijack all packages that use log.Print().
func (ctrl *Ctrl) HijackLogWriter() {
	logger.HijackLogWriter(ctrl.logger)
}

func (ctrl *Ctrl) fatal(err error, msg string) error {
	err = errors.WithMessage(err, msg)
	ctrl.logger.Println(logger.Fatal, "main", err)
	ctrl.Exit(nil)
	return err
}

// Main is used to run Controller, it will block until exit or return error.
func (ctrl *Ctrl) Main() error {
	const src = "main"
	defer func() { ctrl.wait <- struct{}{} }()
	// synchronize time
	if ctrl.Test.options.SkipSynchronizeTime {
		ctrl.global.StartTimeSyncerAddLoop()
	} else {
		err := ctrl.global.StartTimeSyncer()
		if err != nil {
			return ctrl.fatal(err, "failed to synchronize time")
		}
	}
	// test client DNS option
	if !ctrl.Test.options.SkipTestClientDNS {
		err := ctrl.global.TestDNSOption(ctrl.clientMgr.GetDNSOptions())
		if err != nil {
			return errors.WithMessage(err, "failed to test client DNS option")
		}
	}
	now := ctrl.global.Now().Local().Format(logger.TimeLayout)
	ctrl.logger.Println(logger.Info, src, "time:", now)
	// start web server
	err := ctrl.web.Deploy()
	if err != nil {
		return ctrl.fatal(err, "failed to deploy web server")
	}
	ctrl.logger.Printf(logger.Info, src, "web server: https://%s/", ctrl.web.Address())
	ctrl.logger.Print(logger.Info, src, "controller is running")
	// wait to load controller keys
	if !ctrl.global.WaitLoadSessionKey() {
		return nil
	}
	ctrl.logger.Print(logger.Info, src, "load session key successfully")
	// load boots
	ctrl.logger.Print(logger.Info, src, "start discover bootstrap node listeners")
	boots, err := ctrl.database.SelectBoot()
	if err != nil {
		ctrl.logger.Println(logger.Error, src, "failed to select boot:", err)
		return nil
	}
	for i := 0; i < len(boots); i++ {
		err = ctrl.boot.Add(boots[i])
		if err != nil {
			ctrl.logger.Println(logger.Error, src, "failed to add boot:", err)
		}
	}
	ctrl.wait <- struct{}{}
	return <-ctrl.exit
}

// Wait is used to wait for Main().
func (ctrl *Ctrl) Wait() {
	<-ctrl.wait
}

// Exit is used to exit with a error.
func (ctrl *Ctrl) Exit(err error) {
	const src = "exit"
	ctrl.once.Do(func() {
		ctrl.Test.Close()
		ctrl.logger.Print(logger.Debug, src, "test module is stopped")
		ctrl.web.Close()
		ctrl.logger.Print(logger.Info, src, "web server is stopped")
		ctrl.boot.Close()
		ctrl.logger.Print(logger.Info, src, "boot is stopped")
		ctrl.handler.Cancel()
		ctrl.worker.Close()
		ctrl.logger.Print(logger.Info, src, "worker is stopped")
		ctrl.handler.Close()
		ctrl.logger.Print(logger.Info, src, "handler is stopped")
		ctrl.actionMgr.Close()
		ctrl.logger.Print(logger.Info, src, "action manager is stopped")
		ctrl.messageMgr.Close()
		ctrl.logger.Print(logger.Info, src, "message manager is stopped")
		ctrl.sender.Close()
		ctrl.logger.Print(logger.Info, src, "sender is stopped")
		ctrl.clientMgr.Close()
		ctrl.logger.Print(logger.Info, src, "client manager is closed")
		ctrl.syncer.Close()
		ctrl.logger.Print(logger.Info, src, "syncer is stopped")
		ctrl.global.Close()
		ctrl.logger.Print(logger.Info, src, "global is stopped")
		ctrl.logger.Print(logger.Info, src, "controller is stopped")
		ctrl.database.Close()
		ctrl.logger.Close()
		ctrl.exit <- err
		close(ctrl.exit)
	})
}

// LoadKeyFromFile is used to load session key and certificate pool from file.
func (ctrl *Ctrl) LoadKeyFromFile(sessionKeyPassword, certPassword []byte) error {
	sessionKey, err := ioutil.ReadFile(sessionKeyFile)
	if err != nil {
		return err
	}
	certData, err := ioutil.ReadFile(CertFile)
	if err != nil {
		return err
	}
	rawHash, err := ioutil.ReadFile(CertHash)
	if err != nil {
		return err
	}
	return ctrl.global.LoadKey(
		sessionKey, sessionKeyPassword,
		certData, rawHash, certPassword,
	)
}

// GetCertPool is used to get certificate pool.
func (ctrl *Ctrl) GetCertPool() *cert.Pool {
	return ctrl.global.CertPool
}

// KeyExchangePublicKey is used to get key exchange public key.
func (ctrl *Ctrl) KeyExchangePublicKey() []byte {
	return ctrl.global.KeyExchangePublicKey()
}

// PublicKey is used to get public key.
func (ctrl *Ctrl) PublicKey() []byte {
	return ctrl.global.PublicKey()
}

// BroadcastKey is used to get broadcast key.
func (ctrl *Ctrl) BroadcastKey() []byte {
	return ctrl.global.BroadcastKey()
}

// AddNodeListener is used to add Node listener.
func (ctrl *Ctrl) AddNodeListener(guid *guid.GUID, tag, mode, network, address string) error {
	nl := &mNodeListener{
		GUID:    guid[:],
		Tag:     tag,
		Mode:    mode,
		Network: network,
		Address: address,
	}
	return ctrl.database.InsertNodeListener(nl)
}

// Synchronize is used to connect a node and start to synchronize.
func (ctrl *Ctrl) Synchronize(
	ctx context.Context,
	guid *guid.GUID,
	listener *bootstrap.Listener,
) error {
	return ctrl.sender.Synchronize(ctx, guid, listener)
}

// Disconnect is used to disconnect Node.
func (ctrl *Ctrl) Disconnect(guid *guid.GUID) error {
	return ctrl.sender.Disconnect(guid)
}

// SendToNode is used to send messages to Node.
func (ctrl *Ctrl) SendToNode(
	ctx context.Context,
	guid *guid.GUID,
	command []byte,
	message interface{},
	deflate bool,
) error {
	return ctrl.sender.SendToNode(ctx, guid, command, message, deflate)
}

// SendToBeacon is used to send messages to Beacon.
func (ctrl *Ctrl) SendToBeacon(
	ctx context.Context,
	guid *guid.GUID,
	command []byte,
	message interface{},
	deflate bool,
) error {
	return ctrl.sender.SendToBeacon(ctx, guid, command, message, deflate)
}

// SendToNodeRT is used to send messages to Node and get response.
func (ctrl *Ctrl) SendToNodeRT(
	ctx context.Context,
	guid *guid.GUID,
	command []byte,
	message messages.RoundTripper,
	deflate bool,
	timeout time.Duration,
) (interface{}, error) {
	return ctrl.messageMgr.SendToNode(ctx, guid, command, message, deflate, timeout)
}

// SendToBeaconRT is used to send messages to Beacon and get response.
func (ctrl *Ctrl) SendToBeaconRT(
	ctx context.Context,
	guid *guid.GUID,
	command []byte,
	message messages.RoundTripper,
	deflate bool,
	timeout time.Duration,
) (interface{}, error) {
	return ctrl.messageMgr.SendToBeacon(ctx, guid, command, message, deflate, timeout)
}

// Broadcast is used to broadcast messages to all Nodes.
func (ctrl *Ctrl) Broadcast(command []byte, message interface{}, deflate bool) error {
	return ctrl.sender.Broadcast(command, message, deflate)
}

// EnableInteractiveMode is used to enable Beacon interactive mode.
func (ctrl *Ctrl) EnableInteractiveMode(guid *guid.GUID) {
	ctrl.sender.EnableInteractiveMode(guid)
}

// EnableInteractiveModeNew is used to enable Beacon interactive mode.
func (ctrl *Ctrl) EnableInteractiveModeNew(ctx context.Context, guid *guid.GUID) error {
	// check already enable interactive mode
	if ctrl.sender.IsInInteractiveMode(guid) {
		return nil
	}
	cm := messages.ChangeMode{Interactive: true}
	now := ctrl.global.Now()
	err := ctrl.sender.SendToBeacon(ctx, guid, messages.CMDBChangeMode, &cm, false)
	if err != nil {
		return err
	}
	latency := ctrl.global.Now().Sub(now)
	// acknowledge and the response will reach together mostly
	const interval = 50 * time.Millisecond
	timer := time.NewTicker(interval)
	defer timer.Stop()
	times := 3*int(latency/interval+1) + 60
	for i := 0; i < times; i++ {
		if ctrl.sender.IsInInteractiveMode(guid) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	const msg = "failed to enable interactive mode: receive response timeout"
	return errors.New(msg)
}

// DisableInteractiveModeNew is used to disable Beacon interactive mode.
func (ctrl *Ctrl) DisableInteractiveModeNew(ctx context.Context, guid *guid.GUID) error {
	// check already disable interactive mode
	if !ctrl.sender.IsInInteractiveMode(guid) {
		return nil
	}
	cm := messages.ChangeMode{Interactive: false}
	now := ctrl.global.Now()
	err := ctrl.sender.SendToBeacon(ctx, guid, messages.CMDBChangeMode, &cm, false)
	if err != nil {
		return err
	}
	latency := ctrl.global.Now().Sub(now)
	// acknowledge and the response will reach together mostly
	const interval = 50 * time.Millisecond
	timer := time.NewTicker(interval)
	defer timer.Stop()
	times := 3*int(latency/interval+1) + 60
	for i := 0; i < times; i++ {
		if !ctrl.sender.IsInInteractiveMode(guid) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	const msg = "failed to disable interactive mode: receive response timeout"
	return errors.New(msg)
}

// DeleteNode is used to delete Node.
func (ctrl *Ctrl) DeleteNode(guid *guid.GUID) error {
	err := ctrl.database.DeleteNode(guid)
	if err != nil {
		const format = "failed to delete node\n%s"
		return errors.Wrapf(err, format, guid.Print())
	}
	ctrl.sender.DeleteNode(guid)
	return nil
}

// DeleteNodeUnscoped is used to unscoped delete Node.
func (ctrl *Ctrl) DeleteNodeUnscoped(guid *guid.GUID) error {
	err := ctrl.database.DeleteNodeUnscoped(guid)
	if err != nil {
		const format = "failed to unscoped delete node\n%s"
		return errors.Wrapf(err, format, guid.Print())
	}
	ctrl.sender.DeleteNode(guid)
	return nil
}

// DeleteBeacon is used to delete Beacon.
func (ctrl *Ctrl) DeleteBeacon(guid *guid.GUID) error {
	err := ctrl.database.DeleteBeacon(guid)
	if err != nil {
		const format = "failed to delete beacon\n%s"
		return errors.Wrapf(err, format, guid.Print())
	}
	ctrl.sender.DeleteBeacon(guid)
	return nil
}

// DeleteBeaconUnscoped is used to unscoped delete Beacon.
func (ctrl *Ctrl) DeleteBeaconUnscoped(guid *guid.GUID) error {
	err := ctrl.database.DeleteBeaconUnscoped(guid)
	if err != nil {
		const format = "failed to unscoped delete beacon\n%s"
		return errors.Wrapf(err, format, guid.Print())
	}
	ctrl.sender.DeleteBeacon(guid)
	return nil
}

// ShellCode is used to send a shellcode to Beacon and return the execute result.
func (ctrl *Ctrl) ShellCode(
	ctx context.Context,
	guid *guid.GUID,
	method string,
	data []byte,
	timeout time.Duration,
) error {
	shellcode := messages.ShellCode{
		Method:    method,
		ShellCode: data,
	}
	if timeout < 1 {
		timeout = 10 * time.Second
	}
	reply, err := ctrl.messageMgr.SendToBeacon(ctx, guid,
		messages.CMDBShellCode, &shellcode, true, timeout)
	if err != nil {
		return err
	}
	result := reply.(*messages.ShellCodeResult)
	if result.Err != "" {
		return errors.New(result.Err)
	}
	return nil
}

// SingleShell is used to send command to Beacon, Beacon will use system shell to
// execute command and return the execute result, Controller select a decoder to
// decode the result, usually GBK in Windows, UTF-8 to other platform.
// <warning> this command can't block, otherwise it will return get reply timeout.
func (ctrl *Ctrl) SingleShell(
	ctx context.Context,
	guid *guid.GUID,
	cmd string,
	decoder string,
	timeout time.Duration,
) ([]byte, error) {
	d := mahonia.NewDecoder(decoder)
	if d == nil {
		return nil, errors.New("invalid decoder: " + decoder)
	}
	shell := messages.SingleShell{
		Command: cmd,
	}
	if timeout < 1 {
		timeout = 15 * time.Second
	}
	reply, err := ctrl.messageMgr.SendToBeacon(ctx, guid,
		messages.CMDBSingleShell, &shell, true, timeout)
	if err != nil {
		return nil, err
	}
	output := reply.(*messages.SingleShellOutput)
	buf := bytes.Buffer{}
	buf.Write(output.Output)
	if output.Err != "" {
		buf.WriteString(output.Err)
	}
	return buf.Bytes(), nil
}
