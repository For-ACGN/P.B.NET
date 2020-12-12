package kcp

import (
	"crypto/sha256"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

const (
	iter   = 1024
	keyLen = 32

	dataShards   = 10
	parityShards = 3
)

type listener struct {
	*kcp.Listener
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.AcceptKCP()
	if err != nil {
		return nil, err
	}
	conn.SetStreamMode(true)
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, _ = io.ReadFull(conn, make([]byte, 1))
	return conn, nil
}

func checkPasswordAndSalt(password, salt []byte) error {
	if len(password) < 16 {
		return errors.New("password length < 16")
	}
	if len(salt) < 8 {
		return errors.New("salt length < 8")
	}
	return nil
}

// Listen listens for incoming KCP packets addressed to the local address.
func Listen(address string, password, salt []byte) (net.Listener, error) {
	err := checkPasswordAndSalt(password, salt)
	if err != nil {
		return nil, err
	}
	key := pbkdf2.Key(password, salt, iter, keyLen, sha256.New)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	l, err := kcp.ListenWithOptions(address, block, dataShards, parityShards)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &listener{Listener: l}, nil
}

// Conn is a kcp UDP session wrapper.
type Conn struct {
	// must close rawConn manually to prevent goroutine leak
	// in package github.com/lucas-clemente/quic-go
	// go m.listen() in newPacketHandlerMap()
	rawConn net.PacketConn

	*kcp.UDPSession
}

// Close is used to close kcp session wrapper.
func (c *Conn) Close() error {
	err := c.UDPSession.Close()
	_ = c.rawConn.Close()
	return err
}

// Dial connects to the remote address with packet encryption.
func Dial(address string, password, salt []byte) (net.Conn, error) {
	err := checkPasswordAndSalt(password, salt)
	if err != nil {
		return nil, err
	}
	key := pbkdf2.Key(password, salt, iter, keyLen, sha256.New)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rawConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	conn, err := kcp.NewConn(address, block, dataShards, parityShards, rawConn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// conn, err := kcp.DialWithOptions(address, block, dataShards, parityShards)
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }
	// var ok bool
	// defer func() {
	// 	if !ok {
	// 		_ = conn.Close()
	// 	}
	// }()
	conn.SetStreamMode(true)

	// prevent server side block
	_, err = conn.Write([]byte{0})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// ok = true
	return &Conn{
		rawConn:    rawConn,
		UDPSession: conn,
	}, nil
}
