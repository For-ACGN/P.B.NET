package nmap

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// Result contain scan output and job extra information.
type Result struct {
	// Output is the scan output.
	Output *Output

	// Extra is the job extra information.
	Extra string

	// Error is the running error.
	Error error

	// WorkerID is the worker id, if worker appear panic,
	// we can trace stack to fix problem.
	WorkerID int
}

// Scanner is a nmap scanner wrapper, it will receive scan job and
// send job to worker, worker will call nmap to scan target and save
// scan result, after scan finish, worker will parse output file.
type Scanner struct {
	jobCh   <-chan *Job // receive scan jobs
	workers int         // worker number
	opts    *Options    // default job options

	// Result is a stream, it will never be closed.
	Result chan *Result

	// for load balance if scanner default options
	// use local IP, worker will select one from it
	localIPs   map[string]bool
	localIPsMu sync.Mutex

	// for control operations
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	rwm     sync.RWMutex
	wg      sync.WaitGroup
}

// NewScanner is used to create a new nmap scanner.
func NewScanner(jobCh <-chan *Job, worker int, opts *Options) *Scanner {
	scanner := Scanner{
		jobCh:   jobCh,
		workers: worker,
		opts:    opts,
		Result:  make(chan *Result, 1024),
	}
	if opts != nil && len(opts.LocalIP) > 0 {
		l := len(opts.LocalIP)
		scanner.localIPs = make(map[string]bool, l)
		for i := 0; i < l; i++ {
			scanner.localIPs[opts.LocalIP[i]] = false
		}
	}
	return &scanner
}

// Start is used to start scanner, it will start to process scan jobs.
func (s *Scanner) Start() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.start()
}

func (s *Scanner) start() error {
	if s.started {
		return errors.New("scanner is started")
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.started = true
	return nil
}

// Stop is used to stop scanner, it will kill all processing jobs.
func (s *Scanner) Stop() {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.stop()
	s.wg.Wait()
}

func (s *Scanner) stop() {
	if !s.started {
		return
	}
	s.cancel()
	s.started = false
}

// Restart is used to restart slaver.
func (s *Scanner) Restart() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.stop()
	s.wg.Wait()
	return s.start()
}

// IsStarted is used to check slaver is started.
func (s *Scanner) IsStarted() bool {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.started
}

func (s *Scanner) Pause() {

}

func (s *Scanner) Continue() {

}
