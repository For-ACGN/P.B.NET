package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/ed25519"
	"project/internal/crypto/rand"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/xnet"
	"project/internal/xpanic"
)

type client struct {
	ctx *Node

	node      *bootstrap.Node
	guid      []byte // node guid
	closeFunc func()

	conn      *conn
	heartbeat chan struct{}
	inSync    int32
	syncM     sync.Mutex

	closeOnce  sync.Once
	stopSignal chan struct{}
	wg         sync.WaitGroup
}

// when guid != ctrl guid for forwarder
// when guid == ctrl guid for register
// switch Register() or Connect() after newClient()
func newClient(
	ctx context.Context,
	node *Node,
	n *bootstrap.Node,
	guid []byte,
	closeFunc func(),
) (*client, error) {
	host, port, err := net.SplitHostPort(n.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cfg := xnet.Config{
		Network: n.Network,
		Timeout: node.client.Timeout,
	}
	cfg.TLSConfig = &tls.Config{
		Rand:       rand.Reader,
		Time:       node.global.Now,
		ServerName: host,
		RootCAs:    x509.NewCertPool(),
		MinVersion: tls.VersionTLS12,
	}
	// add CA certificates
	for _, cert := range node.global.Certificates() {
		cfg.TLSConfig.RootCAs.AddCert(cert)
	}
	// set proxy
	p, _ := node.global.GetProxyClient(node.client.ProxyTag)
	cfg.Dialer = p.DialContext
	// resolve domain name
	result, err := node.global.ResolveWithContext(ctx, host, &node.client.DNSOpts)
	if err != nil {
		return nil, err
	}
	var conn *xnet.Conn
	for i := 0; i < len(result); i++ {
		cfg.Address = net.JoinHostPort(result[i], port)
		c, err := xnet.DialContext(ctx, n.Mode, &cfg)
		if err == nil {
			conn = xnet.NewConn(c, node.global.Now())
			break
		}
	}
	if conn == nil {
		return nil, errors.Errorf("failed to connect node: %s", n.Address)
	}

	// handshake
	client := client{
		ctx:       node,
		node:      n,
		guid:      guid,
		closeFunc: closeFunc,
	}
	err = client.handshake(conn)
	if err != nil {
		_ = conn.Close()
		const format = "failed to handshake with node: %s"
		return nil, errors.WithMessagef(err, format, n.Address)
	}
	client.conn = newConn(node, conn, guid, connUsageClient)
	return &client, nil
}

func (client *client) handshake(conn *xnet.Conn) error {
	_ = conn.SetDeadline(client.ctx.global.Now().Add(client.ctx.client.Timeout))
	// about check connection
	sizeByte := make([]byte, 1)
	_, err := io.ReadFull(conn, sizeByte)
	if err != nil {
		return errors.Wrap(err, "failed to receive check connection size")
	}
	size := int(sizeByte[0])
	checkData := make([]byte, size)
	_, err = io.ReadFull(conn, checkData)
	if err != nil {
		return errors.Wrap(err, "failed to receive check connection data")
	}
	_, err = conn.Write(random.New().Bytes(size))
	if err != nil {
		return errors.Wrap(err, "failed to send check connection data")
	}
	// receive certificate
	cert, err := conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive certificate")
	}
	if !client.verifyCertificate(cert, client.node.Address, client.guid) {
		client.conn.Log(logger.Exploit, protocol.ErrInvalidCertificate)
		return protocol.ErrInvalidCertificate
	}
	// send role
	_, err = conn.Write(protocol.Node.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to send role")
	}
	// send self guid
	_, err = conn.Write(client.ctx.global.GUID())
	return err
}

