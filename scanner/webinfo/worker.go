package webinfo

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

func (wi *WebInfo) updateWorkerStatus(id int, status *WorkerStatus) {
	wi.workerStatusRWM.Lock()
	defer wi.workerStatusRWM.Unlock()
	if status.Idle != 0 {
		wi.workerStatus[id].Idle = status.Idle
	}
	if status.Active != 0 {
		wi.workerStatus[id].Active = status.Active
	}
}

func (wi *WebInfo) logf(lv logger.Level, format string, log ...interface{}) {
	wi.logger.Printf(lv, "web info", format, log...)
}

func (wi *WebInfo) log(lv logger.Level, log ...interface{}) {
	wi.logger.Println(lv, "web info", log...)
}

func (wi *WebInfo) worker(id int) {
	wi.logf(logger.Debug, "worker %d started", id)
	defer wi.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			wi.log(logger.Fatal, xpanic.Printf(r, "WebInfo.worker-%d", id))
			// restart
			time.Sleep(time.Second)
			wi.wg.Add(1)
			go wi.worker(id)
		}
	}()
	// process jobs loop
	for {
		select {
		case job := <-wi.jobCh:
			if job == nil {
				return
			}
			begin := time.Now()
			result := wi.process(id, job)
			result.ElapsedTime = time.Since(begin)
			wi.sendResult(result)
		case <-wi.ctx.Done():
			return
		}
	}
}

func (wi *WebInfo) sendResult(result *Result) {
	select {
	case wi.Result <- result:
	case <-wi.ctx.Done():
		return
	}
}

func (wi *WebInfo) process(id int, job *Job) (result *Result) {
	result = &Result{
		Job:      job,
		WorkerID: id,
	}
	var err error
	defer func() {
		result.Error = err
	}()
	// update worker status
	wi.updateWorkerStatus(id, &WorkerStatus{
		Active: time.Now().Unix(),
	})
	defer func() {
		wi.updateWorkerStatus(id, &WorkerStatus{
			Idle: time.Now().Unix(),
		})
	}()

	return
}
