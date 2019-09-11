package controller

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/convert"
	"project/internal/logger"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/xnet"
	"project/internal/xpanic"
)

// NodeGUID != nil for sync or other
// NodeGUID = nil for trust node
// NodeGUID = controller guid for discovery
type clientCfg struct {
	Node       *bootstrap.Node
	NodeGUID   []byte
	MsgHandler func(msg []byte)
	CloseLog   bool
	xnet.Config
}

type client struct {
	ctx        *CTRL
	node       *bootstrap.Node
	guid       []byte
	closeLog   bool
	conn       *xnet.Conn
	slots      []*protocol.Slot
	heartbeatC chan struct{}
	replyTimer *time.Timer
	inClose    int32
	closeOnce  sync.Once
	stopSignal chan struct{}
	wg         sync.WaitGroup
}

func newClient(ctx *CTRL, cfg *clientCfg) (*client, error) {
	cfg.Network = cfg.Node.Network
	cfg.Address = cfg.Node.Address
	cfg.TLSConfig.RootCAs = []string{ctx.global.CACertificateStr()}
	conn, err := xnet.Dial(cfg.Node.Mode, &cfg.Config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	client := client{
		ctx:      ctx,
		node:     cfg.Node,
		guid:     cfg.NodeGUID,
		closeLog: cfg.CloseLog,
	}
	xconn, err := client.handshake(conn)
	if err != nil {
		_ = conn.Close()
		return nil, errors.WithMessage(err, "handshake failed")
	}
	client.conn = xconn
	// init slot
	client.slots = make([]*protocol.Slot, protocol.SlotSize)
	for i := 0; i < protocol.SlotSize; i++ {
		s := &protocol.Slot{
			Available: make(chan struct{}, 1),
			Reply:     make(chan []byte, 1),
			Timer:     time.NewTimer(protocol.RecvTimeout),
		}
		s.Available <- struct{}{}
		client.slots[i] = s
	}
	client.heartbeatC = make(chan struct{}, 1)
	client.replyTimer = time.NewTimer(time.Second)
	client.stopSignal = make(chan struct{})
	// default
	if cfg.MsgHandler == nil {
		client.wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err := xpanic.Error("client panic:", r)
					client.log(logger.FATAL, err)
				}
				client.Close()
				client.wg.Done()
			}()
			protocol.HandleConn(client.conn, client.handleMessage)
		}()
	}
	client.wg.Add(1)
	go client.heartbeat()
	return &client, nil
}

func (client *client) Close() {
	client.closeOnce.Do(func() {
		atomic.StoreInt32(&client.inClose, 1)
		close(client.stopSignal)
		_ = client.conn.Close()
		client.wg.Wait()
		if client.closeLog {
			client.log(logger.INFO, "disconnected")
		}
	})
}

func (client *client) isClosed() bool {
	return atomic.LoadInt32(&client.inClose) != 0
}

func (client *client) logf(l logger.Level, format string, log ...interface{}) {
	b := logger.Conn(client.conn)
	_, _ = fmt.Fprintf(b, format, log...)
	client.ctx.Print(l, "client", b)
}

func (client *client) log(l logger.Level, log ...interface{}) {
	b := logger.Conn(client.conn)
	_, _ = fmt.Fprint(b, log...)
	client.ctx.Print(l, "client", b)
}

func (client *client) logln(l logger.Level, log ...interface{}) {
	b := logger.Conn(client.conn)
	_, _ = fmt.Fprintln(b, log...)
	client.ctx.Print(l, "client", b)
}

// can use client.Close()
func (client *client) handleMessage(msg []byte) {
	if client.isClosed() {
		return
	}
	// cmd(1) + msg id(2) or reply
	if len(msg) < protocol.MsgCMDSize+protocol.MsgIDSize {
		client.log(logger.EXPLOIT, protocol.ErrInvalidMsgSize)
		client.Close()
		return
	}
	switch msg[0] {
	case protocol.NodeReply:
		client.handleReply(msg[protocol.MsgCMDSize:])
	case protocol.NodeHeartbeat:
		client.heartbeatC <- struct{}{}
	case protocol.ErrNullMsg:
		client.log(logger.EXPLOIT, protocol.ErrRecvNullMsg)
		client.Close()
	case protocol.ErrTooBigMsg:
		client.log(logger.EXPLOIT, protocol.ErrRecvTooBigMsg)
		client.Close()
	case protocol.TestMessage:
		client.Reply(msg[protocol.MsgCMDSize:protocol.MsgCMDSize+protocol.MsgIDSize],
			msg[protocol.MsgCMDSize+protocol.MsgIDSize:])
	default:
		client.log(logger.EXPLOIT, protocol.ErrRecvUnknownCMD, msg[protocol.MsgCMDSize:])
		client.Close()
		return
	}
}

