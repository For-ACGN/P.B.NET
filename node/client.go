package node

import (
	"bytes"
	"context"
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
	ctx *Node

	listener *bootstrap.Listener
	GUID     *guid.GUID // Node GUID

	tag       *guid.GUID
	Conn      *conn
	rand      *random.Rand
	heartbeat chan struct{}
	inSync    int32
	syncMu    sync.Mutex

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// NewClient is used to create a client and connect node listener.
// when guid != ctrl guid for forwarder.
// when guid == ctrl guid for register.
func (node *Node) NewClient(
	ctx context.Context,
	listener *bootstrap.Listener,
	guid *guid.GUID,
) (*Client, error) {
	listener = listener.Decrypt()
	host, port, err := net.SplitHostPort(listener.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// set tls config
	tlsConfig, err := node.clientMgr.GetTLSConfig().Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsConfig.Rand = rand.Reader
	tlsConfig.Time = node.global.Now
	tlsConfig.ServerName = host
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}
	// set xnet options
	opts := xnet.Options{
		TLSConfig: tlsConfig,
		Timeout:   node.clientMgr.GetTimeout(),
		Now:       node.global.Now,
	}
	// set proxy
	proxy, err := node.global.ProxyPool.Get(node.clientMgr.GetProxyTag())
	if err != nil {
		return nil, err
	}
	opts.DialContext = proxy.DialContext
	// resolve domain name
	dnsOpts := node.clientMgr.GetDNSOptions()
	result, err := node.global.DNSClient.ResolveContext(ctx, host, dnsOpts)
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
		ctx:      node,
		listener: listener,
		GUID:     guid,
		rand:     random.NewRand(),
	}
	client.Conn = newConn(node, conn, guid, connUsageClient)
	err = client.handshake(ctx, conn)
	if err != nil {
		_ = conn.Close()
		const format = "failed to handshake with node listener: %s"
		return nil, errors.WithMessagef(err, format, listener)
	}
	node.clientMgr.Add(client)
	client.Conn.Log(logger.Debug, "create client")
	return client, nil
}

func (client *Client) handshake(ctx context.Context, conn *xnet.Conn) error {
	// interrupt
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				client.Conn.Log(logger.Fatal, xpanic.Print(r, "Client.handshake"))
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
	_, ok, err := protocol.VerifyCertificate(conn, publicKey, client.GUID)
	if err != nil {
		client.Conn.Log(logger.Exploit, err)
		return err
	}
	if !ok {
		return errors.New("failed to verify node certificate")
	}
	// send role
	_, err = conn.Write(protocol.Node.Bytes())
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
		client.Conn.Logf(logger.Exploit, format, err, spew.Sdump(data[:n]))
		return err
	}
	return nil
}

// Connect is used to start protocol.HandleConn(), if you want to
// start Synchronize(), you must call this function first.
func (client *Client) Connect() error {
	// send connect operation
	_, err := client.Conn.Write([]byte{protocol.NodeOperationConnect})
	if err != nil {
		return errors.Wrap(err, "failed to send connect operation")
	}
	err = client.authenticate()
	if err != nil {
		return err
	}
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
				client.Conn.Log(logger.Fatal, xpanic.Print(r, "client.HandleConn"))
			}
			// logoff forwarder
			client.syncMu.Lock()
			defer client.syncMu.Unlock()
			if client.isSync() {
				client.ctx.forwarder.LogoffClient(client.GUID)
			}
			client.Close()
		}()
		protocol.HandleConn(client.Conn, client.onFrame)
	}()
	timeout := client.ctx.clientMgr.GetTimeout()
	_ = client.Conn.SetDeadline(time.Now().Add(timeout))
	client.Conn.Log(logger.Debug, "connected")
	return nil
}

