package kcp

import (
	"crypto/sha256"
	"io"
	"net"

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
	net.Listener
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	io.ReadFull(conn, make([]byte, 1))
	return conn, nil
}

// Listen listens for incoming KCP packets addressed to the local address.
func Listen(address string, password, salt []byte) (net.Listener, error) {
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

// Dial connects to the remote address with packet encryption.
func Dial(address string, password, salt []byte) (net.Conn, error) {
	key := pbkdf2.Key(password, salt, iter, keyLen, sha256.New)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conn, err := kcp.DialWithOptions(address, block, dataShards, parityShards)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ok bool
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	// prevent server side block
	_, err = conn.Write([]byte{0})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ok = true
	return conn, nil
}
