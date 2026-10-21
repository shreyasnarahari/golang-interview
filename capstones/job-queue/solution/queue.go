// Package solution is the reference implementation of the worker-pool /
// job-queue capstone: a bounded worker pool with context-aware submission,
// graceful draining shutdown, per-job panic recovery, and error collection.
//
// Shutdown safety without errgroup relies on two ordering rules:
//
//   - closing is closed (via closeOnce) BEFORE the internal mutex is taken,
//     so a Submit blocked on a full queue always has an escape hatch and
//     never holds its read lock forever.
//   - jobs is closed only while holding the write lock, after every
//     submitter has released its read lock, so no goroutine can ever send
//     on the closed jobs channel (which would panic).
package solution

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrPoolShutdown is returned by Submit once Shutdown has begun. After
// shutdown starts, no new jobs are accepted; queued and in-flight jobs are
// still drained.
var ErrPoolShutdown = errors.New("jobqueue: pool has been shut down")

// Job is a unit of work. The ctx passed to a job is the pool's job context:
// it is cancelled only when a Shutdown deadline forces the pool to stop
// waiting for in-flight work. A job that returns a non-nil error has that
// error recorded; see Pool.Errors. A job that panics is recovered by the
// pool and recorded as an error — one bad job never kills a worker.
type Job func(ctx context.Context) error

// Pool is a bounded worker pool fed by a channel queue. It is safe for
// concurrent use, but not reusable: once Shutdown has begun it stays shut
// down.
type Pool struct {
	jobs    chan Job
	closing chan struct{} // closed when shutdown begins; unblocks submitters
	done    chan struct{} // closed once every worker has exited

	closeOnce    sync.Once // guards close(closing)
	shutdownOnce sync.Once // guards closed=true and close(jobs)

	jobCtx    context.Context
	jobCancel context.CancelFunc
	workers   sync.WaitGroup

	mu     sync.RWMutex
	closed bool
	errs   []error
}

// New returns a Pool running workers goroutines. The pending-job queue
// holds at most workers jobs, so Submit applies backpressure once workers
// and the queue are both busy. It panics if workers is less than 1.
func New(workers int) *Pool {
	if workers < 1 {
		panic("jobqueue: New requires at least 1 worker")
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		jobs:      make(chan Job, workers),
		closing:   make(chan struct{}),
		done:      make(chan struct{}),
		jobCtx:    ctx,
		jobCancel: cancel,
	}
	p.workers.Add(workers)
	for i := 0; i < workers; i++ {
		go p.work()
	}
	go func() {
		p.workers.Wait()
		p.jobCancel()
		close(p.done)
	}()
	return p
}

// Submit enqueues job for execution and returns nil once the job is queued.
//
// Submit blocks while the queue is full until a worker frees a slot, the
// pool begins shutting down, or ctx is done — whichever comes first. It
// returns ErrPoolShutdown if the pool is (or becomes) shut down, and
// ctx.Err() if ctx expires while waiting.
func (p *Pool) Submit(ctx context.Context, job Job) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return ErrPoolShutdown
	}
	// Holding the read lock while sending guarantees jobs cannot be closed
	// mid-send: Shutdown closes jobs only after acquiring the write lock,
	// which waits for this read lock to be released. A blocked send is
	// released by the closing channel, which Shutdown closes before taking
	// the write lock — so there is no deadlock.
	select {
	case p.jobs <- job:
		return nil
	case <-p.closing:
		return ErrPoolShutdown
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown stops the pool from accepting new jobs and then drains it: every
// queued job runs to completion and Shutdown blocks until all in-flight
// jobs finish, returning nil. If ctx is done first, Shutdown cancels the
// job context and returns ctx.Err(); workers may still be finishing in the
// background. Shutdown is idempotent and safe for concurrent use.
func (p *Pool) Shutdown(ctx context.Context) error {
	// Signal submitters first, without the lock: any Submit blocked on a
	// full queue wakes up and releases its read lock.
	p.closeOnce.Do(func() { close(p.closing) })
	p.shutdownOnce.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.closed = true
		// No submitter holds the read lock while sending anymore, so no
		// send on the closed channel is possible. Workers drain the buffer
		// and exit when the channel is exhausted.
		close(p.jobs)
	})
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		p.jobCancel()
		return ctx.Err()
	}
}

// Errors returns a snapshot of every error recorded so far: each non-nil
// error returned by a job, plus one synthesized error per job panic. The
// returned slice is a copy; mutating it does not affect the pool.
func (p *Pool) Errors() []error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]error(nil), p.errs...)
}

// work is the worker loop: pull jobs until the channel is closed and
// drained.
func (p *Pool) work() {
	defer p.workers.Done()
	for job := range p.jobs {
		p.run(job)
	}
}

// run executes one job, converting a panic into a recorded error so a bad
// job cannot take down its worker.
func (p *Pool) run(job Job) {
	defer func() {
		if r := recover(); r != nil {
			p.record(fmt.Errorf("jobqueue: job panicked: %v", r))
		}
	}()
	if err := job(p.jobCtx); err != nil {
		p.record(err)
	}
}

func (p *Pool) record(err error) {
	p.mu.Lock()
	p.errs = append(p.errs, err)
	p.mu.Unlock()
}
