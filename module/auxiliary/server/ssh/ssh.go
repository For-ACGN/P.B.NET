package ssh

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Server is a simple ssh server, it will start a
// simple remote command execute server.
type Server struct {
	network string
	address string
	config  *ssh.ServerConfig

	listener net.Listener

	ctx    context.Context
	cancel context.CancelFunc
}

// New is used to create a new simple ssh server.
func New(network, address string, cfg *ssh.ServerConfig) (*Server, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}
	_, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}
	srv := Server{
		network: network,
		address: address,
		config:  cfg,
	}
	return &srv, nil
}

// Serve is
func (srv *Server) Serve() {
	listener, err := net.Listen(srv.network, srv.address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = listener.Close() }()
	srv.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go srv.handleConn(conn)
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	sc, nc, req, err := ssh.NewServerConn(conn, srv.config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("welcome:", sc.User())
	select {
	case nc := <-nc:
		if nc != nil {
			fmt.Println(nc.ChannelType())
		}
	case req := <-req:
		if req != nil {
			fmt.Println(req.Type)
		}
	}
}

// Close is
func (srv *Server) Close() {
	srv.listener.Close()
}
