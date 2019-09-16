package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/protocol"
	"project/internal/xpanic"
)

const (
	senderNode   = 0
	senderBeacon = 1
)

type broadcastTask struct {
	Command  []byte      // for Broadcast
	MessageI interface{} // for Broadcast
	Message  []byte      // for BroadcastPlugin
	Result   chan<- *protocol.BroadcastResult
}

type syncSendTask struct {
	Role     protocol.Role
	Target   []byte
	Command  []byte      // for Send
	MessageI interface{} // for Send
	Message  []byte      // for SendPlugin
	Result   chan<- *protocol.SyncResult
}

type syncReceiveTask struct {
	Role   protocol.Role
	GUID   []byte
	Height uint64
}

type sender struct {
	ctx              *CTRL
	maxBufferSize    int
	broadcastQueue   chan *broadcastTask
	syncSendQueue    chan *syncSendTask
	syncReceiveQueue chan *syncReceiveTask

	broadcastDonePool sync.Pool
	syncSendDonePool  sync.Pool
	broadcastRespPool sync.Pool
	syncRespPool      sync.Pool

	syncSendMs  [2]map[string]*sync.Mutex // role can be only one sync at th same time
	syncSendRWM [2]sync.RWMutex           // key=base64(sender guid) 0=node 1=beacon

	guid       *guid.GUID
	stopSignal chan struct{}
	wg         sync.WaitGroup
}

func newSender(ctx *CTRL, cfg *Config) (*sender, error) {
	// check config
	if cfg.SenderWorker < 1 {
		return nil, errors.New("sender number < 1")
	}
	if cfg.SenderQueueSize < 512 {
		return nil, errors.New("sender task queue size < 512")
	}
	sender := sender{
		ctx:              ctx,
		maxBufferSize:    cfg.MaxBufferSize,
		broadcastQueue:   make(chan *broadcastTask, cfg.SenderQueueSize),
		syncSendQueue:    make(chan *syncSendTask, cfg.SenderQueueSize),
		syncReceiveQueue: make(chan *syncReceiveTask, cfg.SenderQueueSize),
		guid:             guid.New(512*cfg.SenderWorker, ctx.global.Now),
		stopSignal:       make(chan struct{}),
	}
	sender.syncSendMs[senderNode] = make(map[string]*sync.Mutex)
	sender.syncSendMs[senderBeacon] = make(map[string]*sync.Mutex)
	// init sync pool
	sender.broadcastDonePool.New = func() interface{} {
		return make(chan *protocol.BroadcastResult, 1)
	}
	sender.syncSendDonePool.New = func() interface{} {
		return make(chan *protocol.SyncResult, 1)
	}
	sender.broadcastRespPool.New = func() interface{} {
		return make(chan *protocol.BroadcastResponse, 1)
	}
	sender.syncRespPool.New = func() interface{} {
		return make(chan *protocol.SyncResponse, 1)
	}
	// start senders
	for i := 0; i < cfg.SenderWorker; i++ {
		sender.wg.Add(1)
		go sender.worker()
	}
	return &sender, nil
}

// Broadcast is used to broadcast message to all nodes
// message will not be saved
func (sender *sender) Broadcast(
	command []byte,
	message interface{},
) (r *protocol.BroadcastResult) {
	done := sender.broadcastDonePool.Get().(chan *protocol.BroadcastResult)
	sender.broadcastQueue <- &broadcastTask{
		Command:  command,
		MessageI: message,
		Result:   done,
	}
	r = <-done
	sender.broadcastDonePool.Put(done)
	return
}

// Broadcast is used to broadcast(Async) message to all nodes
// message will not be saved
func (sender *sender) BroadcastAsync(
	command []byte,
	message interface{},
	done chan<- *protocol.BroadcastResult,
) {
	sender.broadcastQueue <- &broadcastTask{
		Command:  command,
		MessageI: message,
		Result:   done,
	}
}

