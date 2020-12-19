package webinfo

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/task/pauser"
)

// Job contains job information.
type Job struct {
	// URL is the target URL, must make sure URL is valid.
	URL string `toml:"url" json:"url"`

	// Extra is used to store extra information like unit.
	Extra string `toml:"extra" json:"extra"`
}

// Output contain worker detect output.
type Output struct {
	Title string   // page title
	OS    []string // operating system
	CMS   []string // content management system
	WS    []string // web server
	DB    []string // database
	IP    []string // IP address
}

// Result contains collected information.
type Result struct {
	// Job contain raw job information.
	Job *Job `json:"job"`

	// Output is the collect output.
	Output *Output `json:"output"`

	// Error is the running error.
	Error error `json:"error"`

	// WorkerID is the worker id, if worker appear panic,
	// we can trace stack to fix problem.
	WorkerID int `json:"worker_id"`

	// ElapsedTime is the job running time.
	ElapsedTime time.Duration `json:"elapsed_time"`
}

// Config contains web info module configuration.
type Config struct {
	Worker       int
	CMSFinger    []byte
	WappalyzerDB []byte
}

// WebInfo is used to collect web information, it will receive collect
// job and send it to worker, worker will use go-colly to crawl target
// URL, and use wappalyzer to detect target.
type WebInfo struct {
	jobCh     <-chan *Job   // receive collect jobs
	logger    logger.Logger // parent logger
	workerNum int           // worker number

	// store workers status.
	workerStatus    []*WorkerStatus
	workerStatusRWM sync.RWMutex

	// Result is a stream, it will never be closed.
	Result chan *Result

	// for control operations
	started bool
	pauser  *pauser.Pauser
	ctx     context.Context
	cancel  context.CancelFunc
	rwm     sync.RWMutex
	wg      sync.WaitGroup
}

// NewWebInfo is used to create a web info module.
func NewWebInfo(job <-chan *Job, logger logger.Logger, cfg *Config) (*WebInfo, error) {
	workerNum := cfg.Worker
	if workerNum < 1 {
		workerNum = 4
	}
	wi := WebInfo{
		jobCh:        job,
		logger:       logger,
		workerNum:    workerNum,
		workerStatus: make([]*WorkerStatus, workerNum),
		Result:       make(chan *Result, 64*workerNum),
	}
	return &wi, nil
}

// Start is used to start web info, it will start to process jobs.
func (wi *WebInfo) Start() error {
	wi.rwm.Lock()
	defer wi.rwm.Unlock()
	return wi.start()
}

func (wi *WebInfo) start() error {
	if wi.started {
		return errors.New("web info is started")
	}
	wi.pauser = pauser.New()
	wi.ctx, wi.cancel = context.WithCancel(context.Background())
	for i := 0; i < wi.workerNum; i++ {
		wi.wg.Add(1)
		go wi.worker(i)
	}
	wi.started = true
	return nil
}

// Stop is used to stop web info, it will kill all processing jobs.
func (wi *WebInfo) Stop() {
	wi.rwm.Lock()
	defer wi.rwm.Unlock()
	wi.stop()
}

func (wi *WebInfo) stop() {
	if !wi.started {
		return
	}
	wi.pauser.Close()
	wi.cancel()
	wi.wg.Wait()
	// prevent panic before here
	wi.started = false
}

// Restart is used to restart web info.
func (wi *WebInfo) Restart() error {
	wi.rwm.Lock()
	defer wi.rwm.Unlock()
	wi.stop()
	return wi.start()
}

// IsStarted is used to check web info is started.
func (wi *WebInfo) IsStarted() bool {
	wi.rwm.RLock()
	defer wi.rwm.RUnlock()
	return wi.started
}

// Pause is used to pause web info, it will pause all go-colly running
// jobs, wappalyzer will not be paused, because it only use CPU.
func (wi *WebInfo) Pause() {
	wi.rwm.RLock()
	defer wi.rwm.RUnlock()
	if wi.pauser == nil {
		return
	}
	wi.pauser.Pause()
}

// Continue is used to continue web info.
func (wi *WebInfo) Continue() {
	wi.rwm.RLock()
	defer wi.rwm.RUnlock()
	if wi.pauser == nil {
		return
	}
	wi.pauser.Continue()
}
