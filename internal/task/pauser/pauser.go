package pauser

import (
	"context"
	"sync/atomic"
)

// states about pauser.
const (
	_ int32 = iota
	StateRunning
	StatePaused
	StateClosed
)

// Pauser is used to pause a loop.
type Pauser struct {
	state   *int32
	pauseCh chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
}

// New is used to create a new pauser.
func New() *Pauser {
	state := StateRunning
	p := Pauser{
		state:   &state,
		pauseCh: make(chan struct{}, 1),
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	return &p
}

// Paused is used to check need pause in current loop.
// This function is called in loop.
func (pauser *Pauser) Paused() {
	if atomic.LoadInt32(pauser.state) != StatePaused {
		return
	}
	select {
	case <-pauser.pauseCh:
		atomic.StoreInt32(pauser.state, StateRunning)
	case <-pauser.ctx.Done():
		atomic.StoreInt32(pauser.state, StateClosed)
	}
}

// Pause is used to pause current loop.
func (pauser *Pauser) Pause() {
	atomic.StoreInt32(pauser.state, StatePaused)
}

// Continue is used to continue current loop.
func (pauser *Pauser) Continue() {
	if atomic.LoadInt32(pauser.state) != StatePaused {
		return
	}
	select {
	case pauser.pauseCh <- struct{}{}:
	default:
	}
}

// Close is used to close pauser.
func (pauser *Pauser) Close() {
	pauser.cancel()
	atomic.StoreInt32(pauser.state, StateClosed)
}

// State is used to get current state.
func (pauser *Pauser) State() int32 {
	return atomic.LoadInt32(pauser.state)
}
