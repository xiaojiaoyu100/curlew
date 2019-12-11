package curlew

import (
	"context"
	"fmt"
	"runtime/debug"
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
					done := make(chan struct{})
					go func() {
						defer func() {
							w.d.WorkerPool <- w
							done <- struct{}{}
							if r := recover(); r != nil {
								w.d.monitor(fmt.Errorf("job crash: job = %#v, err = %#v, debug = %s", 1, r, string(debug.Stack())))
							}
						}()
						ctx, cancel := context.WithTimeout(context.TODO(), w.d.MaxJobRunningTimeout)
						defer cancel()
						if err := j.Fn(ctx, j.Arg); err != nil {
							w.d.monitor(fmt.Errorf("job = %#v, err = %#v", j, err))
						}
					}()
					<-done
				}
			}
		}
	}()
}