func (client *client) verifyCertificate(cert []byte, address string, guid []byte) bool {
	if len(cert) != 2*ed25519.SignatureSize {
		return false
	}
	// verify certificate
	buffer := bytes.Buffer{}
	buffer.WriteString(address)
	buffer.Write(guid)
	if bytes.Equal(guid, protocol.CtrlGUID) {
		certWithCtrlGUID := cert[ed25519.SignatureSize:]
		return client.ctx.global.CtrlVerify(buffer.Bytes(), certWithCtrlGUID)
	}
	certWithNodeGUID := cert[:ed25519.SignatureSize]
	return client.ctx.global.CtrlVerify(buffer.Bytes(), certWithNodeGUID)
}

// if return error, must close manually
func (client *client) Connect() error {
	// send operation
	_, err := client.conn.Write([]byte{2})
	if err != nil {
		return errors.Wrap(err, "failed to send operation")
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
				client.conn.Log(logger.Fatal, xpanic.Error(r, "client.HandleConn"))
			}
			client.Close()
		}()
		protocol.HandleConn(client.conn, client.onFrame)
	}()
	return nil
}

func (client *client) authenticate() error {
	// receive challenge
	challenge, err := client.conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive challenge")
	}
	if len(challenge) < 2048 || len(challenge) > 4096 {
		err = errors.New("invalid challenge size")
		client.conn.Log(logger.Exploit, err)
		return err
	}
	// send signature
	err = client.conn.SendMessage(client.ctx.global.Sign(challenge))
	if err != nil {
		return errors.Wrap(err, "failed to send challenge signature")
	}
	resp, err := client.conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive authentication response")
	}
	if !bytes.Equal(resp, protocol.AuthSucceed) {
		return errors.WithStack(protocol.ErrAuthenticateFailed)
	}
	return nil
}

func (client *client) isSync() bool {
	return atomic.LoadInt32(&client.inSync) != 0
}

func (client *client) onFrame(frame []byte) {
	if client.conn.onFrame(frame) {
		return
	}
	if frame[0] == protocol.ConnReplyHeartbeat {
		select {
		case client.heartbeat <- struct{}{}:
		case <-client.stopSignal:
		}
	}
	if client.isSync() {
		if client.onFrameAfterSync(frame) {
			return
		}
	}
	const format = "unknown command: %d\nframe:\n%s"
	client.conn.Logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
	client.Close()
}

func (client *client) onFrameAfterSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	switch frame[0] {
	case protocol.CtrlSendToNodeGUID:
		client.conn.HandleSendToNodeGUID(id, data)
	case protocol.CtrlSendToNode:
		client.conn.HandleSendToNode(id, data)
	case protocol.CtrlAckToNodeGUID:
		client.conn.HandleAckToNodeGUID(id, data)
	case protocol.CtrlAckToNode:
		client.conn.HandleAckToNode(id, data)
	case protocol.CtrlSendToBeaconGUID:
		client.conn.HandleSendToBeaconGUID(id, data)
	case protocol.CtrlSendToBeacon:
		client.conn.HandleSendToBeacon(id, data)
	case protocol.CtrlAckToBeaconGUID:
		client.conn.HandleAckToBeaconGUID(id, data)
	case protocol.CtrlAckToBeacon:
		client.conn.HandleAckToBeacon(id, data)
	case protocol.CtrlBroadcastGUID:
		client.conn.HandleBroadcastGUID(id, data)
	case protocol.CtrlBroadcast:
		client.conn.HandleBroadcast(id, data)
	case protocol.CtrlAnswerGUID:
		client.conn.HandleAnswerGUID(id, data)
	case protocol.CtrlAnswer:
		client.conn.HandleAnswer(id, data)
	case protocol.NodeSendGUID:
		client.conn.HandleNodeSendGUID(id, data)
	case protocol.NodeSend:
		client.conn.HandleNodeSend(id, data)
	case protocol.NodeAckGUID:
		client.conn.HandleNodeAckGUID(id, data)
	case protocol.NodeAck:
		client.conn.HandleNodeAck(id, data)
	case protocol.BeaconSendGUID:
		client.conn.HandleBeaconSendGUID(id, data)
	case protocol.BeaconSend:
		client.conn.HandleBeaconSend(id, data)
	case protocol.BeaconAckGUID:
		client.conn.HandleBeaconAckGUID(id, data)
	case protocol.BeaconAck:
		client.conn.HandleBeaconAck(id, data)
	case protocol.BeaconQueryGUID:
		client.conn.HandleBeaconQueryGUID(id, data)
	case protocol.BeaconQuery:
		client.conn.HandleBeaconQuery(id, data)
	default:
		return false
	}
	return true
}

