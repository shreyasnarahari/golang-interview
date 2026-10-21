# Design: Distributed Job Scheduler / Queue

Design a system where clients submit jobs (send email, transcode video, run a report) and a fleet of workers executes them reliably. Go focus: worker pools, `context` cancellation, graceful shutdown.

---

## 1. Requirements

### Functional
- Enqueue a job with a type, payload, and optional delay or cron schedule.
- At-least-once execution with automatic retries and exponential backoff.
- Dead-letter queue (DLQ) for jobs that exhaust retries.
- Job status queryable by ID; cancelable before execution.
- Per-queue priority and isolation (bulk emails must not starve payment webhooks).

### Non-functional
- Durable: no accepted job is lost on a broker or worker crash.
- 100k+ jobs/s enqueue throughput; workers scale horizontally.
- Duplicate execution is possible (at-least-once) — handlers must be idempotent.
- p99 enqueue-to-start latency within seconds for the default queue.

## 2. API Sketch

```go
// Job is a unit of work submitted to the scheduler.
type Job struct {
	ID          string
	Queue       string
	Type        string
	Payload     []byte
	RunAt       time.Time // zero means immediately
	MaxAttempts int
}

// Enqueuer accepts jobs; Workers consume them.
type Enqueuer interface {
	Enqueue(ctx context.Context, j Job) (string, error)
	Status(ctx context.Context, id string) (JobStatus, error)
	Cancel(ctx context.Context, id string) error
}
```

## 3. High-Level Architecture

```
 ┌────────┐    ┌──────────┐    ┌─────────────────────────┐
 │clients │───▶│ API tier │───▶│ broker                  │
 └────────┘    └──────────┘    │  ready queues (per prio)│
                               │  delayed set (zset/heap)│
                               └────┬───────────┬────────┘
                          ┌─────────┘           │
                          ▼                     ▼
                   ┌─────────────┐       ┌─────────────┐
                   │ scheduler   │       │ worker      │───▶ ACK ──▶ status store
                   │ (promotes   │       │ fleets      │───▶ NACK ─▶ retry / DLQ
                   │ due jobs)   │       │ (leased pop)│
                   └─────────────┘       └─────────────┘
```

Broker options: Postgres with `SELECT ... FOR UPDATE SKIP LOCKED` (simple, transactional with the rest of the app), Redis Streams / sorted sets (fast, weaker durability), or Kafka (huge scale, no per-job leasing — you build it yourself).

## 4. Data Model

Job record: `{id, queue, type, payload, state, attempts, max_attempts, run_at, leased_by, lease_expires_at, created_at, updated_at}`. States: `pending → running → done | failed | dead`.

**At-least-once via leasing:** a worker pops a job by acquiring a lease (visibility timeout). On success it ACKs; on crash the lease expires and another worker re-acquires the job — so a job can run twice. Handlers must be idempotent: dedupe on a job-scoped idempotency key, or make effects naturally idempotent.

**Backoff:** `run_at = now + base·2^attempts + jitter` (jitter prevents a thundering herd of simultaneous retries). After `max_attempts`, move to the DLQ with the full failure history for inspection and manual replay.

## 5. Scaling & Bottlenecks

- **Polling contention:** `SKIP LOCKED` lets hundreds of workers poll one Postgres table without blocking each other — the standard answer to "why not just use the database".
- **Delayed jobs:** a scheduler process promotes due jobs (`run_at <= now`) into ready queues; make promotion idempotent so you can run two schedulers for HA.
- **Long-running jobs:** heartbeats extend the lease; a dead worker's lease expires and the job is retried.
- **Poison messages:** never retry a job that fails deterministically — bound `max_attempts` and alert on DLQ growth.
- **Queue isolation:** separate worker pools per queue so a backlog in one queue cannot starve another.

## 6. Go-Specific Deep Dives

- The single-node version is [capstones/job-queue](../../capstones/job-queue/): an in-process queue with a worker pool, retries, and graceful shutdown. Walk through it end to end.
- **Worker pool:** fixed N goroutines ranging over a jobs channel; size N from downstream capacity, not guesswork. See the worker-pool and fan-out patterns in [core-topics/06](../../core-topics/06-goroutines-channels-concurrency/) and [core-topics/07](../../core-topics/07-concurrency-patterns-sync-package/).
- **Graceful shutdown:** `signal.NotifyContext` catches SIGTERM; stop accepting new jobs, cancel the root `context`, workers finish the in-flight job and ACK it, `sync.WaitGroup` waits for drain before exit. Cancellation must reach the *handler* — the whole tree is covered in [core-topics/08](../../core-topics/08-context-cancellation-timeouts/).
- **Per-attempt timeouts:** wrap each attempt in `context.WithTimeout`; a hung handler must not hold a worker forever.
- **Delayed execution:** a single `container/heap` of `run_at` timestamps plus one `time.Timer` re-armed to the heap head — never one timer per job.
- **Backoff with jitter:** `time.Duration(rand.Int64N(int64(backoff)))` added to the exponential base; know why pure exponential backoff without jitter re-synchronizes failures.
- **Supervision:** `errgroup`-style structured concurrency so one panicking worker (recovered) doesn't silently shrink the pool — relaunch it and count restarts as a metric.

## What the Interviewer Is Listening For

- You volunteer that at-least-once means duplicates, and handlers must be idempotent.
- Lease/visibility-timeout mechanics, not just "the queue handles it".
- A shutdown story that ACKs in-flight work rather than dropping it.
- Backoff *with jitter*, and a DLQ with a replay path.
