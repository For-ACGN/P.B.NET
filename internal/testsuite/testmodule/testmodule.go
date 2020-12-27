package testmodule

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"project/internal/module"
)

// Module is a mock module that implemented module.Module.
type Module struct {
	started    bool
	startedRWM sync.RWMutex

	// for operation
	mu sync.Mutex
	wg sync.WaitGroup
}

// Name is used to get the name of the mock module.
func (*Module) Name() string {
	return "mock module"
}

// Description is used to get the mock module description.
func (*Module) Description() string {
	return "Mock module is used to test."
}

// Start is used to start mock module.
func (mod *Module) Start() error {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	return mod.start()
}

func (mod *Module) start() error {
	mod.startedRWM.Lock()
	defer mod.startedRWM.Unlock()
	if mod.started {
		return errors.New("already started")
	}
	mod.started = true
	return nil
}

// Stop is used to stop mock module.
func (mod *Module) Stop() {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	mod.stop()
	mod.wg.Wait()
}

func (mod *Module) stop() {
	mod.startedRWM.Lock()
	defer mod.startedRWM.Unlock()
	if mod.started {
		mod.started = false
	}
}

// Restart is used to restart mock module.
func (mod *Module) Restart() error {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	mod.stop()
	mod.wg.Wait()
	return mod.start()
}

// IsStarted is used to check module is started.
func (mod *Module) IsStarted() bool {
	mod.startedRWM.RLock()
	defer mod.startedRWM.RUnlock()
	return mod.started
}

// Info is used to get the information about the mock module.
func (mod *Module) Info() string {
	mod.startedRWM.RLock()
	defer mod.startedRWM.RUnlock()
	if mod.started {
		return "mock module information(started)"
	}
	return "mock module information(stopped)"
}

// Status is used to get the status about the mock module.
func (mod *Module) Status() string {
	mod.startedRWM.RLock()
	defer mod.startedRWM.RUnlock()
	if mod.started {
		return "mock module status(started)"
	}
	return "mock module status(stopped)"
}

// Methods is used to get the mock module extended methods.
func (*Module) Methods() []*module.Method {
	scan := module.Method{
		Name: "Scan",
		Desc: "Scan is used to scan a host with port, it will return the port status",
		Args: []*module.Value{
			{Name: "host", Type: "string"},
			{Name: "port", Type: "uint16"},
		},
		Rets: []*module.Value{
			{Name: "open", Type: "bool"},
			{Name: "err", Type: "error"},
		},
	}
	return []*module.Method{&scan}
}

// Call is used to call the inner method about module.
func (mod *Module) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "Scan":
		if len(args) != 1 {
			return nil, errors.New("invalid argument number")
		}
		ip, ok := args[0].(string)
		if !ok {
			return nil, errors.New("argument 1 is not a string")
		}
		open, err := mod.Scan(ip)
		return []interface{}{open, err}, nil
	default:
		return nil, fmt.Errorf("unknown method: \"%s\"", method)
	}
}

// Scan is used to scan a ip address(fake function).
func (*Module) Scan(ip string) (bool, error) {
	if ip == "" {
		return false, errors.New("empty ip")
	}
	return true, nil
}

// New is used to create a mock module.
func New() *Module {
	return new(Module)
}
