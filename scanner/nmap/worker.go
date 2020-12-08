package nmap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/system"
	"project/internal/xpanic"
)

// WorkerStatus contains worker status.
type WorkerStatus struct {
	Idle   int64 // unix timestamp
	Active int64 // unix timestamp
}

func (s *Scanner) updateWorkerStatus(id int, status *WorkerStatus) {
	s.workerStatusRWM.Lock()
	defer s.workerStatusRWM.Unlock()
	if status.Idle != 0 {
		s.workerStatus[id].Idle = status.Idle
	}
	if status.Active != 0 {
		s.workerStatus[id].Active = status.Active
	}
}

func (s *Scanner) logf(lv logger.Level, format string, log ...interface{}) {
	s.logger.Printf(lv, "nmap scanner-", format, log...)
}

func (s *Scanner) log(lv logger.Level, log ...interface{}) {
	s.logger.Println(lv, "nmap scanner-", log...)
}

func (s *Scanner) worker(id int) {
	s.logf(logger.Debug, "worker %d started", id)
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			s.log(logger.Fatal, xpanic.Printf(r, "Scanner.worker-%d", id))
			// restart
			time.Sleep(time.Second)
			s.wg.Add(1)
			go s.worker(id)
		}
	}()
	// wait some times for prevent burst
	select {
	case <-time.After(time.Duration(id) * 10 * time.Second):
	case <-s.ctx.Done():
		return
	}
	// process jobs loop
	for {
		select {
		case job := <-s.jobCh:
			if job == nil {
				return
			}
			begin := time.Now()
			result := s.process(id, job)
			result.ElapsedTime = time.Since(begin)
			s.sendResult(result)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scanner) sendResult(result *Result) {
	select {
	case s.Result <- result:
	case <-s.ctx.Done():
		return
	}
}

func (s *Scanner) process(id int, job *Job) (result *Result) {
	result = &Result{
		Extra:    job.Extra,
		WorkerID: id,
	}
	var err error
	defer func() {
		result.Error = err
	}()
	// update status
	s.updateWorkerStatus(id, &WorkerStatus{
		Active: time.Now().Unix(),
	})
	defer func() {
		s.updateWorkerStatus(id, &WorkerStatus{
			Idle: time.Now().Unix(),
		})
	}()
	// set scanner default job options
	if job.Options == nil {
		job.Options = s.opts
		job.isScanner = true
	}
	// set random local IP if it is scanner default job options
	if job.isScanner && s.localIPs != nil {
		job.Options.LocalIP[0] = s.selectLocalIP()
	}
	// set output path
	outputFile := fmt.Sprintf("%d-%d.xml", id, time.Now().Unix())
	outputFile = filepath.Join(s.opts.OutputPath, outputFile)
	// check destination directory can write
	err = checkOutputDirectory(filepath.Dir(outputFile))
	if err != nil {
		err = errors.Errorf("failed to check output directory: %s", err)
		return
	}
	job.outputPath = outputFile
	// generate nmap arguments
	args, err := job.ToArgs()
	if err != nil {
		err = errors.Errorf("failed to generate nmap arguments: %s", err)
		return
	}
	cmd := exec.CommandContext(s.ctx, s.binPath, args...)
	// run nmap
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.Errorf("failed to run nmap: %s\n%s", err, cmdOutput)
		return
	}
	// remove nmap output file, wait output
	time.Sleep(3 * time.Second)
	defer func() {
		err = os.Remove(outputFile)
		if err != nil {
			s.log(logger.Error, "failed to remove nmap output file:", err)
		}
		time.Sleep(time.Second)
	}()
	// parse result
	outputData, err := ioutil.ReadFile(outputFile)
	if err != nil {
		err = errors.Errorf("failed to read output: %s", err)
		return
	}
	output, err := ParseOutput(outputData)
	if err != nil {
		err = errors.Errorf("failed to parse nmap output: %s", err)
		return
	}
	result.Output = output
	return
}

// selectLocalIP is used to random select a local IP if scanner
// default job options use local IP.
func (s *Scanner) selectLocalIP() string {
	s.localIPsMu.Lock()
	defer s.localIPsMu.Unlock()
	for {
		for ip, used := range s.localIPs {
			if !used {
				s.localIPs[ip] = true
				return ip
			}
		}
		// reset all ip used flag
		for ip := range s.localIPs {
			s.localIPs[ip] = false
		}
	}
}

func checkOutputDirectory(dir string) error {
	exist, err := system.IsExist(dir)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return os.MkdirAll(dir, 0750)
}
