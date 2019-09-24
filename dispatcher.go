package curlew

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Dispatcher takes the responsibility of dispatching jobs to available workers.
type Dispatcher struct {
	guard                sync.RWMutex
	MaxWorkerNum         int           // maximum  worker num in the pool
	JobSize              int           // job buffer size
	WorkerIdleTimeout    time.Duration // worker
	MaxJobRunningTimeout time.Duration // job execution timeout
	WorkerPool           chan *Worker  // worker pool
	workers              map[*Worker]struct{}
	jobs                 chan *Job
	runningWorkerNum     int
	monitor              Monitor
	logger               *logrus.Logger
}

// New creates a dispatcher instance.
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

	if d.monitor == nil {
		return nil, errors.New("no monitor provided")
	}

	d.WorkerPool = make(chan *Worker, d.MaxWorkerNum)
	d.workers = make(map[*Worker]struct{})
	d.jobs = make(chan *Job, d.JobSize)

	d.dispatch()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	d.logger = logger

	return &d, nil
}

// EnableDebug enables debug info.
func (d *Dispatcher) EnableDebug() {
	d.logger.SetLevel(logrus.DebugLevel)
}

// Submit submits a job.
func (d *Dispatcher) Submit(j *Job) {
	if j == nil {
		return
	}
	d.jobs <- j
}

// RunningWorkerNum returns the current running worker num.
func (d *Dispatcher) RunningWorkerNum() int {
	d.guard.RLock()
	defer d.guard.RUnlock()
	return d.runningWorkerNum
}

func (d *Dispatcher) add(w *Worker) {
	d.guard.Lock()
	d.workers[w] = struct{}{}
	d.runningWorkerNum++
	d.guard.Unlock()
}

func (d *Dispatcher) remove(w *Worker) {
	d.guard.Lock()
	delete(d.workers, w)
	d.runningWorkerNum--
	d.guard.Unlock()
}

func (d *Dispatcher) dispatch() {
	go func() {
		for j := range d.jobs {
			select {
			case w := <-d.WorkerPool:
				if w.IsClosed() {
					d.logger.Debug("worker is closed")
					NewWorker(d).submit(j)
				} else {
					w.submit(j)
				}
			default:
				if d.RunningWorkerNum() < d.MaxWorkerNum {
					d.logger.Debug("not reach limit yet, create new worker")
					NewWorker(d).submit(j)
				} else {
					d.logger.Debug("reach limit, wait a ready worker")
					w := <-d.WorkerPool
					w.submit(j)
				}
			}
		}
	}()
}
