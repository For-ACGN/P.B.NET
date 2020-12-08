package nmap

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/task/pauser"
)

// Result contain scan output and job extra information.
type Result struct {
	// Output is the scan output.
	Output *Output `json:"output"`

	// Extra is the job extra information.
	Extra string `json:"extra"`

	// Error is the running error.
	Error error `json:"error"`

	// WorkerID is the worker id, if worker appear panic,
	// we can trace stack to fix problem.
	WorkerID int `json:"worker_id"`

	// ElapsedTime is the nmap running time.
	ElapsedTime time.Duration `json:"elapsed_time"`
}

// Scanner is a nmap scanner wrapper, it will receive scan job and
// send job to worker, worker will call nmap to scan target and save
// scan result, after scan finish, worker will parse output file.
type Scanner struct {
	jobCh     <-chan *Job   // receive scan jobs
	workerNum int           // worker number
	logger    logger.Logger // parent logger
	opts      *Options      // default job options

	binPath    string // nmap binary file path.
	outputPath string // nmap output directory path.

	// store workers status.
	workerStatus    []*WorkerStatus
	workerStatusRWM sync.RWMutex

	// for load balance if scanner default options
	// use local IP, worker will select one from it
	localIPs   map[string]bool
	localIPsMu sync.Mutex

	// Result is a stream, it will never be closed.
	Result chan *Result

	// for control operations
	pause   *pauser.Pauser
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	rwm     sync.RWMutex
	wg      sync.WaitGroup
}

// New is used to create a new nmap scanner.
func New(job <-chan *Job, worker int, logger logger.Logger, opts *Options) *Scanner {
	scanner := Scanner{
		jobCh:        job,
		workerNum:    worker,
		logger:       logger,
		opts:         opts,
		binPath:      "nmap",
		outputPath:   "output",
		workerStatus: make([]*WorkerStatus, worker),
		Result:       make(chan *Result, 64*worker),
		pause:        pauser.New(),
	}
	// initialize worker status
	for i := 0; i < worker; i++ {
		scanner.workerStatus[i] = &WorkerStatus{
			Idle: time.Now().Unix(),
		}
	}
	// set scanner options
	if opts == nil {
		return &scanner
	}
	if opts.BinPath != "" {
		scanner.binPath = opts.BinPath
	}
	if opts.OutputPath == "" {
		scanner.outputPath = opts.OutputPath
	}
	if len(opts.LocalIP) != 0 {
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
		return errors.New("nmap scanner is started")
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	for i := 0; i < s.workerNum; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
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
	// prevent panic before here
	s.started = false
}

// Restart is used to restart nmap scanner.
func (s *Scanner) Restart() error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.stop()
	s.wg.Wait()
	return s.start()
}

// IsStarted is used to check nmap scanner is started.
func (s *Scanner) IsStarted() bool {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.started
}

// Pause is used to pause nmap scanner, it will not pause
// created nmap process, the next job will block.
func (s *Scanner) Pause() {
	s.pause.Pause()
}

// Continue is used to continue nmap scanner.
func (s *Scanner) Continue() {
	s.pause.Continue()
}
