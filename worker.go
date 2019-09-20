package curlew

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	guard             sync.RWMutex
	d                 *Dispatcher
	Jobs              chan *Job
	lastBusyTime      time.Time
	workerIdleTimeout time.Duration
	running           bool
	closeChan         chan struct{}
	isClosed          bool
}

func NewWorker(d *Dispatcher) *Worker {
	w := new(Worker)
	w.Jobs = make(chan *Job, 1)
	w.workerIdleTimeout = d.WorkerIdleTimeout
	w.closeChan = make(chan struct{})
	w.isClosed = false
	w.schedule()
	w.d = d
	d.add(w)
	return w
}

func (w *Worker) LastBusyTime() time.Time {
	w.guard.RLock()
	defer w.guard.RUnlock()
	return w.lastBusyTime
}

func (w *Worker) SetLastBusyTime() {
	w.guard.Lock()
	defer w.guard.Unlock()
	w.lastBusyTime = time.Now().UTC()
}

func (w *Worker) close() {
	w.guard.Lock()
	defer w.guard.Unlock()
	close(w.closeChan)
	w.isClosed = true
}

func (w *Worker) IsClosed() bool {
	w.guard.Lock()
	defer w.guard.Unlock()
	return w.isClosed
}

func (w *Worker) submit(job *Job) {
	w.Jobs <- job
}

func (w *Worker) schedule() {
	go func() {
		ticker := time.NewTicker(w.workerIdleTimeout)
		defer ticker.Stop()
		for range ticker.C {
			if w.canClose() {
				w.close()
				w.d.remove(w)
				return
			}
		}
	}()
	go func() {
		var jr *Job
		defer func() {
			if r := recover(); r != nil {
				w.d.monitor(fmt.Errorf("job crash: job = %#v, err = %#v", jr, r))
			}
		}()
		for {
			select {
			case <-w.closeChan:
				return
			case j := <-w.Jobs:
				{
					jr = j
					w.SetRunning(true)
					ctx, cancel := context.WithTimeout(context.TODO(), w.d.MaxJobRunningTimeout)
					if err := j.Fn(ctx, j.Arg); err != nil {
						w.d.monitor(fmt.Errorf("job = %#v, err = %#v", j, err))
					}
					cancel()
					w.SetLastBusyTime()
					w.SetRunning(false)
					w.d.WorkerPool <- w
				}
			}
		}
	}()
}

func (w *Worker) SetRunning(b bool) {
	w.guard.Lock()
	defer w.guard.Unlock()
	w.running = b
}

func (w *Worker) Running() bool {
	w.guard.RLock()
	defer w.guard.RUnlock()
	return w.running
}

func (w *Worker) canClose() bool {
	w.guard.RLock()
	defer w.guard.RUnlock()
	if !w.running && w.lastBusyTime.Add(w.workerIdleTimeout).Before(time.Now().UTC()) {
		return true
	}
	return false
}
