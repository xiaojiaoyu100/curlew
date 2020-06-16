package curlew

import (
	"context"
	"fmt"
	"sync"
)

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
					ctx, cancel := context.WithTimeout(context.TODO(), w.d.MaxJobRunningTimeout)
					if err := j.Fn(ctx, j.Arg); err != nil {
						w.d.monitor(fmt.Errorf("job = %#v, err = %#v", j, err))
					}
					cancel()
					w.d.WorkerPool <- w
				}
			}
		}
	}()
}