// Broadcast is used to broadcast(plugin) message to all nodes
// message will not be saved
func (sender *sender) BroadcastPlugin(
	message []byte,
) (r *protocol.BroadcastResult) {
	done := sender.broadcastDonePool.Get().(chan *protocol.BroadcastResult)
	sender.broadcastQueue <- &broadcastTask{
		Message: message,
		Result:  done,
	}
	r = <-done
	sender.broadcastDonePool.Put(done)
	return
}

// Send is used to send message to Node or Beacon
// if role not online, node will save it
func (sender *sender) Send(
	role protocol.Role,
	target,
	command []byte,
	message interface{},
) (r *protocol.SyncResult) {
	done := sender.syncSendDonePool.Get().(chan *protocol.SyncResult)
	sender.syncSendQueue <- &syncSendTask{
		Role:     role,
		Target:   target,
		Command:  command,
		MessageI: message,
		Result:   done,
	}
	r = <-done
	sender.syncSendDonePool.Put(done)
	return
}

// Send is used to send(async) message to Node or Beacon
// if role not online, node will save it
func (sender *sender) SendAsync(
	role protocol.Role,
	target,
	command []byte,
	message interface{},
	done chan<- *protocol.SyncResult,
) {
	sender.syncSendQueue <- &syncSendTask{
		Role:     role,
		Target:   target,
		Command:  command,
		MessageI: message,
		Result:   done,
	}
}

// Send is used to send(plugin) message to Node or Beacon
// if role not online, node will save it
func (sender *sender) SendPlugin(
	role protocol.Role,
	target,
	message []byte,
) (r *protocol.SyncResult) {
	done := sender.syncSendDonePool.Get().(chan *protocol.SyncResult)
	sender.syncSendQueue <- &syncSendTask{
		Role:    role,
		Target:  target,
		Message: message,
		Result:  done,
	}
	r = <-done
	sender.syncSendDonePool.Put(done)
	return
}

// SyncRecv is used to sync controller receive
// notice node to delete message about Node or Beacon
// only for syncer.worker()
func (sender *sender) SyncReceive(
	role protocol.Role,
	guid []byte,
	height uint64,
) {
	sender.syncReceiveQueue <- &syncReceiveTask{
		Role:   role,
		GUID:   guid,
		Height: height,
	}
}

func (sender *sender) Close() {
	close(sender.stopSignal)
	sender.wg.Wait()
}

func (sender *sender) logf(l logger.Level, format string, log ...interface{}) {
	sender.ctx.Printf(l, "sender", format, log...)
}

func (sender *sender) log(l logger.Level, log ...interface{}) {
	sender.ctx.Print(l, "sender", log...)
}

func (sender *sender) logln(l logger.Level, log ...interface{}) {
	sender.ctx.Println(l, "sender", log...)
}

func (sender *sender) broadcastParallel(token, message []byte) (
	resp []*protocol.BroadcastResponse, success int) {
	sClients := sender.ctx.syncer.SyncerClients()
	l := len(sClients)
	if l == 0 {
		return nil, 0
	}
	// padding channels
	channels := make([]chan *protocol.BroadcastResponse, l)
	for i := 0; i < l; i++ {
		channels[i] = sender.broadcastRespPool.Get().(chan *protocol.BroadcastResponse)
	}
	// broadcast parallel
	index := 0
	for _, sc := range sClients {
		go func(s *sClient) {
			channels[index] <- s.Broadcast(token, message)
		}(sc)
		index += 1
	}
	// get response and put
	resp = make([]*protocol.BroadcastResponse, l)
	for i := 0; i < l; i++ {
		resp[i] = <-channels[i]
		if resp[i].Err == nil {
			success += 1
		}
		sender.broadcastRespPool.Put(channels[i])
	}
	return
}

