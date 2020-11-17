package node

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/xnet"
)

const (
	connUsageServeCtrl = iota
	connUsageServeNode
	connUsageServeBeacon
	connUsageClient
)

type conn struct {
	ctx *Node

	*xnet.Conn
	role protocol.Role
	// usually is role GUID, when role = Ctrl
	// guid is CtrlConn connection GUID
	guid  *guid.GUID
	usage int

	slots []*protocol.Slot

	// for log about role GUID
	guidLine string
	logSrc   string

	// only serve role
	heartbeat *bytes.Buffer
	rand      *random.Rand

	// cache
	nodeGUID guid.GUID

	// user will initialize it in role.Sync()
	SendPool   sync.Pool
	AckPool    sync.Pool
	AnswerPool sync.Pool
	QueryPool  sync.Pool

	inClose    int32
	stopSignal chan struct{}
	closeOnce  sync.Once
}

func newConn(ctx *Node, xConn *xnet.Conn, guid *guid.GUID, usage int) *conn {
	conn := conn{
		ctx:        ctx,
		Conn:       xConn,
		guid:       guid,
		usage:      usage,
		stopSignal: make(chan struct{}),
	}
	_ = xConn.SetDeadline(time.Time{})
	// initialize message slots
	conn.slots = protocol.NewSlots()
	switch usage {
	case connUsageServeCtrl:
		conn.role = protocol.Ctrl
		conn.logSrc = "serve-ctrl"
	case connUsageServeNode:
		conn.role = protocol.Node
		conn.guidLine = "----------------------connected node guid-----------------------"
		conn.logSrc = "serve-node"
	case connUsageServeBeacon:
		conn.role = protocol.Beacon
		conn.guidLine = "---------------------connected beacon guid----------------------"
		conn.logSrc = "serve-beacon"
	case connUsageClient:
		conn.role = protocol.Node
		conn.guidLine = "----------------------connected node guid-----------------------"
		conn.logSrc = "client"
	default:
		panic(fmt.Sprintf("conn: invalid conn usage: %d", usage))
	}
	if usage != connUsageServeCtrl {
		conn.guidLine += "\n%s\n"
	}
	// only serve role handle heartbeat
	if usage != connUsageClient {
		conn.heartbeat = new(bytes.Buffer)
		conn.rand = random.NewRand()
	}
	// cache
	if usage != connUsageServeBeacon {
		conn.nodeGUID = *ctx.global.GUID()
	}
	return &conn
}

// [2019-12-26 21:44:17] [info] <client> disconnected
// ----------------------connected node guid-----------------------
// 4DAC6511AA1B6FA002C1741774ADB65A00953EA8000000005E6C6A2F001B3BC7
// -----------------------connection status------------------------
// local:  tcp 127.0.0.1:2035
// remote: tcp 127.0.0.1:2032
// sent:   5.656 MB received: 5.379 MB
// mode:   tls,  default network: tcp
// connect time: 2019-12-26 21:44:13
// ----------------------------------------------------------------
func (c *conn) Logf(lv logger.Level, format string, log ...interface{}) {
	output := new(bytes.Buffer)
	_, _ = fmt.Fprintf(output, format+"\n", log...)
	c.logExtra(lv, output)
}

func (c *conn) Log(lv logger.Level, log ...interface{}) {
	output := new(bytes.Buffer)
	_, _ = fmt.Fprintln(output, log...)
	c.logExtra(lv, output)
}

func (c *conn) logExtra(lv logger.Level, buf *bytes.Buffer) {
	if c.role != protocol.Ctrl && *c.guid != *protocol.CtrlGUID {
		_, _ = fmt.Fprintf(buf, c.guidLine, c.guid.Hex())
	}
	const conn = "-----------------------connection status------------------------\n%s\n"
	_, _ = fmt.Fprintf(buf, conn, c)
	const endLine = "----------------------------------------------------------------"
	_, _ = fmt.Fprint(buf, endLine)
	c.ctx.logger.Print(lv, c.logSrc, buf)
}

func (c *conn) isClosed() bool {
	return atomic.LoadInt32(&c.inClose) != 0
}

