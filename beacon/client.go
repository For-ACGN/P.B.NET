package beacon

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
	ctx *Beacon

	listener  *bootstrap.Listener
	guid      *guid.GUID // Node GUID
	closeFunc func()

	tag       *guid.GUID
	Conn      *xnet.Conn
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

// NewClient is used to create a client and connect Node's listener.
// when guid != ctrl guid for forwarder
// when guid == ctrl guid for register
func (beacon *Beacon) NewClient(
	ctx context.Context,
	listener *bootstrap.Listener,
	guid *guid.GUID,
	closeFunc func(),
) (*Client, error) {
	listener = listener.Decrypt()
	host, port, err := net.SplitHostPort(listener.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// set tls config
	tlsConfig, err := beacon.clientMgr.GetTLSConfig().Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig.Rand = rand.Reader
	tlsConfig.Time = beacon.global.Now
	tlsConfig.ServerName = host
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}
	// set xnet options
	opts := xnet.Options{
		TLSConfig: tlsConfig,
		Timeout:   beacon.clientMgr.GetTimeout(),
		Now:       beacon.global.Now,
	}
	// set proxy
	proxy, err := beacon.global.ProxyPool.Get(beacon.clientMgr.GetProxyTag())
	if err != nil {
		return nil, err
	}
	opts.DialContext = proxy.DialContext
	// resolve domain name
	dnsOpts := beacon.clientMgr.GetDNSOptions()
	result, err := beacon.global.DNSClient.ResolveContext(ctx, host, dnsOpts)
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
		const format = "failed to connect node listener: %s because: %s"
		return nil, errors.Errorf(format, listener, err)
	}
	// handshake
	client := &Client{
		ctx:       beacon,
		listener:  listener,
		guid:      guid,
		Conn:      conn,
		closeFunc: closeFunc,
		rand:      random.NewRand(),
	}
	err = client.handshake(ctx, conn)
	if err != nil {
		_ = conn.Close()
		const format = "failed to handshake with node listener: %s"
		return nil, errors.WithMessagef(err, format, listener)
	}
	beacon.clientMgr.Add(client)
	client.log(logger.Debug, "create client")
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
	if *client.guid != *protocol.CtrlGUID {
		const format = "----------------------connected node guid-----------------------\n%s\n"
		_, _ = fmt.Fprintf(buf, format, client.guid.Hex())
	}
	const conn = "-----------------------connection status------------------------\n%s\n"
	_, _ = fmt.Fprintf(buf, conn, client.Conn)
	const endLine = "----------------------------------------------------------------"
	_, _ = fmt.Fprint(buf, endLine)
	client.ctx.logger.Print(lv, "client", buf)
}

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
	// about check connection
	err := client.checkConn(conn)
	if err != nil {
		return err
	}
	// verify certificate
	publicKey := client.ctx.global.CtrlPublicKey()
	cert, ok, err := protocol.VerifyCertificate(conn, publicKey, client.guid)
	if err != nil {
		if errors.Cause(err) == protocol.ErrDifferentNodeGUID {
			// TODO update node
			ok, err := client.ctx.driver.UpdateNode(ctx, cert)
			if err != nil {
				return err
			}
			if !ok {
				return err
			}
		} else {
			client.log(logger.Exploit, err)
		}
		return err
	}
	if !ok {
		return errors.New("failed to verify node certificate")
	}
	// send role
	_, err = conn.Write(protocol.Beacon.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to send role")
	}
	// send self guid
	_, err = conn.Write(client.ctx.global.GUID()[:])
	return err
}

func (client *Client) checkConn(conn *xnet.Conn) error {
	size := byte(100 + client.rand.Int(156))
	data := client.rand.Bytes(int(size))
	_, err := conn.Write(append([]byte{size}, data...))
	if err != nil {
		return errors.WithMessage(err, "failed to send check connection data")
	}
	n, err := io.ReadFull(conn, data)
	if err != nil {
		const format = "error in client.checkConn(): %s\n%s"
		client.logf(logger.Exploit, format, err, spew.Sdump(data[:n]))
		return err
	}
	return nil
}

// SetRandomDeadline is used to random set connection deadline.
func (client *Client) SetRandomDeadline(fixed, random int) {
	timeout := time.Duration(fixed+client.rand.Int(random)) * time.Second
	_ = client.Conn.SetDeadline(time.Now().Add(timeout))
}

