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
	close             chan struct{}
}

func NewWorker(d *Dispatcher) *Worker {
	w := new(Worker)
	w.Jobs = make(chan *Job, 1)
	w.workerIdleTimeout = d.WorkerIdleTimeout
	w.close = make(chan struct{})
	w.schedule()
	w.d = d
	d.add(w)
	d.WorkerPool <- w
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

func (w *Worker) submit(job *Job) {
	w.Jobs <- job
}

func (w *Worker) schedule() {
	go func() {
		ticker := time.NewTicker(w.workerIdleTimeout)
		defer ticker.Stop()
		for range ticker.C {
			if w.canClose() {
				close(w.close)
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
			case <-w.close:
				w.d.remove(w)
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