func (c *conn) onFrame(frame []byte) bool {
	if c.isClosed() {
		return true
	}
	// cmd(1) + msg id(2)
	if len(frame) < protocol.FrameCMDSize+protocol.FrameIDSize {
		c.Log(logger.Exploit, protocol.ErrInvalidFrameSize)
		_ = c.Close()
		return true
	}
	switch frame[0] {
	case protocol.ConnReply:
		c.handleReply(frame[protocol.FrameCMDSize:])
	case protocol.ConnGetAddress:
		id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
		address := nettool.EncodeExternalAddress(c.Conn.RemoteAddr().String())
		c.Reply(id, address)
	case protocol.ConnErrRecvNullFrame:
		c.Log(logger.Exploit, protocol.ErrRecvNullFrame)
		_ = c.Close()
	case protocol.ConnErrRecvTooLargeFrame:
		c.Log(logger.Exploit, protocol.ErrRecvTooLargeFrame)
		_ = c.Close()
	case protocol.TestCommand:
		id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
		data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
		c.Reply(id, data)
	default:
		return false
	}
	return true
}

// msg id(2 bytes) + data
func (c *conn) handleReply(reply []byte) {
	l := len(reply)
	if l < protocol.FrameIDSize {
		c.Log(logger.Exploit, protocol.ErrRecvInvalidFrameIDSize)
		_ = c.Close()
		return
	}
	id := int(convert.BEBytesToUint16(reply[:protocol.FrameIDSize]))
	if id > protocol.MaxFrameID {
		c.Log(logger.Exploit, protocol.ErrRecvInvalidFrameID)
		_ = c.Close()
		return
	}
	// must copy
	r := make([]byte, l-protocol.FrameIDSize)
	copy(r, reply[protocol.FrameIDSize:])
	// <security> maybe incorrect msg id
	select {
	case c.slots[id].Reply <- r:
	default:
		c.Log(logger.Exploit, protocol.ErrRecvInvalidReplyID)
		_ = c.Close()
	}
}

func (c *conn) Reply(id, reply []byte) {
	if c.isClosed() {
		return
	}
	l := len(reply)
	// 7 = size(4 Bytes) + ConnReply(1 byte) + msg id(2 bytes)
	b := make([]byte, protocol.FrameHeaderSize+l)
	// write size
	msgSize := protocol.FrameCMDSize + protocol.FrameIDSize + l
	copy(b, convert.BEUint32ToBytes(uint32(msgSize)))
	// write cmd
	b[protocol.FrameLenSize] = protocol.ConnReply
	// write msg id
	copy(b[protocol.FrameLenSize+1:protocol.FrameLenSize+1+protocol.FrameIDSize], id)
	// write data
	copy(b[protocol.FrameHeaderSize:], reply)
	_ = c.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, _ = c.Write(b)
}

func (c *conn) HandleHeartbeat() {
	// <security> fake traffic like client
	fakeSize := 64 + c.rand.Int(256)
	// size(4 Bytes) + heartbeat(1 byte) + fake data
	c.heartbeat.Reset()
	c.heartbeat.Write(convert.BEUint32ToBytes(uint32(1 + fakeSize)))
	c.heartbeat.WriteByte(protocol.ConnReplyHeartbeat)
	c.heartbeat.Write(c.rand.Bytes(fakeSize))
	// send heartbeat data
	_ = c.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, _ = c.Write(c.heartbeat.Bytes())
}

func (c *conn) logExploitGUID(log string, id []byte) {
	c.Log(logger.Exploit, log)
	c.Reply(id, protocol.ReplyHandled)
	_ = c.Close()
}

