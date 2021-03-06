package node

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/dns"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/option"
	"project/internal/xpanic"
)

// clientMgr contains all clients from NewClient() and client options from Config.
// it can generate client tag, you can manage all clients here.
type clientMgr struct {
	ctx *Node

	// options from Config
	timeout   time.Duration
	proxyTag  string
	dnsOpts   dns.Options
	tlsConfig option.TLSConfig
	optsRWM   sync.RWMutex

	guid       *guid.Generator
	clients    map[guid.GUID]*Client
	clientsRWM sync.RWMutex
}

func newClientManager(ctx *Node, config *Config) (*clientMgr, error) {
	cfg := config.Client

	if cfg.Timeout < 10*time.Second {
		return nil, errors.New("client timeout must >= 10 seconds")
	}

	mgr := &clientMgr{
		ctx:       ctx,
		timeout:   cfg.Timeout,
		proxyTag:  cfg.ProxyTag,
		dnsOpts:   cfg.DNSOpts,
		tlsConfig: cfg.TLSConfig,
		guid:      guid.New(4, ctx.global.Now),
		clients:   make(map[guid.GUID]*Client),
	}
	mgr.tlsConfig.CertPool = ctx.global.CertPool
	return mgr, nil
}

func (mgr *clientMgr) GetTimeout() time.Duration {
	mgr.optsRWM.RLock()
	defer mgr.optsRWM.RUnlock()
	return mgr.timeout
}

func (mgr *clientMgr) GetProxyTag() string {
	mgr.optsRWM.RLock()
	defer mgr.optsRWM.RUnlock()
	return mgr.proxyTag
}

func (mgr *clientMgr) GetDNSOptions() *dns.Options {
	mgr.optsRWM.RLock()
	defer mgr.optsRWM.RUnlock()
	return mgr.dnsOpts.Clone()
}

func (mgr *clientMgr) GetTLSConfig() *option.TLSConfig {
	mgr.optsRWM.RLock()
	defer mgr.optsRWM.RUnlock()
	return &mgr.tlsConfig
}

func (mgr *clientMgr) SetTimeout(timeout time.Duration) error {
	if timeout < 10*time.Second {
		return errors.New("timeout must >= 10 seconds")
	}
	mgr.optsRWM.Lock()
	defer mgr.optsRWM.Unlock()
	mgr.timeout = timeout
	return nil
}

func (mgr *clientMgr) SetProxyTag(tag string) error {
	// check proxy is exist
	_, err := mgr.ctx.global.ProxyPool.Get(tag)
	if err != nil {
		return err
	}
	mgr.optsRWM.Lock()
	defer mgr.optsRWM.Unlock()
	mgr.proxyTag = tag
	return nil
}

func (mgr *clientMgr) SetDNSOptions(opts *dns.Options) {
	mgr.optsRWM.Lock()
	defer mgr.optsRWM.Unlock()
	mgr.dnsOpts = *opts.Clone()
}

func (mgr *clientMgr) SetTLSConfig(cfg *option.TLSConfig) error {
	_, err := cfg.Apply()
	if err != nil {
		return errors.WithStack(err)
	}
	mgr.optsRWM.Lock()
	defer mgr.optsRWM.Unlock()
	mgr.tlsConfig = *cfg
	mgr.tlsConfig.CertPool = mgr.ctx.global.CertPool
	return nil
}

// for NewClient()
func (mgr *clientMgr) Add(client *Client) {
	client.tag = mgr.guid.Get()
	mgr.clientsRWM.Lock()
	defer mgr.clientsRWM.Unlock()
	if _, ok := mgr.clients[*client.tag]; !ok {
		mgr.clients[*client.tag] = client
	}
}

// for client.Close()
func (mgr *clientMgr) Delete(tag *guid.GUID) {
	mgr.clientsRWM.Lock()
	defer mgr.clientsRWM.Unlock()
	delete(mgr.clients, *tag)
}

// Clients is used to get all clients.
func (mgr *clientMgr) Clients() map[guid.GUID]*Client {
	mgr.clientsRWM.RLock()
	defer mgr.clientsRWM.RUnlock()
	clients := make(map[guid.GUID]*Client, len(mgr.clients))
	for tag, client := range mgr.clients {
		clients[tag] = client
	}
	return clients
}

// Kill is used to close client. Must use cm.Clients(),
// because client.Close() will use cm.clientsRWM.
func (mgr *clientMgr) Kill(tag *guid.GUID) {
	if client, ok := mgr.Clients()[*tag]; ok {
		client.Close()
	}
}

// Close will close all active clients.
func (mgr *clientMgr) Close() {
	for {
		for _, client := range mgr.Clients() {
			client.Close()
		}
		time.Sleep(10 * time.Millisecond)
		if len(mgr.Clients()) == 0 {
			break
		}
	}
	mgr.guid.Close()
	mgr.ctx = nil
}

