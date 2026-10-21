# Capstone: Job Queue / Worker Pool

Build a bounded worker-pool library in Go. This is the concurrency capstone:
it exercises channels as queues, `context.Context` cancellation, graceful
shutdown, panic isolation, and goroutine lifecycle management in one small,
interview-sized API.

**Skills drilled:** buffered channels, `select` with multiple cases,
`sync.Once` / `sync.WaitGroup` / `sync.RWMutex`, context deadlines,
`recover`, race-detector-driven development, goroutine leak detection.

**Constraint:** standard library only. In particular,
`golang.org/x/sync/errgroup` is **not** allowed — implementing that
machinery yourself is the point of the exercise.

## Layout

```
capstones/job-queue/
├── queue.go            # exercise stub: full API + doc comments, bodies panic TODO
├── queue_test.go       # acceptance tests (RED until you implement the stub)
├── solution/
│   ├── queue.go        # reference implementation (GREEN, race-clean)
│   └── queue_test.go   # the same tests
└── README.md           # this spec
```

Run the exercise tests from the repo root:

```sh
go test ./capstones/job-queue/            # red until implemented
go test -race -count=3 ./capstones/job-queue/solution/
```

## Requirements

### 1. Construction — `New(workers int) *Pool`

Create a pool with a fixed number of worker goroutines and a **bounded**
pending-job queue. The reference implementation sizes the queue at
`workers` pending jobs; any fixed bound derived from `workers` is fine, but
document your choice. Panic on `workers < 1` — a pool with no workers can
never drain and would silently deadlock `Shutdown`.

### 2. Submission — `Submit(ctx context.Context, job Job) error`

```go
type Job func(ctx context.Context) error
```

`Submit` must respect the queue bound. **This capstone chooses blocking
with backpressure**: when the queue is full, `Submit` blocks until a worker
frees a slot, the pool starts shutting down, or `ctx` expires. It returns:

- `nil` once the job is queued,
- `ErrPoolShutdown` if shutdown has begun (before or during the call),
- `ctx.Err()` if the context is cancelled or its deadline passes while waiting.

### 3. Graceful shutdown — `Shutdown(ctx context.Context) error`

Semantics, in order:

1. **Stop accepting** new jobs — concurrent and future `Submit` calls return
   `ErrPoolShutdown`.
2. **Drain** — every job already in the queue runs to completion.
3. **Wait** — block until all in-flight jobs finish, then return `nil`.
4. **Respect the deadline** — if `ctx` expires first, cancel the job context
   (so cooperative jobs can abort) and return `ctx.Err()`. Workers may still
   be winding down in the background; that is the caller's accepted trade-off.

`Shutdown` must be idempotent and safe to call concurrently with `Submit`.

### 4. Error collection — `Errors() []error`

Every non-nil error returned by a job is recorded. `Errors` returns a
snapshot copy in arrival order. (An errgroup-style "first error wins" API is
a stretch goal; the full list is more useful for a job queue, where one
failure usually shouldn't cancel siblings.)

### 5. Panic isolation

A job that panics must not kill its worker goroutine. Recover inside the
worker, record the panic as an error (include the panic value in the
message), and keep processing. A later job must still run.

## API contract

```go
var ErrPoolShutdown = errors.New("jobqueue: pool has been shut down")

type Job func(ctx context.Context) error

func New(workers int) *Pool
func (p *Pool) Submit(ctx context.Context, job Job) error
func (p *Pool) Shutdown(ctx context.Context) error
func (p *Pool) Errors() []error
```

Guarantees the acceptance tests pin down:

- every accepted job is processed exactly once, even during `Shutdown` drain;
- job errors are all retrievable via `Errors()`;
- `Shutdown` returns `nil` only after the queue is empty and in-flight jobs
  have finished (structural drain — no sleeps required to observe it);
- a `Shutdown` context deadline yields `context.DeadlineExceeded`;
- a panicking job is recorded as an error and the pool keeps working;
- `Submit` on a full pool blocks, and a deadline on its context unblocks it
  with `context.DeadlineExceeded`;
- after `Shutdown` returns `nil`, the pool's goroutines are gone
  (`runtime.NumGoroutine` returns to baseline);
- all of the above is clean under `go test -race`.

## Design trade-offs to think (and talk) about

**Blocking `Submit` vs. `ErrQueueFull`.** This spec picks blocking with
context-aware backpressure. Justification: a job queue sits between a
producer and scarce workers; pushing back on the producer is usually the
correct systems answer (it propagates load instead of dropping work), and
`ctx` gives the caller a clean escape hatch (deadline → give up). The
`ErrQueueFull` alternative is better when the caller has somewhere else to
put the work (spill to disk, retry queue) or must never stall — e.g. an HTTP
handler that should return 503 fast. Be ready to defend either and to say
how you'd add `TrySubmit` on top.

**Queue capacity.** Bounded is non-negotiable in production: an unbounded
queue turns backpressure into an OOM. Sizing it at `workers` keeps the
default conservative; real systems size it from latency/throughput targets.

**Where the `jobs` channel close happens.** The classic bug: `Shutdown`
closes the jobs channel while a `Submit` is mid-send → panic. The reference
solves it with an RWMutex: submitters hold the read lock across the send,
and shutdown closes a separate `closing` channel *before* taking the write
lock, so blocked submitters always wake up. Alternatives: a dispatcher
goroutine that owns the queue, or an atomic flag + `recover` around the
send (works, but is a smell).

**What context does a job get?** Here jobs receive a pool-owned context
cancelled only when a shutdown deadline is forced. Passing the *Submit*
context to the job is a defensible alternative (per-job cancellation), but
it surprises callers who expect submit-time deadlines to stop applying once
the job starts.

## Stretch goals

1. `TrySubmit(job Job) error` — non-blocking submit returning `ErrQueueFull`.
2. errgroup mode: first error cancels the job context for siblings.
3. Per-job timeout enforced by the pool.
4. Metrics: `Submitted()`, `Completed()`, `Failed()`, queue depth, per-job
   latency histogram.
5. Priority queues (two channels, `select` with a default fallback).
6. Dynamic worker scaling between a min and max based on queue depth.
7. `SubmitAndWait(ctx, job) error` returning that specific job's error —
   think about how to route results back (per-submission result channel).

## What the interviewer is listening for

- **Backpressure vocabulary:** you can explain why the queue is bounded and
  what "blocking submit" buys you over "drop when full" — and when you'd
  choose the opposite.
- **The closed-channel-send panic:** you know sending on a closed channel
  panics, and you can describe a concrete protocol that makes it impossible
  (lock ordering, dedicated `closing` channel, or a queue-owning goroutine).
- **Drain vs. force-quit:** `Shutdown` draining queued work, the deadline as
  the escape hatch, and what happens to in-flight jobs after a forced return
  (cooperative cancellation via the job context — jobs must check `ctx`).
- **Panic hygiene:** recover belongs at the goroutine boundary (per job),
  not scattered inside jobs; an unrecovered panic in any goroutine kills the
  whole process.
- **Leak awareness:** every goroutine you spawn has an exit path;
  `runtime.NumGoroutine` deltas and `-race` are part of "done", not extras.
- **Honest limitations:** exactly-once processing per accepted job, no
  persistence (a crash loses the queue), fairness is FIFO-by-luck, not
  guaranteed.
