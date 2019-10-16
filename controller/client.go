package controller

import (
	"bytes"
	"fmt"
	"net"
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
type clientOpts struct {
	NodeGUID   []byte
	MsgHandler func(msg []byte)
}

type client struct {
	ctx  *CTRL
	node *bootstrap.Node
	guid []byte // node guid

	conn      *xnet.Conn
	slots     []*protocol.Slot
	heartbeat chan struct{}

	closing    int32
	closeOnce  sync.Once
	stopSignal chan struct{}
	wg         sync.WaitGroup
}

func newClient(ctx *CTRL, node *bootstrap.Node, opts *clientOpts) (*client, error) {
	xnetCfg := xnet.Config{
		Network: node.Network,
		Address: node.Address,
	}
	xnetCfg.TLSConfig.RootCAs = ctx.global.CACertificatesStr()
	conn, err := xnet.Dial(node.Mode, &xnetCfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if opts == nil {
		opts = new(clientOpts)
	}
	client := client{
		ctx:  ctx,
		node: node,
		guid: opts.NodeGUID,
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
	client.heartbeat = make(chan struct{}, 1)
	client.stopSignal = make(chan struct{})
	if opts.MsgHandler == nil {
		// <warning> not add wg
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err := xpanic.Error("client panic:", r)
					client.log(logger.Fatal, err)
				}
				client.Close()
			}()
			protocol.HandleConn(client.conn, client.handleMessage)
		}()
	}
	client.wg.Add(1)
	go client.sendHeartbeatLoop()
	return &client, nil
}

func (client *client) Status() *xnet.Status {
	return client.conn.Status()
}

func (client *client) isClosing() bool {
	return atomic.LoadInt32(&client.closing) != 0
}

func (client *client) Close() {
	client.closeOnce.Do(func() {
		atomic.StoreInt32(&client.closing, 1)
		_ = client.conn.Close()
		close(client.stopSignal)
		client.wg.Wait()
		client.log(logger.Info, "disconnected")
	})
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

func (client *client) handshake(conn net.Conn) (*xnet.Conn, error) {
	dConn := xnet.NewDeadlineConn(conn, time.Minute)
	xConn := xnet.NewConn(dConn, client.ctx.global.Now())
	// receive certificate
	cert, err := xConn.Receive()
	if err != nil {
		return nil, errors.Wrap(err, "receive certificate failed")
	}
	if !client.ctx.verifyCertificate(cert, client.node.Address, client.guid) {
		client.log(logger.Exploit, protocol.ErrInvalidCert)
		return nil, protocol.ErrInvalidCert
	}
	// send role
	_, err = xConn.Write(protocol.Ctrl.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "send role failed")
	}
	// receive challenge
	challenge, err := xConn.Receive()
	if err != nil {
		return nil, errors.Wrap(err, "receive challenge data failed")
	}
	// <danger>
	// receive random challenge data(length 2048-4096)
	// len(challenge) must > len(GUID + Mode + Network + Address)
	// because maybe fake node will send some special data
	// and if controller sign it will destroy net
	if len(challenge) < 2048 || len(challenge) > 4096 {
		err = errors.New("invalid challenge size")
		client.log(logger.Exploit, err)
		return nil, err
	}
	// send signature
	err = xConn.Send(client.ctx.global.Sign(challenge))
	if err != nil {
		return nil, errors.Wrap(err, "send challenge signature failed")
	}
	resp, err := xConn.Receive()
	if err != nil {
		return nil, errors.Wrap(err, "receive authentication response failed")
	}
	if !bytes.Equal(resp, protocol.AuthSucceed) {
		err = errors.WithStack(protocol.ErrAuthFailed)
		client.log(logger.Exploit, err)
		return nil, err
	}
	// remove deadline conn
	return xnet.NewConn(conn, client.ctx.global.Now()), nil
}

// can use client.Close()
func (client *client) handleMessage(msg []byte) {
	const (
		cmd = protocol.MsgCMDSize
		id  = protocol.MsgCMDSize + protocol.MsgIDSize
	)
	if client.isClosing() {
		return
	}
	// cmd(1) + msg id(2) or reply
	if len(msg) < id {
		client.log(logger.Exploit, protocol.ErrInvalidMsgSize)
		client.Close()
		return
	}
	switch msg[0] {
	case protocol.NodeReply:
		client.handleReply(msg[cmd:])
	case protocol.NodeHeartbeat:
		client.heartbeat <- struct{}{}
	case protocol.ErrCMDRecvNullMsg:
		client.log(logger.Exploit, protocol.ErrRecvNullMsg)
		client.Close()
	case protocol.ErrCMDTooBigMsg:
		client.log(logger.Exploit, protocol.ErrRecvTooBigMsg)
		client.Close()
	case protocol.TestCommand:
		client.Reply(msg[cmd:id], msg[id:])
	default:
		client.log(logger.Exploit, protocol.ErrRecvUnknownCMD, msg)
		client.Close()
		return
	}
}

func (client *client) sendHeartbeatLoop() {
	defer client.wg.Done()
	var err error
	rand := random.New(0)
	buffer := bytes.NewBuffer(nil)
	for {
		t := time.Duration(30+rand.Int(60)) * time.Second
		select {
		case <-time.After(t):
			// <security> fake traffic like client
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
			case <-client.heartbeat:
			case <-time.After(t):
				client.log(logger.Warning, "receive heartbeat timeout")
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

// msg id(2 bytes) + data
func (client *client) handleReply(reply []byte) {
	l := len(reply)
	if l < protocol.MsgIDSize {
		client.log(logger.Exploit, protocol.ErrRecvInvalidMsgIDSize)
		client.Close()
		return
	}
	id := int(convert.BytesToUint16(reply[:protocol.MsgIDSize]))
	if id > protocol.MaxMsgID {
		client.log(logger.Exploit, protocol.ErrRecvInvalidMsgID)
		client.Close()
		return
	}
	// must copy
	r := make([]byte, l-protocol.MsgIDSize)
	copy(r, reply[protocol.MsgIDSize:])
	// <security> maybe wrong msg id
	select {
	case client.slots[id].Reply <- r:
	default:
		client.log(logger.Exploit, protocol.ErrRecvInvalidReplyID)
		client.Close()
	}
}

func (client *client) Reply(id, reply []byte) {
	if client.isClosing() {
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

// send command and receive reply
// size(4 Bytes) + command(1 Byte) + msg_id(2 bytes) + data
// data(general) max size = MaxMsgSize -MsgCMDSize -MsgIDSize
func (client *client) Send(cmd uint8, data []byte) ([]byte, error) {
	if client.isClosing() {
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
				if !client.slots[id].Timer.Stop() {
					<-client.slots[id].Timer.C
				}
				client.slots[id].Timer.Reset(protocol.RecvTimeout)
				select {
				case r := <-client.slots[id].Reply:
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
				// try next slot
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