func (client *Client) authenticate() error {
	// receive challenge
	challenge, err := client.Conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive challenge")
	}
	if len(challenge) != protocol.ChallengeSize {
		err = errors.New("invalid challenge size")
		client.Conn.Log(logger.Exploit, err)
		return err
	}
	// send signature
	err = client.Conn.SendMessage(client.ctx.global.Sign(challenge))
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
			_ = client.Conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
			_, err = client.Conn.Write(buffer.Bytes())
			if err != nil {
				_ = client.Conn.Close()
				return
			}
			// receive reply
			select {
			case <-client.heartbeat:
			case <-sleeper.SleepSecond(30, 60):
				client.Conn.Log(logger.Warning, "receive heartbeat timeout")
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

func (client *Client) isSync() bool {
	return atomic.LoadInt32(&client.inSync) != 0
}

func (client *Client) onFrame(frame []byte) {
	if client.Conn.onFrame(frame) {
		return
	}
	if frame[0] == protocol.ConnReplyHeartbeat {
		select {
		case client.heartbeat <- struct{}{}:
		case <-client.stopSignal:
		}
		return
	}
	if client.isSync() {
		if client.onFrameAfterSync(frame) {
			return
		}
	}
	const format = "unknown command: %d\nframe:\n%s"
	client.Conn.Logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
	client.Close()
}

// Synchronize is used to switch to synchronize mode.
func (client *Client) Synchronize() error {
	client.syncMu.Lock()
	defer client.syncMu.Unlock()
	if client.isSync() {
		return errors.New("already synchronize")
	}
	// initialize sync pool
	client.Conn.SendPool.New = func() interface{} {
		return protocol.NewSend()
	}
	client.Conn.AckPool.New = func() interface{} {
		return protocol.NewAcknowledge()
	}
	client.Conn.AnswerPool.New = func() interface{} {
		return protocol.NewAnswer()
	}
	client.Conn.QueryPool.New = func() interface{} {
		return protocol.NewQuery()
	}
	// must presume, or may be lost message.
	var err error
	atomic.StoreInt32(&client.inSync, 1)
	defer func() {
		if err != nil {
			atomic.StoreInt32(&client.inSync, 0)
		}
	}()
	resp, err := client.Conn.SendCommand(protocol.NodeSync, nil)
	if err != nil {
		return errors.Wrap(err, "failed to receive synchronize response")
	}
	if !bytes.Equal(resp, []byte{protocol.NodeSync}) {
		err = errors.Errorf("failed to start to synchronize: %s", resp)
		return err // can't return directly
	}
	err = client.ctx.forwarder.RegisterClient(client)
	if err != nil {
		return err
	}
	client.Conn.Logf(logger.Debug, "start to synchronize\nlistener: %s", client.listener)
	return nil
}

func (client *Client) onFrameAfterSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	if client.onFrameAfterSyncAboutCtrl(frame[0], id, data) {
		return true
	}
	if client.onFrameAfterSyncAboutNode(frame[0], id, data) {
		return true
	}
	if client.onFrameAfterSyncAboutBeacon(frame[0], id, data) {
		return true
	}
	return false
}

func (client *Client) onFrameAfterSyncAboutCtrl(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.CtrlSendToNodeGUID:
		client.Conn.HandleSendToNodeGUID(id, data)
	case protocol.CtrlSendToNode:
		client.Conn.HandleSendToNode(id, data)
	case protocol.CtrlAckToNodeGUID:
		client.Conn.HandleAckToNodeGUID(id, data)
	case protocol.CtrlAckToNode:
		client.Conn.HandleAckToNode(id, data)
	case protocol.CtrlSendToBeaconGUID:
		client.Conn.HandleSendToBeaconGUID(id, data)
	case protocol.CtrlSendToBeacon:
		client.Conn.HandleSendToBeacon(id, data)
	case protocol.CtrlAckToBeaconGUID:
		client.Conn.HandleAckToBeaconGUID(id, data)
	case protocol.CtrlAckToBeacon:
		client.Conn.HandleAckToBeacon(id, data)
	case protocol.CtrlBroadcastGUID:
		client.Conn.HandleBroadcastGUID(id, data)
	case protocol.CtrlBroadcast:
		client.Conn.HandleBroadcast(id, data)
	case protocol.CtrlAnswerGUID:
		client.Conn.HandleAnswerGUID(id, data)
	case protocol.CtrlAnswer:
		client.Conn.HandleAnswer(id, data)
	default:
		return false
	}
	return true
}

func (client *Client) onFrameAfterSyncAboutNode(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.NodeSendGUID:
		client.Conn.HandleNodeSendGUID(id, data)
	case protocol.NodeSend:
		client.Conn.HandleNodeSend(id, data)
	case protocol.NodeAckGUID:
		client.Conn.HandleNodeAckGUID(id, data)
	case protocol.NodeAck:
		client.Conn.HandleNodeAck(id, data)
	default:
		return false
	}
	return true
}

func (client *Client) onFrameAfterSyncAboutBeacon(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.BeaconSendGUID:
		client.Conn.HandleBeaconSendGUID(id, data)
	case protocol.BeaconSend:
		client.Conn.HandleBeaconSend(id, data)
	case protocol.BeaconAckGUID:
		client.Conn.HandleBeaconAckGUID(id, data)
	case protocol.BeaconAck:
		client.Conn.HandleBeaconAck(id, data)
	case protocol.BeaconQueryGUID:
		client.Conn.HandleQueryGUID(id, data)
	case protocol.BeaconQuery:
		client.Conn.HandleQuery(id, data)
	default:
		return false
	}
	return true
}

// Status is used to get connection status.
func (client *Client) Status() *xnet.Status {
	return client.Conn.Status()
}

// Close is used to disconnect node.
func (client *Client) Close() {
	client.closeOnce.Do(func() {
		_ = client.Conn.Close()
		if client.stopSignal != nil {
			close(client.stopSignal)
		}
		client.wg.Wait()
		client.ctx.clientMgr.Delete(client.tag)
		if client.stopSignal != nil {
			client.Conn.Log(logger.Info, "disconnected")
		}
	})
}
