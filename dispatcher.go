package curlew

import (
	"errors"
	"sync"
	"time"
)

type Dispatcher struct {
	guard                sync.RWMutex
	MaxWorkerNum         int
	JobSize              int
	WorkerIdleTimeout    time.Duration
	MaxJobRunningTimeout time.Duration
	WorkerPool           chan *Worker
	workers              map[*Worker]struct{}
	jobs                 chan *Job
	runningWorkerNum     int
	monitor              Monitor
}

func New(setters ...Setter) (*Dispatcher, error) {
	d := Dispatcher{
		MaxWorkerNum:         16,
		JobSize:              16,
		WorkerIdleTimeout:    time.Second * 60,
		MaxJobRunningTimeout: 10 * time.Second,
	}

	for _, setter := range setters {
		if err := setter(&d); err != nil {
			return nil, err
		}
	}

	if d.MaxWorkerNum < 1 {
		return nil, errors.New("must have at least one worker in the pool")
	}

	if d.JobSize < 1 {
		return nil, errors.New("must have at least one job buffered in the channel")
	}

	d.WorkerPool = make(chan *Worker, d.MaxWorkerNum)
	d.workers = make(map[*Worker]struct{})
	d.jobs = make(chan *Job, d.JobSize)
	d.dispatch()

	return &d, nil
}

func (d *Dispatcher) SubmitAsync(j *Job) {
	go func() {
		if j == nil {
			return
		}
		d.jobs <- j
	}()
}

func (d *Dispatcher) RunningWorkerNum() int {
	d.guard.RLock()
	defer d.guard.RUnlock()
	return d.runningWorkerNum
}

func (d *Dispatcher) Add(w *Worker) {
	d.guard.Lock()
	d.workers[w] = struct{}{}
	d.runningWorkerNum++
	d.guard.Unlock()
}

func (d *Dispatcher) Remove(w *Worker) {
	d.guard.Lock()
	delete(d.workers, w)
	d.runningWorkerNum--
	d.guard.Unlock()
}

func (d *Dispatcher) dispatch() {
	go func() {
		for {
			select {
			case j := <-d.jobs:
				select {
				case w := <-d.WorkerPool:
					w.submitAsync(j)
				default:
					if d.RunningWorkerNum() < d.MaxWorkerNum {
						NewWorker(d)
					}
					d.SubmitAsync(j)
				}
			}
		}
	}()
}
