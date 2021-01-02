package beacon

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"hash"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/crypto/aes"
	"project/internal/crypto/ed25519"
	"project/internal/crypto/hmac"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/xpanic"
)

type driver struct {
	ctx *Beacon

	// store node listeners
	nodeListeners      map[guid.GUID]map[uint64]*bootstrap.Listener
	nodeListenersIndex uint64
	nodeListenersRWM   sync.RWMutex

	// about Sleeper.SleepSecond() in query
	sleepFixed  atomic.Value
	sleepRandom atomic.Value

	// interactive mode
	interactive   atomic.Value
	interactiveMu sync.Mutex

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newDriver(ctx *Beacon, config *Config) (*driver, error) {
	cfg := config.Driver

	if cfg.SleepFixed < 5 {
		return nil, errors.New("driver SleepFixed must >= 5")
	}
	if cfg.SleepRandom < 10 {
		return nil, errors.New("driver SleepRandom must >= 10")
	}

	driver := driver{
		ctx:           ctx,
		nodeListeners: make(map[guid.GUID]map[uint64]*bootstrap.Listener),
	}
	driver.SetSleepTime(cfg.SleepFixed, cfg.SleepRandom)
	interactive := cfg.Interactive
	driver.interactive.Store(interactive)
	driver.context, driver.cancel = context.WithCancel(context.Background())
	return &driver, nil
}

func (driver *driver) Drive() {
	driver.wg.Add(3)
	go driver.clientWatcher()
	go driver.queryLoop()
	go driver.modeWatcher()
}

func (driver *driver) Close() {
	driver.cancel()
	driver.wg.Wait()
	driver.ctx = nil
}

// NodeListener is used to get all Node listeners.
func (driver *driver) NodeListeners() map[guid.GUID]map[uint64]*bootstrap.Listener {
	nodeListeners := make(map[guid.GUID]map[uint64]*bootstrap.Listener)
	driver.nodeListenersRWM.RLock()
	defer driver.nodeListenersRWM.RUnlock()
	for nodeGUID, listeners := range driver.nodeListeners {
		nodeListeners[nodeGUID] = make(map[uint64]*bootstrap.Listener)
		for index, listener := range listeners {
			nodeListeners[nodeGUID][index] = listener
		}
	}
	return nodeListeners
}

// AddNodeListeners is used to add Node listeners(must be encrypted).
func (driver *driver) AddNodeListener(guid *guid.GUID, listener *bootstrap.Listener) {
	driver.nodeListenersRWM.Lock()
	defer driver.nodeListenersRWM.Unlock()
	// check Node GUID is exist
	listeners, ok := driver.nodeListeners[*guid]
	if !ok {
		listeners = make(map[uint64]*bootstrap.Listener)
		driver.nodeListeners[*guid] = listeners
	}
	// compare listeners
	for _, nListener := range listeners {
		if listener.Equal(nListener) {
			return
		}
	}
	index := driver.nodeListenersIndex
	listeners[index] = listener
	driver.nodeListenersIndex++
}

// DeleteNodeListener is used to delete Node listener.
func (driver *driver) DeleteNodeListener(guid *guid.GUID, index uint64) error {
	driver.nodeListenersRWM.Lock()
	defer driver.nodeListenersRWM.Unlock()
	// check Node GUID is exist
	listeners, ok := driver.nodeListeners[*guid]
	if !ok {
		return errors.New("node is not exist")
	}
	if _, ok := listeners[index]; ok {
		delete(listeners, index)
		return nil
	}
	return errors.New("node listener is not exist")
}

// DeleteAllNodeListener is used to delete Node's all listeners.
func (driver *driver) DeleteAllNodeListener(guid *guid.GUID) {
	driver.nodeListenersRWM.Lock()
	defer driver.nodeListenersRWM.Unlock()
	delete(driver.nodeListeners, *guid)
}

func (driver *driver) SetSleepTime(fixed, random uint) {
	driver.sleepFixed.Store(fixed)
	driver.sleepRandom.Store(random)
}

func (driver *driver) GetSleepTime() (uint, uint) {
	fixed := driver.sleepFixed.Load().(uint)
	rand := driver.sleepRandom.Load().(uint)
	return fixed, rand
}

func (driver *driver) EnableInteractiveMode() {
	driver.interactiveMu.Lock()
	defer driver.interactiveMu.Unlock()
	driver.interactive.Store(true)
}

func (driver *driver) DisableInteractiveMode() error {
	driver.interactiveMu.Lock()
	defer driver.interactiveMu.Unlock()
	if !driver.IsInInteractiveMode() {
		return errors.New("already disable interactive mode")
	}
	// check virtual connections manager

	driver.interactive.Store(false)
	return nil
}

// IsInInteractiveMode is used to check is in interactive mode.
func (driver *driver) IsInInteractiveMode() bool {
	return driver.interactive.Load().(bool)
}

// UpdateNode is used to update Node if client find different GUID in certificate.
func (driver *driver) UpdateNode(ctx context.Context, cert *protocol.Certificate) (bool, error) {
	// select a Node's listener
	var listener *bootstrap.Listener
	for _, listeners := range driver.NodeListeners() {
		for _, l := range listeners {
			listener = l
			break
		}
		break
	}
	if listener == nil {
		// TODO get more nodes
		return false, errors.New("no node listener")
	}
	// use protocol.CtrlGUID to skip check node guid in certificate
	client, err := driver.ctx.NewClient(ctx, listener, protocol.CtrlGUID, nil)
	if err != nil {
		return false, err
	}
	defer client.Close()
	// send connect operation and authenticate
	conn := client.Conn
	_, err = conn.Write([]byte{protocol.BeaconOperationUpdate})
	if err != nil {
		return false, errors.Wrap(err, "failed to send update operation")
	}
	err = client.Authenticate()
	if err != nil {
		return false, err
	}
	// pack request
	buf := bytes.NewBuffer(make([]byte, 0, guid.Size+ed25519.PublicKeySize))
	buf.Write(cert.GUID[:])
	buf.Write(cert.PublicKey)
	requestData := buf.Bytes()
	beaconGUID := driver.ctx.global.GUID()
	unr := protocol.UpdateNodeRequest{GUID: *beaconGUID}
	var h hash.Hash
	func() { // for clean session key
		sessionKey := driver.ctx.global.SessionKey()
		key := sessionKey.Get()
		defer sessionKey.Put(key)
		// encrypt
		unr.EncData, err = aes.CBCEncrypt(requestData, key, key[:aes.IVSize])
		if err != nil {
			panic("driver UpdateNode internal error: " + err.Error())
		}
		// HMAC
		h = hmac.New(sha256.New, key)
	}()
	h.Write(beaconGUID[:])
	h.Write(requestData)
	unr.Hash = h.Sum(nil)
	// send request
	err = unr.Validate()
	if err != nil {
		panic("driver UpdateNode internal error: " + err.Error())
	}
	buf.Reset()
	unr.Pack(buf)
	_, err = buf.WriteTo(conn)
	if err != nil {
		return false, errors.Wrap(err, "failed to send update node request")
	}
	// read response
	client.SetRandomDeadline(15, 30)
	_, err = buf.ReadFrom(io.LimitReader(conn, protocol.UpdateNodeResponseSize))
	if err != nil {
		return false, errors.Wrap(err, "failed to receive update node response")
	}
	response := protocol.NewUpdateNodeResponse()
	err = response.Unpack(buf.Bytes())
	if err != nil {
		return false, errors.WithStack(err)
	}
	var ok []byte // only the first byte is useful
	func() {      // for clean session key
		sessionKey := driver.ctx.global.SessionKey()
		key := sessionKey.Get()
		defer sessionKey.Put(key)
		// encrypt
		ok, err = aes.CBCDecrypt(response.EncData, key, key[:aes.IVSize])
		if err != nil {
			panic("driver UpdateNode internal error: " + err.Error())
		}
	}()
	h.Reset()
	h.Write(beaconGUID[:])
	h.Write(ok)
	if subtle.ConstantTimeCompare(h.Sum(nil), response.Hash) != 1 {
		return false, errors.New("incorrect hash in update node response")
	}
	if ok[0] != protocol.UpdateNodeResponseOK {
		return false, nil
	}
	return true, nil
}

// func (driver *driver) logf(lv logger.Level, format string, log ...interface{}) {
// 	driver.ctx.logger.Printf(lv, "driver", format, log...)
// }

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
	if !driver.IsInInteractiveMode() {
		return
	}
	// check is enough
	if len(driver.ctx.sender.Clients()) >= driver.ctx.sender.GetMaxConns() {
		return
	}
	// connect node
	for nodeGUID, listeners := range driver.NodeListeners() {
		var listener *bootstrap.Listener
		for _, listener = range listeners {
			break
		}
		if listener == nil {
			continue
		}
		// tempListener := listener.Decrypt()
		_ = driver.ctx.sender.Synchronize(driver.context, &nodeGUID, listener)
		// tempListener.Destroy()
		if len(driver.ctx.sender.Clients()) >= driver.ctx.sender.GetMaxConns() {
			return
		}
	}
}

