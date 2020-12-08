package nmap

import (
	"time"

	"project/internal/logger"
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
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			s.log(logger.Fatal, xpanic.Printf(r, "Scanner.worker-%d", id))
		}
		// restart
		time.Sleep(time.Second)
		s.wg.Add(1)
		go s.worker(id)
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
			s.process(id, job)
		case <-s.ctx.Done():
			return
		}
	}
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

func (s *Scanner) process(id int, job *Job) {
	// update status
	s.updateWorkerStatus(id, &WorkerStatus{
		Active: time.Now().Unix(),
	})
	defer func() {
		s.updateWorkerStatus(id, &WorkerStatus{
			Idle: time.Now().Unix(),
		})
	}()

}
