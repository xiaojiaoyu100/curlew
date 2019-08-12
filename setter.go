package curlew

import "time"

// Setter configures a Dispatcher.
type Setter func(d *Dispatcher) error

// WithMaxWorkerNum configures maximum number of workers in the pool.
func WithMaxWorkerNum(num int) Setter {
	return func(d *Dispatcher) error {
		d.MaxWorkerNum = num
		return nil
	}
}

func WithJobSize(size int) Setter {
	return func(d *Dispatcher) error {
		d.JobSize = size
		return nil
	}
}

func WithWorkerIdleTimeout(t time.Duration) Setter {
	return func(d *Dispatcher) error {
		d.WorkerIdleTimeout = t
		return nil
	}
}

func WithMonitor(monitor Monitor) Setter {
	return func(d *Dispatcher) error {
		d.monitor = monitor
		return nil
	}
}