func (sender *sender) syncSendParallel(token, message []byte) (
	resp []*protocol.SyncResponse, success int) {
	sClients := sender.ctx.syncer.SyncerClients()
	l := len(sClients)
	if l == 0 {
		return nil, 0
	}
	// padding channels
	channels := make([]chan *protocol.SyncResponse, l)
	for i := 0; i < l; i++ {
		channels[i] = sender.syncRespPool.Get().(chan *protocol.SyncResponse)
	}
	// sync send parallel
	index := 0
	for _, sc := range sClients {
		go func(s *sClient) {
			channels[index] <- s.SyncSend(token, message)
		}(sc)
		index += 1
	}
	// get response and put
	resp = make([]*protocol.SyncResponse, l)
	for i := 0; i < l; i++ {
		resp[i] = <-channels[i]
		if resp[i].Err == nil {
			success += 1
		}
		sender.syncRespPool.Put(channels[i])
	}
	return
}

func (sender *sender) syncReceiveParallel(token, message []byte) {
	sClients := sender.ctx.syncer.SyncerClients()
	l := len(sClients)
	if l == 0 {
		return
	}
	// must copy
	msg := make([]byte, len(message))
	copy(msg, message)
	// sync receive parallel
	for _, sc := range sClients {
		go func(s *sClient) {
			s.SyncReceive(token, msg)
		}(sc)
	}
}

// DeleteSyncSendM is used to delete syncSendM
// if delete role, must delete it
func (sender *sender) DeleteSyncSendM(role protocol.Role, guid string) {
	i := 0
	switch role {
	case protocol.Beacon:
		i = senderBeacon
	case protocol.Node:
		i = senderNode
	default:
		panic("invalid role")
	}
	sender.syncSendRWM[i].Lock()
	if _, ok := sender.syncSendMs[i][guid]; ok {
		delete(sender.syncSendMs[i], guid)
	}
	sender.syncSendRWM[i].Unlock()
}

// make sure send lock exist
func (sender *sender) lockRole(role protocol.Role, guid string) {
	i := 0
	switch role {
	case protocol.Beacon:
		i = senderBeacon
	case protocol.Node:
		i = senderNode
	}
	sender.syncSendRWM[i].Lock()
	if m, ok := sender.syncSendMs[i][guid]; ok {
		sender.syncSendRWM[i].Unlock()
		m.Lock()
	} else {
		sender.syncSendMs[i][guid] = new(sync.Mutex)
		sender.syncSendRWM[i].Unlock()
		sender.syncSendMs[i][guid].Lock()
	}
}

func (sender *sender) unlockRole(role protocol.Role, guid string) {
	i := 0
	switch role {
	case protocol.Beacon:
		i = senderBeacon
	case protocol.Node:
		i = senderNode
	}
	sender.syncSendRWM[i].RLock()
	if m, ok := sender.syncSendMs[i][guid]; ok {
		sender.syncSendRWM[i].RUnlock()
		m.Unlock()
	} else {
		sender.syncSendRWM[i].RUnlock()
	}
}

