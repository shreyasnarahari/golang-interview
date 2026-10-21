// Package jobqueue is the exercise stub for the worker-pool / job-queue
// capstone. Implement every method on Pool until the acceptance tests in
// queue_test.go go green, then compare your work against the reference
// implementation in solution/.
//
// You will build a bounded worker pool that teaches three concurrency
// fundamentals at once: channels as queues, context-driven cancellation,
// and graceful shutdown with draining.
package jobqueue

import (
	"context"
	"errors"
)

// ErrPoolShutdown is returned by Submit once Shutdown has begun. After
// shutdown starts, no new jobs are accepted; queued and in-flight jobs are
// still drained.
var ErrPoolShutdown = errors.New("jobqueue: pool has been shut down")

// Job is a unit of work. The ctx passed to a job is the pool's job context:
// it is cancelled only when a Shutdown deadline forces the pool to stop
// waiting for in-flight work. A job that returns a non-nil error has that
// error recorded; see Pool.Errors. A job that panics is recovered by the
// pool and recorded as an error — one bad job must never kill a worker.
type Job func(ctx context.Context) error

// Pool is a bounded worker pool. Jobs are submitted through Submit and
// executed by a fixed set of worker goroutines fed by a bounded queue.
//
// The pool is safe for concurrent use by multiple submitters, but it is not
// reusable: once Shutdown has begun the pool stays shut down.
type Pool struct {
	// TODO: add unexported state (jobs channel, closing/done channels,
	// sync.Once pair for shutdown, mutex guarding the closed flag and the
	// collected errors, job context, worker WaitGroup).
}

// New returns a Pool running workers goroutines. The pending-job queue is
// bounded: it holds at most workers jobs not yet picked up by a worker, so
// Submit applies backpressure once workers and the queue are both busy. It
// panics if workers is less than 1.
func New(workers int) *Pool {
	panic("TODO: implement New")
}

// Submit enqueues job for execution and returns nil once the job is queued.
//
// Submit blocks while the queue is full (backpressure) until a worker frees
// a slot, the pool begins shutting down, or ctx is done — whichever comes
// first. It returns ErrPoolShutdown if the pool is (or becomes) shut down,
// and ctx.Err() if ctx is cancelled or its deadline passes while waiting.
func (p *Pool) Submit(ctx context.Context, job Job) error {
	panic("TODO: implement (*Pool).Submit")
}

// Shutdown stops the pool from accepting new jobs and then drains it: every
// job already queued runs to completion, and Shutdown blocks until all
// in-flight jobs finish. It returns nil once the pool is fully drained.
//
// If ctx is done before draining completes, Shutdown cancels the job context
// (asking in-flight jobs to abort) and returns ctx.Err(); the workers may
// still be finishing in the background. Shutdown is idempotent and safe to
// call concurrently with Submit.
func (p *Pool) Shutdown(ctx context.Context) error {
	panic("TODO: implement (*Pool).Shutdown")
}

// Errors returns a snapshot of every error recorded so far: each non-nil
// error returned by a job, plus one synthesized error per job panic. The
// returned slice is a copy; mutating it does not affect the pool.
func (p *Pool) Errors() []error {
	panic("TODO: implement (*Pool).Errors")
}