// Authenticate is used to authenticate to Node.
// Connect and UpdateNode() need it.
func (client *Client) Authenticate() error {
	// receive challenge
	challenge, err := client.Conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive challenge")
	}
	if len(challenge) != protocol.ChallengeSize {
		err = errors.New("invalid challenge size")
		client.log(logger.Exploit, err)
		return err
	}
	// send signature
	err = client.Conn.Send(client.ctx.global.Sign(challenge))
	if err != nil {
		return errors.Wrap(err, "failed to send challenge signature")
	}
	resp, err := client.Conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive authentication response")
	}
	if !bytes.Equal(resp, protocol.AuthSucceed) {
		return errors.WithStack(protocol.ErrAuthenticateFailed)
	}
	return nil
}

// Connect is used to start protocol.HandleConn(), if you want to
// start Synchronize(), you must call this function first.
func (client *Client) Connect() error {
	// send connect operation
	_, err := client.Conn.Write([]byte{protocol.BeaconOperationConnect})
	if err != nil {
		return errors.Wrap(err, "failed to send connect operation")
	}
	err = client.Authenticate()
	if err != nil {
		return err
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
		protocol.HandleConn(client.Conn, client.onFrame)
	}()
	timeout := client.ctx.clientMgr.GetTimeout()
	_ = client.Conn.SetDeadline(time.Now().Add(timeout))
	client.log(logger.Debug, "connected")
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
		case <-sleeper.Sleep(30, 60):
			// <security> fake traffic like client
			fakeSize := 64 + client.rand.Int(256)
			// size(4 Bytes) + heartbeat(1 byte) + fake data
			buffer.Reset()
			buffer.Write(convert.BEUint32ToBytes(uint32(1 + fakeSize)))
			buffer.WriteByte(protocol.ConnSendHeartbeat)
			buffer.Write(client.rand.Bytes(fakeSize))
			// send
			_ = client.Conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
			_, err = client.Conn.Write(buffer.Bytes())
			if err != nil {
				return
			}
			// receive reply
			select {
			case <-client.heartbeat:
			case <-sleeper.Sleep(30, 60):
				client.log(logger.Error, "receive heartbeat timeout")
				_ = client.Conn.Close()
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
	_ = client.Conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, _ = client.Conn.Write(b)
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

// Synchronize is used to switch to synchronize mode
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
	resp, err := client.send(protocol.BeaconSync, nil)
	if err != nil {
		return errors.Wrap(err, "failed to receive synchronize response")
	}
	if !bytes.Equal(resp, []byte{protocol.NodeSync}) {
		err = errors.Errorf("failed to start to synchronize: %s", resp)
		return err // can't return directly
	}
	client.logf(logger.Debug, "start to synchronize\nlistener: %s", client.listener)
	return nil
}

func (client *Client) onFrameAfterSync(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.CtrlSendToBeaconGUID:
		client.handleSendToBeaconGUID(id, data)
	case protocol.CtrlSendToBeacon:
		client.handleSendToBeacon(id, data)
	case protocol.CtrlAckToBeaconGUID:
		client.handleAckToBeaconGUID(id, data)
	case protocol.CtrlAckToBeacon:
		client.handleAckToBeacon(id, data)
	case protocol.CtrlAnswerGUID:
		client.handleAnswerGUID(id, data)
	case protocol.CtrlAnswer:
		client.handleAnswer(id, data)
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

func (client *Client) handleSendToBeaconGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid send to beacon guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckSendToBeaconGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleAckToBeaconGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid ack to beacon guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckAckToBeaconGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleAnswerGUID(id, data []byte) {
	if len(data) != guid.Size {
		client.logExploitGUID("invalid answer guid size", id)
		return
	}
	if client.ctx.syncer.CheckGUIDSliceTimestamp(data) {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckAnswerGUIDSlice(data) {
		client.reply(id, protocol.ReplyUnhandled)
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) logExploit(log string, err error, obj interface{}) {
	client.logf(logger.Exploit, log+": %s\n%s", err, spew.Sdump(obj))
	client.Close()
}

func (client *Client) logfExploit(format string, obj interface{}) {
	client.logf(logger.Exploit, format+"\n%s", spew.Sdump(obj))
	client.Close()
}

func (client *Client) handleSendToBeacon(id, data []byte) {
	send := client.ctx.worker.GetSendFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutSendToPool(send)
		}
	}()
	err := send.Unpack(data)
	if err != nil {
		client.logExploit("invalid send to beacon data", err, send)
		return
	}
	err = send.Validate()
	if err != nil {
		client.logExploit("invalid send to beacon", err, send)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&send.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckSendToBeaconGUID(&send.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		if send.RoleGUID != *client.ctx.global.GUID() {
			client.logfExploit("different beacon guid in send to beacon", send)
			return
		}
		client.ctx.worker.AddSend(send)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleAckToBeacon(id, data []byte) {
	ack := client.ctx.worker.GetAcknowledgeFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutAcknowledgeToPool(ack)
		}
	}()
	err := ack.Unpack(data)
	if err != nil {
		client.logExploit("invalid ack to beacon data", err, ack)
		return
	}
	err = ack.Validate()
	if err != nil {
		client.logExploit("invalid ack to beacon", err, ack)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&ack.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckAckToBeaconGUID(&ack.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		if ack.RoleGUID != *client.ctx.global.GUID() {
			client.logfExploit("different beacon guid in ack to beacon", ack)
			return
		}
		client.ctx.worker.AddAcknowledge(ack)
		put = false
	} else {
		client.reply(id, protocol.ReplyHandled)
	}
}

func (client *Client) handleAnswer(id, data []byte) {
	answer := client.ctx.worker.GetAnswerFromPool()
	put := true
	defer func() {
		if put {
			client.ctx.worker.PutAnswerToPool(answer)
		}
	}()
	err := answer.Unpack(data)
	if err != nil {
		client.logExploit("invalid answer data", err, answer)
		return
	}
	err = answer.Validate()
	if err != nil {
		client.logExploit("invalid answer", err, answer)
		return
	}
	expired, timestamp := client.ctx.syncer.CheckGUIDTimestamp(&answer.GUID)
	if expired {
		client.reply(id, protocol.ReplyExpired)
		return
	}
	if client.ctx.syncer.CheckAnswerGUID(&answer.GUID, timestamp) {
		client.reply(id, protocol.ReplySucceed)
		if answer.BeaconGUID != *client.ctx.global.GUID() {
			client.logfExploit("different beacon guid in answer", answer)
			return
		}
		client.ctx.worker.AddAnswer(answer)
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
	_ = client.Conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
	_, err := client.Conn.Write(b)
	if err != nil {
		_ = client.Conn.Close()
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

// SendCommand is used to send command and receive reply
func (client *Client) SendCommand(cmd uint8, data []byte) ([]byte, error) {
	return client.send(cmd, data)
}

// Send is used to send message to Controller.
func (client *Client) Send(
	guid *guid.GUID,
	data *bytes.Buffer,
) (sr *protocol.SendResponse) {
	sr = &protocol.SendResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, sr.Err = client.send(protocol.BeaconSendGUID, guid[:])
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		sr.Err = protocol.GetReplyError(reply)
		return
	}
	reply, sr.Err = client.send(protocol.BeaconSend, data.Bytes())
	if sr.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		sr.Err = errors.New(string(reply))
	}
	return
}

// Acknowledge is used to notice Controller that Beacon has received this message.
func (client *Client) Acknowledge(
	guid *guid.GUID,
	data *bytes.Buffer,
) (ar *protocol.AcknowledgeResponse) {
	ar = &protocol.AcknowledgeResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, ar.Err = client.send(protocol.BeaconAckGUID, guid[:])
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		ar.Err = protocol.GetReplyError(reply)
		return
	}
	reply, ar.Err = client.send(protocol.BeaconAck, data.Bytes())
	if ar.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		ar.Err = errors.New(string(reply))
	}
	return
}

// Query is used to query message from Controller.
func (client *Client) Query(
	guid *guid.GUID,
	data *bytes.Buffer,
) (q *protocol.QueryResponse) {
	q = &protocol.QueryResponse{
		Role: protocol.Node,
		GUID: client.guid,
	}
	var reply []byte
	reply, q.Err = client.send(protocol.BeaconQueryGUID, guid[:])
	if q.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplyUnhandled) {
		q.Err = protocol.GetReplyError(reply)
		return
	}
	reply, q.Err = client.send(protocol.BeaconQuery, data.Bytes())
	if q.Err != nil {
		return
	}
	if !bytes.Equal(reply, protocol.ReplySucceed) {
		q.Err = errors.New(string(reply))
	}
	return
}

// Status is used to get connection status
func (client *Client) Status() *xnet.Status {
	return client.Conn.Status()
}

// Close is used to disconnect Node
func (client *Client) Close() {
	client.closeOnce.Do(func() {
		atomic.StoreInt32(&client.inClose, 1)
		_ = client.Conn.Close()
		if client.stopSignal != nil {
			close(client.stopSignal)
			protocol.DestroySlots(client.slots)
		}
		client.wg.Wait()
		client.ctx.clientMgr.Delete(client.tag)
		if client.closeFunc != nil {
			client.closeFunc()
		}
		if client.stopSignal != nil {
			client.log(logger.Debug, "disconnected")
		}
	})
}
