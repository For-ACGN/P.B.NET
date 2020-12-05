package nmap

import (
	"context"
	"sync"
)

// Scanner is a nmap scanner wrapper.
type Scanner struct {
	jobCh <-chan *Job // receive scan jobs.
	opts  *Options    // default job options.

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScanner is used to create a new nmap scanner.
func NewScanner(jobCh <-chan *Job, opts *Options) (*Scanner, error) {

	return nil, nil
}
