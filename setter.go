package curlew

// Setter configures a Dispatcher.
type Setter func(d *Dispatcher) error

// WithMaxWorkerNum configures maximum number of workers in the pool.
func WithMaxWorkerNum(num int) Setter {
	return func(d *Dispatcher) error {
		d.MaxWorkerNum = num
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