// queryLoop is used to query message from Controller.
func (driver *driver) queryLoop() {
	defer func() {
		if r := recover(); r != nil {
			driver.log(logger.Fatal, xpanic.Print(r, "driver.queryLoop"))
			// restart queryLoop
			time.Sleep(time.Second)
			go driver.queryLoop()
		} else {
			driver.wg.Done()
		}
	}()
	sleeper := random.NewSleeper()
	defer sleeper.Stop()
	for {
		sleepFixed := driver.sleepFixed.Load().(uint)
		sleepRandom := driver.sleepRandom.Load().(uint)
		select {
		case <-sleeper.SleepSecond(sleepFixed, sleepRandom):
			driver.query()
		case <-driver.context.Done():
			return
		}
	}
}

func (driver *driver) query() {
	// check if connect some Nodes(maybe in interactive mode)
	if len(driver.ctx.sender.Clients()) > 0 {
		err := driver.ctx.sender.Query()
		if err != nil {
			driver.log(logger.Warning, "failed to query:", err)
		}
		return
	}
}

func (driver *driver) modeWatcher() {
	defer func() {
		if r := recover(); r != nil {
			driver.log(logger.Fatal, xpanic.Print(r, "driver.modeWatcher"))
			// restart queryLoop
			time.Sleep(time.Second)
			go driver.modeWatcher()
		} else {
			driver.wg.Done()
		}
	}()
	driver.watchMode()
}

func (driver *driver) watchMode() {

}
