package module

import (
	"sync"

	"github.com/pkg/errors"
)

// Manager is the module manager.
type Manager struct {
	// key = module tag
	modules map[string]Module
	closed  bool
	rwm     sync.RWMutex
}

// NewManager is used to create a module manager.
func NewManager() *Manager {
	return &Manager{
		modules: make(map[string]Module),
	}
}

// Add is used to add a module to manager.
func (m *Manager) Add(tag string, module Module) error {
	if tag == "" {
		return errors.New("empty module tag")
	}
	m.rwm.Lock()
	defer m.rwm.Unlock()
	if m.closed {
		return errors.New("proxy server manager closed")
	}
	if _, ok := m.modules[tag]; !ok {
		m.modules[tag] = module
		return nil
	}
	return errors.Errorf("module %s is already exists", tag)
}

// Delete is used to delete a module by tag.
func (m *Manager) Delete(tag string) error {
	if tag == "" {
		return errors.New("empty module tag")
	}
	m.rwm.Lock()
	defer m.rwm.Unlock()
	if module, ok := m.modules[tag]; ok {
		module.Stop()
		delete(m.modules, tag)
		return nil
	}
	return errors.Errorf("module %s is not exist", tag)
}

// Get is used to get a module by tag.
func (m *Manager) Get(tag string) (Module, error) {
	if tag == "" {
		return nil, errors.New("empty module tag")
	}
	m.rwm.RLock()
	defer m.rwm.RUnlock()
	if module, ok := m.modules[tag]; ok {
		return module, nil
	}
	return nil, errors.Errorf("module %s is not exist", tag)
}

// Start is used to start a module by tag.
func (m *Manager) Start(tag string) error {
	module, err := m.Get(tag)
	if err != nil {
		return err
	}
	return module.Start()
}

// Stop is used to stop a module by tag.
func (m *Manager) Stop(tag string) error {
	module, err := m.Get(tag)
	if err != nil {
		return err
	}
	module.Stop()
	return nil
}

// Restart is used to restart a module by tag.
func (m *Manager) Restart(tag string) error {
	module, err := m.Get(tag)
	if err != nil {
		return err
	}
	return module.Restart()
}

// Info is used to get module information by tag.
func (m *Manager) Info(tag string) (string, error) {
	module, err := m.Get(tag)
	if err != nil {
		return "", err
	}
	return module.Info(), nil
}

// Status is used to get module status by tag.
func (m *Manager) Status(tag string) (string, error) {
	module, err := m.Get(tag)
	if err != nil {
		return "", err
	}
	return module.Status(), nil
}

// Modules is used to get all modules.
func (m *Manager) Modules() map[string]Module {
	m.rwm.RLock()
	defer m.rwm.RUnlock()
	modules := make(map[string]Module, len(m.modules))
	for tag, module := range m.modules {
		modules[tag] = module
	}
	return modules
}

// Close is used to stop all modules.
func (m *Manager) Close() {
	m.rwm.Lock()
	defer m.rwm.Unlock()
	if m.closed {
		return
	}
	for tag, module := range m.modules {
		module.Stop()
		delete(m.modules, tag)
	}
	// prevent panic before here
	m.closed = true
}