// messageMgr is used to manage messages that send to Controller.
// It will return the reply about the message.
type messageMgr struct {
	ctx *Node

	// 2 * sender.Timeout
	timeout time.Duration

	slots    map[guid.GUID]chan interface{}
	slotsRWM sync.RWMutex

	slotPool  sync.Pool
	timerPool sync.Pool

	guid *guid.Generator

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newMessageManager(ctx *Node, config *Config) *messageMgr {
	cfg := config.Sender

	mgr := messageMgr{
		ctx:     ctx,
		timeout: 2 * cfg.Timeout,
		slots:   make(map[guid.GUID]chan interface{}),
		guid:    guid.New(64, ctx.global.Now),
	}
	mgr.slotPool.New = func() interface{} {
		return make(chan interface{}, 1)
	}
	mgr.timerPool.New = func() interface{} {
		timer := time.NewTimer(time.Minute)
		timer.Stop()
		return timer
	}
	mgr.context, mgr.cancel = context.WithCancel(context.Background())
	mgr.wg.Add(1)
	go mgr.cleaner()
	return &mgr
}

func (mgr *messageMgr) createSlot() (*guid.GUID, chan interface{}) {
	id := mgr.guid.Get()
	ch := mgr.slotPool.Get().(chan interface{})
	mgr.slotsRWM.Lock()
	defer mgr.slotsRWM.Unlock()
	mgr.slots[*id] = ch
	return id, ch
}

func (mgr *messageMgr) destroySlot(id *guid.GUID, ch chan interface{}) {
	mgr.slotsRWM.Lock()
	defer mgr.slotsRWM.Unlock()
	// when read channel timeout, defer call destroySlot(),
	// the channel maybe has reply, try to clean it.
	select {
	case <-ch:
	default:
	}
	mgr.slotPool.Put(ch)
	delete(mgr.slots, *id)
}

// SendToNode is used to send message to Node and get the reply.
func (mgr *messageMgr) Send(
	ctx context.Context,
	command []byte,
	message messages.RoundTripper,
	deflate bool,
	timeout time.Duration,
) (interface{}, error) {
	// set message id
	id, reply := mgr.createSlot()
	defer mgr.destroySlot(id, reply)
	message.SetID(id)
	// send
	err := mgr.ctx.sender.Send(ctx, command, message, deflate)
	if err != nil {
		return nil, err
	}
	// get reply
	timer := mgr.timerPool.Get().(*time.Timer)
	defer mgr.timerPool.Put(timer)
	if timeout < 1 {
		timeout = mgr.timeout
	}
	timer.Reset(timeout)
	select {
	case resp := <-reply:
		if !timer.Stop() {
			<-timer.C
		}
		return resp, nil
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return nil, ctx.Err()
	case <-timer.C:
		return nil, errors.New("get reply timeout")
	}
}

// SendToNodeFromPlugin is used to send message to Node and get the reply.
func (mgr *messageMgr) SendFromPlugin(
	command []byte,
	message []byte,
	deflate bool,
	timeout time.Duration,
) ([]byte, error) {
	request := &messages.PluginRequest{
		Request: message,
	}
	reply, err := mgr.Send(mgr.context, command, request, deflate, timeout)
	if err != nil {
		return nil, err
	}
	return reply.([]byte), nil
}

// HandleReply is used to set reply, handler.Handle functions will call it.
func (mgr *messageMgr) HandleReply(id *guid.GUID, reply interface{}) {
	mgr.slotsRWM.RLock()
	defer mgr.slotsRWM.RUnlock()
	if ch, ok := mgr.slots[*id]; ok {
		select {
		case ch <- reply:
		default:
		}
	}
}

func (mgr *messageMgr) cleaner() {
	defer func() {
		if r := recover(); r != nil {
			buf := xpanic.Print(r, "messageMgr.cleaner")
			mgr.ctx.logger.Print(logger.Fatal, "message-manager", buf)
			// restart message cleaner
			time.Sleep(time.Second)
			go mgr.cleaner()
		} else {
			mgr.wg.Done()
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mgr.clean()
		case <-mgr.context.Done():
			return
		}
	}
}

func (mgr *messageMgr) clean() {
	mgr.slotsRWM.Lock()
	defer mgr.slotsRWM.Unlock()
	newMap := make(map[guid.GUID]chan interface{}, len(mgr.slots))
	for id, message := range mgr.slots {
		newMap[id] = message
	}
	mgr.slots = newMap
}

func (mgr *messageMgr) Close() {
	mgr.cancel()
	mgr.wg.Wait()
	mgr.guid.Close()
	mgr.ctx = nil
}
