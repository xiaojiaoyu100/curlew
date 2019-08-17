package curlew

import (
	"context"
)

// Handler provides the job function signature.
type Handler func(ctx context.Context, arg interface{}) error

// Job defines a job.
type Job struct {
	Fn  Handler
	Arg interface{}
}

// NewJob creates a job instance.
func NewJob() *Job {
	job := new(Job)
	return job
}
