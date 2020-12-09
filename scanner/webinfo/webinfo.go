package webinfo

import (
	"context"
	"sync"
	"time"

	"project/internal/logger"
	"project/internal/task/pauser"
)

// Job contains job information.
type Job struct {
	// URL is the target URL, must make sure URL is valid.
	URL string `toml:"url" json:"url"`

	// Extra is used to store extra information like unit.
	// It is not the nmap argument.
	Extra string `toml:"extra" json:"extra"`
}

// Output contain worker detect output.
type Output struct {
	URL string
	OS  []string // operating system
	CMS []string // content management system
	WS  []string // web server
	DB  []string // database
	IP  []string // IP address
}

// Result contains collected information.
type Result struct {
	// Job contain raw job information.
	Job *Job `json:"job"`

	// Output is the scan output.
	Output *Output `json:"output"`

	// Error is the running error.
	Error error `json:"error"`

	// WorkerID is the worker id, if worker appear panic,
	// we can trace stack to fix problem.
	WorkerID int `json:"worker_id"`

	// ElapsedTime is the nmap running time.
	ElapsedTime time.Duration `json:"elapsed_time"`
}

// Config contains web info module configuration.
type Config struct {
	Worker       int
	CMSFinger    []byte
	WappalyzerDB []byte
}

// WebInfo is used to collect web information, it will receive scan
// job and send it to worker, worker will use go-colly to crawl target
// URL, and use wappalyzer to detect target.
type WebInfo struct {
	jobCh  <-chan *Job   // receive scan jobs
	logger logger.Logger // parent logger

	workerNum int // worker number

	// store workers status.
	workerStatus    []*WorkerStatus
	workerStatusRWM sync.RWMutex

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

// NewWebInfo is used to create a web info module.
func NewWebInfo(job <-chan *Job, logger logger.Logger, cfg *Config) (*WebInfo, error) {
	return nil, nil
}