func (client *client) sendHeartbeatLoop() {
	defer client.wg.Done()
	var err error
	r := random.New()
	buffer := bytes.NewBuffer(nil)
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	for {
		timer.Reset(time.Duration(30+r.Int(60)) * time.Second)
		select {
		case <-timer.C:
			// <security> fake traffic like client
			fakeSize := 64 + r.Int(256)
			// size(4 Bytes) + heartbeat(1 byte) + fake data
			buffer.Reset()
			buffer.Write(convert.Uint32ToBytes(uint32(1 + fakeSize)))
			buffer.WriteByte(protocol.ConnSendHeartbeat)
			buffer.Write(r.Bytes(fakeSize))
			// send
			_ = client.conn.SetWriteDeadline(time.Now().Add(protocol.SendTimeout))
			_, err = client.conn.Write(buffer.Bytes())
			if err != nil {
				return
			}
			// receive reply
			timer.Reset(time.Duration(30+r.Int(60)) * time.Second)
			select {
			case <-client.heartbeat:
			case <-timer.C:
				client.conn.Log(logger.Warning, "receive heartbeat timeout")
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

// Sync is used to switch to sync mode
func (client *client) Sync() error {
	client.syncM.Lock()
	defer client.syncM.Unlock()
	if client.isSync() {
		return nil
	}
	resp, err := client.conn.SendCommand(protocol.NodeSync, nil)
	if err != nil {
		return errors.Wrap(err, "failed to receive sync response")
	}
	if !bytes.Equal(resp, []byte{protocol.NodeSync}) {
		return errors.Errorf("failed to start sync: %s", resp)
	}
	// initialize sync pool
	client.conn.SendPool.New = func() interface{} {
		return &protocol.Send{
			GUID:      make([]byte, guid.Size),
			RoleGUID:  make([]byte, guid.Size),
			Message:   make([]byte, aes.BlockSize),
			Hash:      make([]byte, sha256.Size),
			Signature: make([]byte, ed25519.SignatureSize),
		}
	}
	client.conn.AckPool.New = func() interface{} {
		return &protocol.Acknowledge{
			GUID:      make([]byte, guid.Size),
			RoleGUID:  make([]byte, guid.Size),
			SendGUID:  make([]byte, guid.Size),
			Signature: make([]byte, ed25519.SignatureSize),
		}
	}
	client.conn.AnswerPool.New = func() interface{} {
		return &protocol.Answer{
			GUID:       make([]byte, guid.Size),
			BeaconGUID: make([]byte, guid.Size),
			Message:    make([]byte, aes.BlockSize),
			Hash:       make([]byte, sha256.Size),
			Signature:  make([]byte, ed25519.SignatureSize),
		}
	}
	client.conn.QueryPool.New = func() interface{} {
		return &protocol.Query{
			GUID:       make([]byte, guid.Size),
			BeaconGUID: make([]byte, guid.Size),
			Signature:  make([]byte, ed25519.SignatureSize),
		}
	}
	// TODO register
	// client.ctx.forwarder.RegisterNode(client)
	atomic.StoreInt32(&client.inSync, 1)
	return nil
}

// Status is used to get connection status
func (client *client) Status() *xnet.Status {
	return client.conn.Status()
}

// Close is used to disconnect node
func (client *client) Close() {
	client.closeOnce.Do(func() {
		_ = client.conn.Close()
		close(client.stopSignal)
		client.wg.Wait()
		if client.closeFunc != nil {
			client.closeFunc()
		}
		client.conn.Log(logger.Info, "disconnected")
	})
}
