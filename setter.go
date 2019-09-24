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

// WithJobSize configures job buffer size.
func WithJobSize(size int) Setter {
	return func(d *Dispatcher) error {
		d.JobSize = size
		return nil
	}
}

// WithWorkerIdleTimeout configures worker idle timeout.
func WithWorkerIdleTimeout(t time.Duration) Setter {
	return func(d *Dispatcher) error {
		d.WorkerIdleTimeout = t
		return nil
	}
}

// WithMonitor configures a monitor.
func WithMonitor(monitor Monitor) Setter {
	return func(d *Dispatcher) error {
		d.monitor = monitor
		return nil
	}
}