func (sender *sender) worker() {
	defer func() {
		if r := recover(); r != nil {
			err := xpanic.Error("sender.worker() panic:", r)
			sender.log(logger.Fatal, err)
			// restart worker
			time.Sleep(time.Second)
			sender.wg.Add(1)
			go sender.worker()
		}
		sender.wg.Done()
	}()
	var (
		// task
		bt  *broadcastTask
		sst *syncSendTask
		srt *syncReceiveTask

		// key
		node   *mNode
		beacon *mBeacon
		aesKey []byte
		aesIV  []byte

		// temp
		nodeSyncer   *nodeSyncer
		beaconSyncer *beaconSyncer
		roleGUID     string
		token        []byte
		err          error
	)
	// prepare buffer, msgpack encoder, base64 encoder
	// syncReceiveTask = 1 + guid.Size + 8
	minBufferSize := guid.Size + 9
	buffer := bytes.NewBuffer(make([]byte, minBufferSize))
	msgpackEncoder := msgpack.NewEncoder(buffer)
	base64Encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	hash := sha256.New()
	// prepare task objects
	preB := &protocol.Broadcast{
		SenderRole: protocol.Ctrl,
		SenderGUID: protocol.CtrlGUID,
	}
	preSS := &protocol.SyncSend{
		SenderRole: protocol.Ctrl,
		SenderGUID: protocol.CtrlGUID,
	}
	preSR := &protocol.SyncRecv{}
	// start handle task
	for {
		// check buffer capacity
		if buffer.Cap() > sender.maxBufferSize {
			buffer = bytes.NewBuffer(make([]byte, minBufferSize))
		}
		select {
		// --------------------------sync receive-------------------------
		case srt = <-sender.syncReceiveQueue:
			// check role
			if srt.Role != protocol.Node && srt.Role != protocol.Beacon {
				panic("sender.sender(): invalid srt.Role")
			}
			preSR.GUID = sender.guid.Get()
			preSR.Height = srt.Height
			preSR.Role = srt.Role
			preSR.RoleGUID = srt.GUID
			// sign
			buffer.Reset()
			buffer.Write(preSR.GUID)
			buffer.Write(convert.Uint64ToBytes(preSR.Height))
			buffer.WriteByte(preSR.Role.Byte())
			buffer.Write(preSR.RoleGUID)
			preSR.Signature = sender.ctx.global.Sign(buffer.Bytes())
			// pack syncReceive & token
			buffer.Reset()
			err = msgpackEncoder.Encode(&preSR)
			if err != nil {
				panic(err)
			}
			// send
			token = append(protocol.Ctrl.Bytes(), preSR.GUID...)
			sender.syncReceiveParallel(token, buffer.Bytes())
		// ---------------------------sync send---------------------------
		case sst = <-sender.syncSendQueue:
			result := protocol.SyncResult{}
			// check role
			if sst.Role != protocol.Node && sst.Role != protocol.Beacon {
				if sst.Result != nil {
					result.Err = protocol.ErrInvalidRole
					sst.Result <- &result
				}
				continue
			}
			preSS.GUID = sender.guid.Get()
			// pack message(interface)
			if sst.MessageI != nil {
				buffer.Reset()
				err = msgpackEncoder.Encode(sst.MessageI)
				if err != nil {
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
				sst.Message = append(sst.Command, buffer.Bytes()...)
			}
			// set key
			switch sst.Role {
			case protocol.Beacon:
				beacon, err = sender.ctx.db.SelectBeacon(sst.Target)
				if err != nil {
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
				aesKey = beacon.SessionKey
				aesIV = beacon.SessionKey[:aes.IVSize]
			case protocol.Node:
				node, err = sender.ctx.db.SelectNode(sst.Target)
				if err != nil {
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
				aesKey = node.SessionKey
				aesIV = node.SessionKey[:aes.IVSize]
			default:
				panic("invalid sst.Role")
			}
			// hash
			hash.Reset()
			hash.Write(sst.Message)
			preSS.Hash = hash.Sum(nil)
			// encrypt
			preSS.Message, err = aes.CBCEncrypt(sst.Message, aesKey, aesIV)
			if err != nil {
				if sst.Result != nil {
					result.Err = err
					sst.Result <- &result
				}
				continue
			}
			preSS.ReceiverRole = sst.Role
			preSS.ReceiverGUID = sst.Target
			// set sync height
			buffer.Reset()
			_, _ = base64Encoder.Write(sst.Target)
			_ = base64Encoder.Close()
			roleGUID = buffer.String()
			sender.lockRole(sst.Role, roleGUID)
			switch sst.Role {
			case protocol.Beacon:
				beaconSyncer, err = sender.ctx.db.SelectBeaconSyncer(sst.Target)
				if err != nil {
					sender.unlockRole(sst.Role, roleGUID)
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
				beaconSyncer.RLock()
				preSS.Height = beaconSyncer.CtrlSend
				beaconSyncer.RUnlock()
			case protocol.Node:
				nodeSyncer, err = sender.ctx.db.SelectNodeSyncer(sst.Target)
				if err != nil {
					sender.unlockRole(sst.Role, roleGUID)
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
				nodeSyncer.RLock()
				preSS.Height = nodeSyncer.CtrlSend
				nodeSyncer.RUnlock()
			default:
				sender.unlockRole(sst.Role, roleGUID)
				panic("invalid sst.Role")
			}
			// sign
			buffer.Reset()
			buffer.Write(preSS.GUID)
			buffer.Write(convert.Uint64ToBytes(preSS.Height))
			buffer.Write(preSS.Message)
			buffer.Write(preSS.Hash)
			buffer.WriteByte(preSS.SenderRole.Byte())
			buffer.Write(preSS.SenderGUID)
			buffer.WriteByte(preSS.ReceiverRole.Byte())
			buffer.Write(preSS.ReceiverGUID)
			preSS.Signature = sender.ctx.global.Sign(buffer.Bytes())
			// pack protocol.syncSend and token
			buffer.Reset()
			err = msgpackEncoder.Encode(&preSS)
			if err != nil {
				sender.unlockRole(sst.Role, roleGUID)
				if sst.Result != nil {
					result.Err = err
					sst.Result <- &result
				}
				continue
			}
			// !!! think order
			// first must add send height
			switch sst.Role {
			case protocol.Beacon:
				err = sender.ctx.db.UpdateBSCtrlSend(sst.Target, preSS.Height+1)
			case protocol.Node:
				err = sender.ctx.db.UpdateNSCtrlSend(sst.Target, preSS.Height+1)
			default:
				sender.unlockRole(sst.Role, roleGUID)
				panic("invalid sst.Role")
			}
			if err != nil {
				sender.unlockRole(sst.Role, roleGUID)
				if sst.Result != nil {
					result.Err = err
					sst.Result <- &result
				}
				continue
			}
			// !!! think order
			// second send
			token = append(protocol.Ctrl.Bytes(), preSS.GUID...)
			result.Response, result.Success =
				sender.syncSendParallel(token, buffer.Bytes())
			// !!! think order
			// rollback send height
			if result.Success == 0 {
				switch sst.Role {
				case protocol.Beacon:
					err = sender.ctx.db.UpdateBSCtrlSend(sst.Target, preSS.Height)
				case protocol.Node:
					err = sender.ctx.db.UpdateNSCtrlSend(sst.Target, preSS.Height)
				default:
					sender.unlockRole(sst.Role, roleGUID)
					panic("invalid sst.Role")
				}
				if err != nil {
					sender.unlockRole(sst.Role, roleGUID)
					if sst.Result != nil {
						result.Err = err
						sst.Result <- &result
					}
					continue
				}
			}
			sender.unlockRole(sst.Role, roleGUID)
			if sst.Result != nil {
				sst.Result <- &result
			}
		// ---------------------------broadcast---------------------------
		case bt = <-sender.broadcastQueue:
			result := protocol.BroadcastResult{}
			preB.GUID = sender.guid.Get()
			// pack message
			if bt.MessageI != nil {
				buffer.Reset()
				err = msgpackEncoder.Encode(bt.MessageI)
				if err != nil {
					if bt.Result != nil {
						result.Err = err
						bt.Result <- &result
					}
					continue
				}
				bt.Message = append(bt.Command, buffer.Bytes()...)
			}
			// hash
			hash.Reset()
			hash.Write(bt.Message)
			preB.Hash = hash.Sum(nil)
			// encrypt
			preB.Message, err = sender.ctx.global.Encrypt(bt.Message)
			if err != nil {
				if bt.Result != nil {
					result.Err = err
					bt.Result <- &result
				}
				continue
			}
			// sign
			buffer.Reset()
			buffer.Write(preB.GUID)
			buffer.Write(preB.Message)
			buffer.Write(preB.Hash)
			buffer.WriteByte(preB.SenderRole.Byte())
			buffer.Write(preB.SenderGUID)
			preB.Signature = sender.ctx.global.Sign(buffer.Bytes())
			// pack broadcast & token
			buffer.Reset()
			err = msgpackEncoder.Encode(&preB)
			if err != nil {
				if bt.Result != nil {
					result.Err = err
					bt.Result <- &result
				}
				continue
			}
			// send
			token = append(protocol.Ctrl.Bytes(), preB.GUID...)
			result.Response, result.Success =
				sender.broadcastParallel(token, buffer.Bytes())
			if bt.Result != nil {
				bt.Result <- &result
			}
		case <-sender.stopSignal:
			return
		}
	}
}
