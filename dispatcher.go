package curlew

import (
	"errors"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Dispatcher takes the responsibility of dispatching jobs to available workers.
type Dispatcher struct {
	guard                sync.RWMutex
	MaxWorkerNum         int           // maximum  worker num in the pool
	MaxJobRunningTimeout time.Duration // job execution timeout
	WorkerPool           chan *Worker  // worker pool
	workers              map[*Worker]struct{}
	jobs                 chan *Job
	runningWorkerNum     int
	monitor              Monitor
	logger               *logrus.Logger
}

func defaultWorkerNum() int {
	num := runtime.NumCPU()
	if num <= 0 {
		return 2
	}
	return 2 * num
}

// New creates a dispatcher instance.
func New(setters ...Setter) (*Dispatcher, error) {
	d := Dispatcher{
		MaxWorkerNum:         defaultWorkerNum(),
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

	if d.monitor == nil {
		return nil, errors.New("no monitor provided")
	}

	d.WorkerPool = make(chan *Worker, d.MaxWorkerNum)
	d.workers = make(map[*Worker]struct{})
	d.jobs = make(chan *Job, 1)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	d.logger = logger

	d.dispatch()

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

// NewWorker 生成一个工作协程
func (d *Dispatcher) NewWorker() *Worker {
	w := newWorker(d)
	d.add(w)
	return w
}

func (d *Dispatcher) dispatch() {
	go func() {
		for j := range d.jobs {
			select {
			case w := <-d.WorkerPool:
				d.logger.Debug("Worker is ready to submit a task.")
				w.submit(j)
			default:
				if d.RunningWorkerNum() < d.MaxWorkerNum {
					d.logger.Debug("Not reach limit yet, create a new worker to submit a task.")
					d.NewWorker().submit(j)
				} else {
					w := <-d.WorkerPool
					d.logger.Debug("Reach limit, wait a ready worker")
					w.submit(j)
				}
			}
		}
	}()
}
