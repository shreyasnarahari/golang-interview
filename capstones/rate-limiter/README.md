# Capstone: Rate Limiter

Build a thread-safe, per-key rate-limiter library — the kind of component
that sits in front of an API gateway and decides, per user or per IP,
whether a request may proceed. It exercises the concurrency topics from
Level 3 (`sync.Mutex`, atomics, goroutine lifecycle) plus careful `time`
handling, in a package small enough to finish in one sitting.

## Requirements

1. **Token bucket** — `NewTokenBucket(rate int, per time.Duration)` returns a
   limiter whose `Allow(key string) bool` lets each key burst up to `rate`
   requests and then refills `rate` tokens per `per` interval. Refill is
   **lazy**: token counts are computed from elapsed wall-clock time on each
   access. No background goroutines.
2. **Sliding-window counter** — `NewSlidingWindow(limit int, window
   time.Duration)` returns a limiter allowing at most `limit` requests per
   key per trailing `window`, using the weighted two-generation-counter
   approximation.
3. **Eviction** — per-key state must not grow without bound. Keys idle
   longer than a caller-supplied duration are dropped, either lazily via
   `EvictIdle` or in the background via a `Sweeper` with `Stop()`.
4. **Concurrency safety** — every exported method is safe for concurrent
   use; the acceptance tests must pass under `go test -race`.

## API contract

```go
type Limiter interface {
	Allow(key string) bool
}

type IdleEvictor interface {
	EvictIdle(maxIdle time.Duration) int
}

func NewTokenBucket(rate int, per time.Duration) *TokenBucket  // panics on non-positive args
func (tb *TokenBucket) Allow(key string) bool                  // consumes one token if allowed
func (tb *TokenBucket) EvictIdle(maxIdle time.Duration) int
func (tb *TokenBucket) ActiveKeys() int

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow
func (sw *SlidingWindow) Allow(key string) bool
func (sw *SlidingWindow) EvictIdle(maxIdle time.Duration) int
func (sw *SlidingWindow) ActiveKeys() int

func NewSweeper(l IdleEvictor, interval, maxIdle time.Duration) *Sweeper
func (s *Sweeper) Stop()  // idempotent; blocks until the goroutine exits
```

Semantics that the tests pin down:

- A key's **first** request sees a full bucket / empty window — it is
  always allowed.
- Burst-then-deny: exactly `rate` (or `limit`) requests pass, the next is
  denied.
- Refill over time: after roughly one refill period at least one more
  request fits. Tests use tens of milliseconds; keep total sleeps small.
- Keys are fully independent.
- Concurrent hammering of one hot key admits **exactly** `rate`/`limit`
  requests when refill is negligible — no off-by-one, no lost updates.

## Working the exercise

`limiter.go` (package `ratelimiter`) is the stub: the full exported API
with doc comments, every body `panic("TODO: implement ...")`. The
acceptance tests in `limiter_test.go` are red by design. Make them green,
then compare with `solution/` (package `solution`), which passes the same
tests including `-race -count=3`.

```text
go test ./capstones/rate-limiter/            # red until you implement
go test -race ./capstones/rate-limiter/solution/
```

## Design notes and trade-offs

**Token bucket, lazy refill.** Store per key `(tokens float64, lastSeen
time.Time)`; on access compute `tokens += elapsed * rate`, clamp to the
burst size, then try to spend one token. Float tokens avoid a separate
"remainder" timestamp. One `sync.Mutex` guards the whole map: the critical
section is a few arithmetic ops, so a single mutex beats per-key locks
until key counts get huge. Sharding the map (`map[shard]struct{mu; m}`) is
the standard next step if a profile says otherwise.

**Sliding window, weighted approximation.** An exact sliding window needs a
log of timestamps per key — O(n) memory and GC pressure. The standard
approximation keeps two fixed-window counters per key:

```text
estimate = previousWindowCount * (1 - elapsed/window) + currentWindowCount
```

The previous window is weighted by how much of it still overlaps the
trailing window. O(1) memory per key, at the cost of being approximate
near window boundaries — usually fine for abuse prevention, wrong for
billing.

**Eviction.** Unbounded key sets are the classic production bug in
hand-rolled limiters (one map entry per attacking IP forever). Two
options: lazy eviction on access (zero goroutines, but untouched keys
linger) or a `Sweeper` goroutine on a `time.Ticker`. The sweeper shows the
full lifecycle idiom: `done` channel + `sync.Once` for idempotent `Stop` +
`sync.WaitGroup` so `Stop` blocks until the goroutine has exited.

**Time.** `time.Now()` in the hot path is fine (vDSO, no syscall on
Linux). `time.Time` carries a monotonic reading, so `Sub` is immune to
wall-clock jumps — mention this; interviewers ask.

## Stretch goals

Benchmarks live in `solution/bench_test.go`:

```text
go test -run=^$ -bench=. -benchmem ./capstones/rate-limiter/solution/
```

- Show the hot path is **0 allocs/op** once a key's bucket exists (the map
  stores `*bucket`, so only the first sighting of a key allocates).
- Compare `sync.Mutex` vs an atomic-only fixed-window counter
  (`BenchmarkFixedWindowMutex` vs `BenchmarkFixedWindowAtomic`). The
  atomic version wins modestly under contention but is approximate at
  window roll-over (a CAS roll can clobber an in-flight increment) — a
  good illustration that lock-free code trades exactness and complexity
  for throughput.
- Try `b.RunParallel` with many distinct keys and watch the single mutex
  become the bottleneck; then sketch map sharding.

## What the interviewer is listening for

- **Lazy refill over a ticker per key** — you don't spawn a goroutine per
  bucket; elapsed-time arithmetic is O(1) and idle keys cost nothing.
- **Correct concurrency story**: one mutex protecting map + bucket state,
  short critical sections, no locks held across I/O; know when you'd
  reach for sharding or `sync.Map` (rarely — say why).
- **Memory discipline**: unbounded maps are a leak; eviction policy (idle
  TTL), and why lazy-only eviction still bounds memory for *active*
  traffic but not for one-shot scans.
- **Algorithm trade-offs**: fixed window (cheap, boundary bursts of 2x),
  sliding log (exact, O(n) memory), sliding counter (the O(1)
  compromise), token bucket (smooth bursts), leaky bucket (smooth
  outflow, queues).
- **Race hygiene**: `go test -race` clean, exact-count concurrent tests
  instead of sleeps-and-hopes, idempotent `Stop`, no goroutine leaks.
- **Precision pitfalls**: float token math is fine here (deny-side errors
  only), but know that exact accounting would use integer
  nanoseconds/tokens.