func (c *conn) HandleSendToNodeGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid send to node guid size", id)
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckSendToNodeGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleSendToBeaconGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid send to beacon guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckSendToBeaconGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAckToNodeGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid ack to node guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAckToNodeGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAckToBeaconGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid ack to beacon guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAckToBeaconGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBroadcastGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid broadcast guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBroadcastGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAnswerGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid answer guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAnswerGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleNodeSendGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid node send guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckNodeSendGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleNodeAckGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid node ack guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckNodeAckGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBeaconSendGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid beacon send guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBeaconSendGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBeaconAckGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid beacon ack guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBeaconAckGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleQueryGUID(id, data []byte) {
	if len(data) != guid.Size {
		c.logExploitGUID("invalid query guid size", id)
		return
	}
	if c.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckQueryGUIDSlice(data) {
		c.Reply(id, protocol.ReplyUnhandled)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) logExploit(log string, err error, obj interface{}) {
	c.Logf(logger.Exploit, log+": %s\n%s", err, spew.Sdump(obj))
	_ = c.Close()
}

func (c *conn) logfExploit(format string, obj interface{}) {
	c.Logf(logger.Exploit, format+"\n%s", spew.Sdump(obj))
	_ = c.Close()
}

func (c *conn) HandleSendToNode(id, data []byte) {
	send := c.ctx.worker.GetSendFromPool()
	put := true
	defer func() {
		if put {
			c.ctx.worker.PutSendToPool(send)
		}
	}()
	err := send.Unpack(data)
	if err != nil {
		c.logExploit("invalid send to node data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		c.logExploit("invalid send to node", err, send)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckSendToNodeGUID(&send.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		if send.RoleGUID == c.nodeGUID {
			c.ctx.worker.AddSend(send)
			put = false
		} else {
			c.ctx.forwarder.SendToNode(&send.GUID, data, c.guid)
		}
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAckToNode(id, data []byte) {
	ack := c.ctx.worker.GetAcknowledgeFromPool()
	put := true
	defer func() {
		if put {
			c.ctx.worker.PutAcknowledgeToPool(ack)
		}
	}()
	err := ack.Unpack(data)
	if err != nil {
		c.logExploit("invalid ack to node data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		c.logExploit("invalid ack to node", err, ack)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAckToNodeGUID(&ack.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		if ack.RoleGUID == c.nodeGUID {
			c.ctx.worker.AddAcknowledge(ack)
			put = false
		} else {
			c.ctx.forwarder.AckToNode(&ack.GUID, data, c.guid)
		}
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleSendToBeacon(id, data []byte) {
	send := c.SendPool.Get().(*protocol.Send)
	defer c.SendPool.Put(send)
	err := send.Unpack(data)
	if err != nil {
		c.logExploit("invalid send to beacon data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		c.logExploit("invalid send to beacon", err, send)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckSendToBeaconGUID(&send.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.SendToBeacon(&send.RoleGUID, &send.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAckToBeacon(id, data []byte) {
	ack := c.AckPool.Get().(*protocol.Acknowledge)
	defer c.AckPool.Put(ack)
	err := ack.Unpack(data)
	if err != nil {
		c.logExploit("invalid ack to beacon data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		c.logExploit("invalid ack to beacon", err, ack)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAckToBeaconGUID(&ack.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.AckToBeacon(&ack.RoleGUID, &ack.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBroadcast(id, data []byte) {
	broadcast := c.ctx.worker.GetBroadcastFromPool()
	put := true
	defer func() {
		if put {
			c.ctx.worker.PutBroadcastToPool(broadcast)
		}
	}()
	err := broadcast.Unpack(data)
	if err != nil {
		c.logExploit("invalid broadcast data", err, broadcast)
		return
	}
	err = broadcast.Validate()
	if err != nil {
		c.logExploit("invalid broadcast", err, broadcast)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&broadcast.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBroadcastGUID(&broadcast.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.Broadcast(&broadcast.GUID, data, c.guid)
		c.ctx.worker.AddBroadcast(broadcast)
		put = false
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleAnswer(id, data []byte) {
	answer := c.AnswerPool.Get().(*protocol.Answer)
	defer c.AnswerPool.Put(answer)
	err := answer.Unpack(data)
	if err != nil {
		c.logExploit("invalid answer data", err, answer)
		return
	}
	err = answer.Validate()
	if err != nil {
		c.logExploit("invalid answer", err, answer)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&answer.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckAnswerGUID(&answer.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.Answer(&answer.BeaconGUID, &answer.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleNodeSend(id, data []byte) {
	send := c.SendPool.Get().(*protocol.Send)
	defer c.SendPool.Put(send)
	err := send.Unpack(data)
	if err != nil {
		c.logExploit("invalid node send data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		c.logExploit("invalid node send", err, send)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckNodeSendGUID(&send.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.NodeSend(&send.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleNodeAck(id, data []byte) {
	ack := c.AckPool.Get().(*protocol.Acknowledge)
	defer c.AckPool.Put(ack)
	err := ack.Unpack(data)
	if err != nil {
		c.logExploit("invalid node ack data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		c.logExploit("invalid node ack", err, ack)
		return
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckNodeAckGUID(&ack.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.NodeAck(&ack.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBeaconSend(id, data []byte) {
	send := c.SendPool.Get().(*protocol.Send)
	defer c.SendPool.Put(send)
	err := send.Unpack(data)
	if err != nil {
		c.logExploit("invalid beacon send data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		c.logExploit("invalid beacon send", err, send)
		return
	}
	if c.usage == connUsageServeBeacon {
		if send.RoleGUID != *c.guid {
			c.logfExploit("different beacon guid in beacon send", send)
			return
		}
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBeaconSendGUID(&send.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.BeaconSend(&send.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleBeaconAck(id, data []byte) {
	ack := c.AckPool.Get().(*protocol.Acknowledge)
	defer c.AckPool.Put(ack)
	err := ack.Unpack(data)
	if err != nil {
		c.logExploit("invalid beacon ack data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		c.logExploit("invalid beacon ack", err, ack)
		return
	}
	if c.usage == connUsageServeBeacon {
		if ack.RoleGUID != *c.guid {
			c.logfExploit("different beacon guid in beacon ack", ack)
			return
		}
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckBeaconAckGUID(&ack.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.BeaconAck(&ack.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

func (c *conn) HandleQuery(id, data []byte) {
	query := c.QueryPool.Get().(*protocol.Query)
	defer c.QueryPool.Put(query)
	err := query.Unpack(data)
	if err != nil {
		c.logExploit("invalid query data", err, query)
		return
	}
	err = query.Validate()
	if err != nil {
		c.logExploit("invalid query", err, query)
		return
	}
	if c.usage == connUsageServeBeacon {
		if query.BeaconGUID != *c.guid {
			c.logfExploit("different beacon guid in query", query)
			return
		}
	}
	expired, timestamp := c.ctx.syncer.CheckGUIDTimestamp(&query.GUID)
	if expired {
		c.Reply(id, protocol.ReplyExpired)
		return
	}
	if c.ctx.syncer.CheckQueryGUID(&query.GUID, timestamp) {
		c.Reply(id, protocol.ReplySucceed)
		c.ctx.forwarder.Query(&query.GUID, data, c.guid)
	} else {
		c.Reply(id, protocol.ReplyHandled)
	}
}

// send is used to send command and receive reply
func (c *conn) send(cmd uint8, data []byte) ([]byte, error) {
	if c.isClosed() {
		return nil, protocol.ErrConnClosed
	}
	for {
		// select available slot
		for id := 0; id < protocol.SlotSize; id++ {
			select {
			case <-c.slots[id].Available:
				return c.sendAndWaitReply(cmd, data, id)
			case <-c.stopSignal:
				return nil, protocol.ErrConnClosed
			default:
			}
		}
		// if full wait 1 second
		select {
		case <-time.After(time.Second):
		case <-c.stopSignal:
			return nil, protocol.ErrConnClosed
		}
	}
}

func (c *conn) sendAndWaitReply(cmd uint8, data []byte, id int) ([]byte, error) {
	l := len(data)
	b := make([]byte, protocol.FrameHeaderSize+l)
	// write MsgLen
	msgSize := protocol.FrameCMDSize + protocol.FrameIDSize + l
	copy(b, convert.BEUint32ToBytes(uint32(msgSize)))
	// write cmd
	b[protocol.FrameLenSize] = cmd
	// write msg id
	copy(b[protocol.FrameLenSize+1:protocol.FrameLenSize+1+protocol.FrameIDSize],
		convert.BEUint16ToBytes(uint16(id)))
	// write data
	copy(b[protocol.FrameHeaderSize:], data)
	// send
	_ = c.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, err := c.Write(b)
	if err != nil {
		return nil, err
	}
	// wait for reply
	c.slots[id].Timer.Reset(protocol.RecvTimeout)
	select {
	case r := <-c.slots[id].Reply:
		c.slots[id].Timer.Stop()
		c.slots[id].Available <- struct{}{}
		return r, nil
	case <-c.slots[id].Timer.C:
		_ = c.Close()
		return nil, protocol.ErrRecvReplyTimeout
	case <-c.stopSignal:
		return nil, protocol.ErrConnClosed
	}
}

// SendMessage is a wrapper of xnet.Conn.Send()
func (c *conn) SendMessage(msg []byte) error {
	return c.Conn.Send(msg)
}

// SendCommand is used to send command and receive reply
func (c *conn) SendCommand(cmd uint8, data []byte) ([]byte, error) {
	return c.send(cmd, data)
}

// Send is used to send message to Controller
func (c *conn) Send(
	guid *guid.GUID,
	data *bytes.Buffer,
) (sr *protocol.SendResponse) {
	sr = &protocol.SendResponse{
		Role: c.role,
		GUID: c.guid,
	}
	var reply []byte
	reply, sr.Err = c.send(protocol.NodeSendGUID, guid[:])
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		sr.Err = protocol.GetReplyError(reply)
		return
	}
	reply, sr.Err = c.send(protocol.NodeSend, data.Bytes())
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		sr.Err = errors.New(string(reply))
	}
	return
}

// Acknowledge is used to notice Controller that Node has received this message
func (c *conn) Acknowledge(
	guid *guid.GUID,
	data *bytes.Buffer,
) (ar *protocol.AcknowledgeResponse) {
	ar = &protocol.AcknowledgeResponse{
		Role: c.role,
		GUID: c.guid,
	}
	var reply []byte
	reply, ar.Err = c.send(protocol.NodeAckGUID, guid[:])
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		ar.Err = protocol.GetReplyError(reply)
		return
	}
	reply, ar.Err = c.send(protocol.NodeAck, data.Bytes())
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		ar.Err = errors.New(string(reply))
	}
	return
}

// -------------------------------------------forwarder----------------------------------------------

// SendToNode is used to forward Controller send message to Node
func (c *conn) SendToNode(guid, data []byte) {
	reply, err := c.send(protocol.CtrlSendToNodeGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlSendToNode, data)
}

// AckToNode is used to forward Controller acknowledge to Node
func (c *conn) AckToNode(guid, data []byte) {
	reply, err := c.send(protocol.CtrlAckToNodeGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlAckToNode, data)
}

// SendToBeacon is used to forward Controller send message to Beacon
func (c *conn) SendToBeacon(guid, data []byte) {
	reply, err := c.send(protocol.CtrlSendToBeaconGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlSendToBeacon, data)
}

// AckToBeacon is used to forward Controller acknowledge to Beacon
func (c *conn) AckToBeacon(guid, data []byte) {
	reply, err := c.send(protocol.CtrlAckToBeaconGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlAckToBeacon, data)
}

// Broadcast is used to forward Controller broadcast message to Nodes
func (c *conn) Broadcast(guid, data []byte) {
	reply, err := c.send(protocol.CtrlBroadcastGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlBroadcast, data)
}

// Answer is used to forward Controller answer to Beacon
func (c *conn) Answer(guid, data []byte) {
	reply, err := c.send(protocol.CtrlAnswerGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.CtrlAnswer, data)
}

// NodeSend is used to forward Node send
func (c *conn) NodeSend(guid, data []byte) {
	reply, err := c.send(protocol.NodeSendGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.NodeSend, data)
}

// NodeAck is used to forward Node acknowledge
func (c *conn) NodeAck(guid, data []byte) {
	reply, err := c.send(protocol.NodeAckGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.NodeAck, data)
}

// BeaconSend is used to forward Beacon send
func (c *conn) BeaconSend(guid, data []byte) {
	reply, err := c.send(protocol.BeaconSendGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.BeaconSend, data)
}

// BeaconAck is used to forward Beacon acknowledge
func (c *conn) BeaconAck(guid, data []byte) {
	reply, err := c.send(protocol.BeaconAckGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.BeaconAck, data)
}

// Query is used to forward Beacon query
func (c *conn) Query(guid, data []byte) {
	reply, err := c.send(protocol.BeaconQueryGUID, guid)
	if err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	_, _ = c.send(protocol.BeaconQuery, data)
}

func (c *conn) Close() (err error) {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.inClose, 1)
		err = c.Conn.Close()
		close(c.stopSignal)
		protocol.DestroySlots(c.slots)
	})
	return
}
