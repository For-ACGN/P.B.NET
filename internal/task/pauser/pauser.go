package pauser

import (
	"context"
	"sync/atomic"
)

// states about pauser
const (
	_ int32 = iota
	StateRunning
	StatePaused
	StateCancel
)

// Pauser is used to pause in a loop.
type Pauser struct {
	ctx     context.Context
	state   *int32
	pauseCh chan struct{}
}

// New is used to create a new pauser.
// context is used to prevent block.
func New(ctx context.Context) *Pauser {
	state := StateRunning
	return &Pauser{
		ctx:     ctx,
		state:   &state,
		pauseCh: make(chan struct{}, 1),
	}
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
		atomic.StoreInt32(pauser.state, StateCancel)
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

// State is used to get current state.
func (pauser *Pauser) State() int32 {
	return atomic.LoadInt32(pauser.state)
}
