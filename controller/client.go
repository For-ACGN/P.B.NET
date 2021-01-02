package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/convert"
	"project/internal/crypto/rand"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/xnet"
	"project/internal/xpanic"
)

// Client is used to connect Node's listener.
type Client struct {
	ctx *Ctrl

	listener  *bootstrap.Listener
	guid      *guid.GUID // Node GUID
	closeFunc func()

	tag       *guid.GUID
	conn      *xnet.Conn
	rand      *random.Rand
	slots     []*protocol.Slot
	heartbeat chan struct{}
	inSync    int32
	syncMu    sync.Mutex

	inClose    int32
	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// NewClient is used to create a client and connect Node listener.
// when GUID == nil      for trust node
// when GUID != CtrlGUID for sender client
// when GUID == CtrlGUID for discovery
func (ctrl *Ctrl) NewClient(
	ctx context.Context,
	listener *bootstrap.Listener,
	guid *guid.GUID,
	closeFunc func(),
) (*Client, error) {
	listener = listener.Decrypt()
	// set tls config
	host, port, err := net.SplitHostPort(listener.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig, err := ctrl.clientMgr.GetTLSConfig().Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig.Rand = rand.Reader
	tlsConfig.Time = ctrl.global.Now
	tlsConfig.ServerName = host
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}
	// set xnet options
	opts := xnet.Options{
		TLSConfig: tlsConfig,
		Timeout:   ctrl.clientMgr.GetTimeout(),
		Now:       ctrl.global.Now,
	}
	// set proxy
	proxy, err := ctrl.global.ProxyPool.Get(ctrl.clientMgr.GetProxyTag())
	if err != nil {
		return nil, err
	}
	opts.DialContext = proxy.DialContext
	// resolve domain name
	dnsOpts := ctrl.clientMgr.GetDNSOptions()
	result, err := ctrl.global.DNSClient.ResolveContext(ctx, host, dnsOpts)
	if err != nil {
		return nil, err
	}
	// dial
	var conn *xnet.Conn
	for i := 0; i < len(result); i++ {
		address := net.JoinHostPort(result[i], port)
		conn, err = xnet.DialContext(ctx, listener.Mode, listener.Network, address, &opts)
		if err == nil {
			break
		}
	}
	if err != nil {
		const format = "failed to connect node listener %s, because %s"
		return nil, errors.Errorf(format, listener, err)
	}
	// handshake
	client := &Client{
		ctx:       ctrl,
		listener:  listener,
		guid:      guid,
		conn:      conn,
		closeFunc: closeFunc,
		rand:      random.NewRand(),
	}
	err = client.handshake(ctx, conn)
	if err != nil {
		_ = conn.Close()
		const format = "failed to handshake with node listener: %s"
		return nil, errors.WithMessagef(err, format, listener)
	}
	// initialize message slots
	client.slots = protocol.NewSlots()
	client.stopSignal = make(chan struct{})
	// heartbeat
	client.heartbeat = make(chan struct{}, 1)
	client.wg.Add(1)
	go client.sendHeartbeatLoop()
	// handle connection
	// <warning> don't add wg
	go func() {
		defer func() {
			if r := recover(); r != nil {
				client.log(logger.Fatal, xpanic.Print(r, "client.HandleConn"))
			}
			client.Close()
		}()
		protocol.HandleConn(client.conn, client.onFrame)
	}()
	ctrl.clientMgr.Add(client)
	client.log(logger.Info, "connected")
	return client, nil
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
func (client *Client) logf(lv logger.Level, format string, log ...interface{}) {
	output := new(bytes.Buffer)
	_, _ = fmt.Fprintf(output, format+"\n", log...)
	client.logExtra(lv, output)
}

func (client *Client) log(lv logger.Level, log ...interface{}) {
	output := new(bytes.Buffer)
	_, _ = fmt.Fprintln(output, log...)
	client.logExtra(lv, output)
}

func (client *Client) logExtra(lv logger.Level, buf *bytes.Buffer) {
	if client.guid != nil {
		const format = "----------------------connected node guid-----------------------\n%s\n"
		_, _ = fmt.Fprintf(buf, format, client.guid.Hex())
	}
	const conn = "-----------------------connection status------------------------\n%s\n"
	_, _ = fmt.Fprintf(buf, conn, client.conn)
	const endLine = "----------------------------------------------------------------"
	_, _ = fmt.Fprint(buf, endLine)
	client.ctx.logger.Print(lv, "client", buf)
}

// Zero—Knowledge Proof
func (client *Client) handshake(ctx context.Context, conn *xnet.Conn) error {
	// interrupt
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				client.log(logger.Fatal, xpanic.Print(r, "Client.handshake"))
			}
			wg.Done()
		}()
		select {
		case <-done:
		case <-ctx.Done():
			client.Close()
		}
	}()
	defer func() {
		close(done)
		wg.Wait()
	}()
	_ = conn.SetDeadline(time.Now().Add(client.ctx.clientMgr.GetTimeout()))
	// check connection
	err := client.checkConn(conn)
	if err != nil {
		return err
	}
	// verify certificate
	publicKey := client.ctx.global.PublicKey()
	_, ok, err := protocol.VerifyCertificate(conn, publicKey, client.guid)
	if err != nil {
		client.log(logger.Exploit, err)
		return err
	}
	if !ok {
		return errors.New("failed to verify node certificate")
	}
	// send role
	_, err = conn.Write(protocol.Ctrl.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to send role")
	}
	// receive challenge
	challenge, err := conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive challenge")
	}
	// <danger>
	// maybe fake node will send some special data
	// and if controller sign it will destroy net
	if len(challenge) != protocol.ChallengeSize {
		err = errors.New("invalid challenge size")
		client.log(logger.Exploit, err)
		return err
	}
	// send signature
	err = conn.Send(client.ctx.global.Sign(challenge))
	if err != nil {
		return errors.Wrap(err, "failed to send challenge signature")
	}
	resp, err := conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive authentication response")
	}
	if !bytes.Equal(resp, protocol.AuthSucceed) {
		err = errors.WithStack(protocol.ErrAuthenticateFailed)
		client.log(logger.Exploit, err)
		return err
	}
	return conn.SetDeadline(time.Time{})
}

