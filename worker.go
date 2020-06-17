package curlew

import (
	"context"
	"fmt"
	"sync"
)

// Worker 工作协程
type Worker struct {
	guard sync.RWMutex
	d     *Dispatcher
	Jobs  chan *Job
}

func newWorker(d *Dispatcher) *Worker {
	w := &Worker{}
	w.d = d
	w.Jobs = make(chan *Job, 1)
	w.schedule()
	return w
}

func (w *Worker) submit(j *Job) {
	if j == nil {
		return
	}
	w.Jobs <- j
}

func (w *Worker) schedule() {
	go func() {
		for {
			select {
			case j := <-w.Jobs:
				{
					w.exec(j)
				}
			}
		}
	}()
}

func (w *Worker) exec(j *Job) {
	ctx, cancel := context.WithTimeout(context.TODO(), w.d.MaxJobRunningTimeout)
	defer func() {
		if r := recover(); r != nil {
			w.d.monitor(fmt.Errorf("fn panic: job = %#v, recover() = %#v", j, r))
		}
		cancel()
	}()
	if err := j.Fn(ctx, j.Arg); err != nil {
		w.d.monitor(fmt.Errorf("job = %#v, err = %#v", j, err))
	}
	w.d.WorkerPool <- w
}