func (client *client) heartbeat() {
	defer client.wg.Done()
	var err error
	rand := random.New(0)
	buffer := bytes.NewBuffer(nil)
	for {
		t := time.Duration(30+rand.Int(60)) * time.Second
		select {
		case <-time.After(t):
			// <security> fake flow like client
			fakeSize := 64 + rand.Int(256)
			// size(4 Bytes) + heartbeat(1 byte) + fake data
			buffer.Reset()
			buffer.Write(convert.Uint32ToBytes(uint32(1 + fakeSize)))
			buffer.WriteByte(protocol.CtrlHeartbeat)
			buffer.Write(rand.Bytes(fakeSize))
			// send
			_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
			_, err = client.conn.Write(buffer.Bytes())
			if err != nil {
				return
			}
			select {
			case <-client.heartbeatC:
			case <-time.After(t):
				_ = client.conn.Close()
				return
			case <-client.stopSignal:
				return
			}
		case <-client.stopSignal:
			return
		}
	}
}

func (client *client) Reply(id, reply []byte) {
	if client.isClosed() {
		return
	}
	l := len(reply)
	// 7 = size(4 Bytes) + NodeReply(1 byte) + msg id(2 bytes)
	b := make([]byte, protocol.MsgHeaderSize+l)
	// write size
	msgSize := protocol.MsgCMDSize + protocol.MsgIDSize + l
	copy(b, convert.Uint32ToBytes(uint32(msgSize)))
	// write cmd
	b[protocol.MsgLenSize] = protocol.NodeReply
	// write msg id
	copy(b[protocol.MsgLenSize+1:protocol.MsgLenSize+1+protocol.MsgIDSize], id)
	// write data
	copy(b[protocol.MsgHeaderSize:], reply)
	_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, _ = client.conn.Write(b)
}

// msg id(2 bytes) + data
func (client *client) handleReply(reply []byte) {
	l := len(reply)
	if l < protocol.MsgIDSize {
		client.log(logger.EXPLOIT, protocol.ErrRecvInvalidMsgIDSize)
		client.Close()
		return
	}
	id := int(convert.BytesToUint16(reply[:protocol.MsgIDSize]))
	if id > protocol.MaxMsgID {
		client.log(logger.EXPLOIT, protocol.ErrRecvInvalidMsgID)
		client.Close()
		return
	}
	// must copy
	r := make([]byte, l-protocol.MsgIDSize)
	copy(r, reply[protocol.MsgIDSize:])
	// <security> maybe wrong msg id
	client.replyTimer.Reset(time.Second)
	select {
	case client.slots[id].Reply <- r:
		client.replyTimer.Stop()
	case <-client.replyTimer.C:
		client.log(logger.EXPLOIT, protocol.ErrRecvInvalidReply)
		client.Close()
	}
}

// send command and receive reply
// size(4 Bytes) + command(1 Byte) + msg_id(2 bytes) + data
func (client *client) Send(cmd uint8, data []byte) ([]byte, error) {
	if client.isClosed() {
		return nil, protocol.ErrConnClosed
	}
	for {
		for id := 0; id < protocol.SlotSize; id++ {
			select {
			case <-client.slots[id].Available:
				l := len(data)
				b := make([]byte, protocol.MsgHeaderSize+l)
				// write MsgLen
				msgSize := protocol.MsgCMDSize + protocol.MsgIDSize + l
				copy(b, convert.Uint32ToBytes(uint32(msgSize)))
				// write cmd
				b[protocol.MsgLenSize] = cmd
				// write msg id
				copy(b[protocol.MsgLenSize+1:protocol.MsgLenSize+1+protocol.MsgIDSize],
					convert.Uint16ToBytes(uint16(id)))
				// write data
				copy(b[protocol.MsgHeaderSize:], data)
				// send
				_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
				_, err := client.conn.Write(b)
				if err != nil {
					return nil, err
				}
				// wait for reply
				client.slots[id].Timer.Reset(protocol.RecvTimeout)
				select {
				case r := <-client.slots[id].Reply:
					client.slots[id].Timer.Stop()
					client.slots[id].Available <- struct{}{}
					return r, nil
				case <-client.slots[id].Timer.C:
					client.Close()
					return nil, protocol.ErrRecvTimeout
				case <-client.stopSignal:
					return nil, protocol.ErrConnClosed
				}
			case <-client.stopSignal:
				return nil, protocol.ErrConnClosed
			default:
			}
		}
		// if full wait 1 second
		select {
		case <-time.After(time.Second):
		case <-client.stopSignal:
			return nil, protocol.ErrConnClosed
		}
	}
}