// TODO improve it, server side use "white" method
func (client *Client) checkConn(conn *xnet.Conn) error {
	size := byte(100 + client.rand.Int(156))
	data := client.rand.Bytes(int(size))
	_, err := conn.Write(append([]byte{size}, data...))
	if err != nil {
		return errors.WithMessage(err, "failed to send check connection data")
	}
	n, err := io.ReadFull(conn, data)
	if err != nil {
		const format = "error in client.checkConn():\n%s\n%s"
		client.logf(logger.Exploit, format, err, spew.Sdump(data[:n]))
		return err
	}
	return nil
}

func (client *Client) sendHeartbeatLoop() {
	defer client.wg.Done()
	var err error
	buffer := new(bytes.Buffer)
	sleeper := random.NewSleeper()
	defer sleeper.Stop()
	for {
		select {
		case <-sleeper.SleepSecond(30, 60):
			// <security> fake traffic like client
			fakeSize := 64 + client.rand.Int(256)
			// size(4 Bytes) + heartbeat(1 byte) + fake data
			buffer.Reset()
			buffer.Write(convert.BEUint32ToBytes(uint32(1 + fakeSize)))
			buffer.WriteByte(protocol.ConnSendHeartbeat)
			buffer.Write(client.rand.Bytes(fakeSize))
			// send
			_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
			_, err = client.conn.Write(buffer.Bytes())
			if err != nil {
				return
			}
			// receive reply
			select {
			case <-client.heartbeat:
			case <-sleeper.SleepSecond(30, 60):
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

func (client *Client) isClosed() bool {
	return atomic.LoadInt32(&client.inClose) != 0
}

func (client *Client) isSync() bool {
	return atomic.LoadInt32(&client.inSync) != 0
}

// can use client.Close()
func (client *Client) onFrame(frame []byte) {
	if client.isClosed() {
		return
	}
	// cmd(1) + msg id(2) or reply
	if len(frame) < protocol.FrameCMDSize+protocol.FrameIDSize {
		client.log(logger.Exploit, protocol.ErrInvalidFrameSize)
		client.Close()
		return
	}
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	if client.isSync() {
		if client.onFrameAfterSync(frame[0], id, data) {
			return
		}
	}
	switch frame[0] {
	case protocol.ConnReply:
		client.handleReply(frame[protocol.FrameCMDSize:])
	case protocol.ConnReplyHeartbeat:
		select {
		case client.heartbeat <- struct{}{}:
		case <-client.stopSignal:
		}
	case protocol.ConnErrRecvNullFrame:
		client.log(logger.Exploit, protocol.ErrRecvNullFrame)
		client.Close()
	case protocol.ConnErrRecvTooLargeFrame:
		client.log(logger.Exploit, protocol.ErrRecvTooLargeFrame)
		client.Close()
	case protocol.TestCommand:
		client.reply(id, data)
	default:
		const format = "unknown command: %d\nframe:\n%s"
		client.logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
		client.Close()
	}
}

func (client *Client) reply(id, reply []byte) {
	if client.isClosed() {
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
	_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, _ = client.conn.Write(b)
}

// msg id(2 bytes) + data
func (client *Client) handleReply(reply []byte) {
	l := len(reply)
	if l < protocol.FrameIDSize {
		client.log(logger.Exploit, protocol.ErrRecvInvalidFrameIDSize)
		client.Close()
		return
	}
	id := int(convert.BEBytesToUint16(reply[:protocol.FrameIDSize]))
	if id > protocol.MaxFrameID {
		client.log(logger.Exploit, protocol.ErrRecvInvalidFrameID)
		client.Close()
		return
	}
	// must copy
	r := make([]byte, l-protocol.FrameIDSize)
	copy(r, reply[protocol.FrameIDSize:])
	// <security> maybe incorrect msg id
	select {
	case client.slots[id].Reply <- r:
	default:
		client.log(logger.Exploit, protocol.ErrRecvInvalidReplyID)
		client.Close()
	}
}

// Synchronize is used to switch to synchronize mode.
func (client *Client) Synchronize() error {
	client.syncMu.Lock()
	defer client.syncMu.Unlock()
	if client.isSync() {
		return errors.New("already synchronize")
	}
	// must presume, or may be lost message.
	var err error
	atomic.StoreInt32(&client.inSync, 1)
	defer func() {
		if err != nil {
			atomic.StoreInt32(&client.inSync, 0)
		}
	}()
	resp, err := client.send(protocol.CtrlSync, nil)
	if err != nil {
		return errors.Wrap(err, "failed to receive synchronize response")
	}
	if !bytes.Equal(resp, []byte{protocol.NodeSync}) {
		err = errors.Errorf("failed to start to synchronize: %s", resp)
		return err // can't return directly
	}
	client.logf(logger.Info, "start to synchronize\nlistener: %s", client.listener)
	return nil
}

func (client *Client) onFrameAfterSync(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.NodeSendGUID:
		client.handleNodeSendGUID(id, data)
	case protocol.NodeSend:
		client.handleNodeSend(id, data)
	case protocol.NodeAckGUID:
		client.handleNodeAckGUID(id, data)
	case protocol.NodeAck:
		client.handleNodeAck(id, data)
	case protocol.BeaconSendGUID:
		client.handleBeaconSendGUID(id, data)
	case protocol.BeaconSend:
		client.handleBeaconSend(id, data)
	case protocol.BeaconAckGUID:
		client.handleBeaconAckGUID(id, data)
	case protocol.BeaconAck:
		client.handleBeaconAck(id, data)
	case protocol.BeaconQueryGUID:
		client.handleQueryGUID(id, data)
	case protocol.BeaconQuery:
		client.handleQuery(id, data)
	default:
		return false
	}
	return true
}

func (client *Client) logExploitGUID(log string, id []byte) {
	client.log(logger.Exploit, log)
	client.reply(id, protocol.ReplyHandled)
	client.Close()
}

func (client *Client) handleNodeSendGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid node send guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckNodeSendGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleNodeAckGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid node ack guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckNodeAckGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleBeaconSendGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid beacon send guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckBeaconSendGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleBeaconAckGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid beacon ack guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckBeaconAckGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleQueryGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid query guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckQueryGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) logExploit(log string, err error, obj interface{}) {
	client.logf(logger.Exploit, log+": %s\n%s", err, spew.Sdump(obj))
	client.Close()
}

func (client *Client) handleNodeSend(id, data []byte) {
	send := client.ctx.worker.GetSendFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutSendToPool(send)
		}
	}()
	err := send.Unpack(data)
	if err != nil {
		client.logExploit("invalid node send data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		client.logExploit("invalid node send", err, send)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckNodeSendGUID(&send.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		client.ctx.worker.AddNodeSend(send)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleNodeAck(id, data []byte) {
	ack := client.ctx.worker.GetAcknowledgeFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutAcknowledgeToPool(ack)
		}
	}()
	err := ack.Unpack(data)
	if err != nil {
		client.logExploit("invalid node ack data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		client.logExploit("invalid node ack", err, ack)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckNodeAckGUID(&ack.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		client.ctx.worker.AddNodeAcknowledge(ack)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleBeaconSend(id, data []byte) {
	send := client.ctx.worker.GetSendFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutSendToPool(send)
		}
	}()
	err := send.Unpack(data)
	if err != nil {
		client.logExploit("invalid beacon send data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		client.logExploit("invalid beacon send", err, send)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckBeaconSendGUID(&send.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		client.ctx.worker.AddBeaconSend(send)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleBeaconAck(id, data []byte) {
	ack := client.ctx.worker.GetAcknowledgeFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutAcknowledgeToPool(ack)
		}
	}()
	err := ack.Unpack(data)
	if err != nil {
		client.logExploit("invalid beacon ack data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		client.logExploit("invalid beacon ack", err, ack)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckBeaconAckGUID(&ack.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		client.ctx.worker.AddBeaconAcknowledge(ack)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleQuery(id, data []byte) {
	query := client.ctx.worker.GetQueryFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutQueryToPool(query)
		}
	}()
	err := query.Unpack(data)
	if err != nil {
		client.logExploit("invalid query data", err, query)
		return
	}
	err = query.Validate()
	if err != nil {
		client.logExploit("invalid query", err, query)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&query.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckQueryGUID(&query.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		client.ctx.worker.AddQuery(query)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

// send is used to send command and receive reply.
func (client *Client) send(cmd uint8, data []byte) ([]byte, error) {
	if client.isClosed() {
		return nil, protocol.ErrConnClosed
	}
	for {
		// select available slot
		for id := 0; id < protocol.SlotSize; id++ {
			select {
			case <-client.slots[id].Available:
				return client.sendAndWaitReply(cmd, data, id)
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

func (client *Client) sendAndWaitReply(cmd uint8, data []byte, id int) ([]byte, error) {
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
	_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, err := client.conn.Write(b)
	if err != nil {
		_ = client.conn.Close()
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
		return nil, protocol.ErrRecvReplyTimeout
	case <-client.stopSignal:
		return nil, protocol.ErrConnClosed
	}
}

// SendCommand is used to send command and receive reply.
func (client *Client) SendCommand(cmd uint8, data []byte) ([]byte, error) {
	return client.send(cmd, data)
}

// SendToNode is used to send message to node.
func (client *Client) SendToNode(
	guid *guid.GUID,
	data *bytes.Buffer,
) (sr *protocol.SendResponse) {
	sr = &protocol.SendResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, sr.Err = client.send(protocol.CtrlSendToNodeGUID, guid[:])
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		sr.Err = protocol.GetReplyError(reply)
		return
	}
	reply, sr.Err = client.send(protocol.CtrlSendToNode, data.Bytes())
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		sr.Err = protocol.GetReplyError(reply)
	}
	return
}

// SendToBeacon is used to send message to beacon.
func (client *Client) SendToBeacon(
	guid *guid.GUID,
	data *bytes.Buffer,
) (sr *protocol.SendResponse) {
	sr = &protocol.SendResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, sr.Err = client.send(protocol.CtrlSendToBeaconGUID, guid[:])
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		sr.Err = protocol.GetReplyError(reply)
		return
	}
	reply, sr.Err = client.send(protocol.CtrlSendToBeacon, data.Bytes())
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		sr.Err = protocol.GetReplyError(reply)
	}
	return
}

// AckToNode is used to notice Node that Controller has received this message.
func (client *Client) AckToNode(
	guid *guid.GUID,
	data *bytes.Buffer,
) (ar *protocol.AcknowledgeResponse) {
	ar = &protocol.AcknowledgeResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, ar.Err = client.send(protocol.CtrlAckToNodeGUID, guid[:])
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	reply, ar.Err = client.send(protocol.CtrlAckToNode, data.Bytes())
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		ar.Err = errors.New(string(reply))
	}
	return
}

// AckToBeacon is used to notice Beacon that Controller has received this message.
func (client *Client) AckToBeacon(
	guid *guid.GUID,
	data *bytes.Buffer,
) (ar *protocol.AcknowledgeResponse) {
	ar = &protocol.AcknowledgeResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, ar.Err = client.send(protocol.CtrlAckToBeaconGUID, guid[:])
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	reply, ar.Err = client.send(protocol.CtrlAckToBeacon, data.Bytes())
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		ar.Err = errors.New(string(reply))
	}
	return
}

// Broadcast is used to broadcast message to all Nodes.
func (client *Client) Broadcast(
	guid *guid.GUID,
	data *bytes.Buffer,
) (br *protocol.BroadcastResponse) {
	br = &protocol.BroadcastResponse{
		GUID: client.guid,
	}
	var reply []byte
	reply, br.Err = client.send(protocol.CtrlBroadcastGUID, guid[:])
	if br.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		br.Err = protocol.GetReplyError(reply)
		return
	}
	// broadcast
	reply, br.Err = client.send(protocol.CtrlBroadcast, data.Bytes())
	if br.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		br.Err = protocol.GetReplyError(reply)
	}
	return
}

// Answer is used to return the result of the Beacon query.
func (client *Client) Answer(
	guid *guid.GUID,
	data *bytes.Buffer,
) (ar *protocol.AnswerResponse) {
	ar = &protocol.AnswerResponse{
		GUID: client.guid,
	}
	var reply []byte
	reply, ar.Err = client.send(protocol.CtrlAnswerGUID, guid[:])
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		return
	}
	reply, ar.Err = client.send(protocol.CtrlAnswer, data.Bytes())
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		ar.Err = errors.New(string(reply))
	}
	return
}

// Status is used to get connection status.
func (client *Client) Status() *xnet.Status {
	return client.conn.Status()
}

// Close is used to disconnect Node.
func (client *Client) Close() {
	client.closeOnce.Do(func() {
		atomic.StoreInt32(&client.inClose, 1)
		_ = client.conn.Close()
		close(client.stopSignal)
		protocol.DestroySlots(client.slots)
		client.wg.Wait()
		client.ctx.clientMgr.Delete(client.tag)
		if client.closeFunc != nil {
			client.closeFunc()
		}
		client.log(logger.Info, "disconnected")
	})
}
