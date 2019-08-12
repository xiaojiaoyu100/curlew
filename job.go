package curlew

import (
	"context"
)

type Handle func(ctx context.Context, arg interface{}) error

type Job struct {
	Fn  Handle
	Arg interface{}
}

func NewJob() *Job {
	job := new(Job)
	return job
}
