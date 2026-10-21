# Design: Distributed Rate Limiter

Design a rate-limiting service that protects backend APIs from overload and abuse. Go focus: `sync` primitives, sharded in-memory state, and the exactness-vs-availability trade-off.

---

## 1. Requirements

### Functional
- Limit requests per key (user ID, API key, IP) against configurable rules: `100 req/min`, `10 req/s with burst 20`.
- Return allow/deny plus remaining quota and retry-after so clients can back off.
- Support multiple algorithms (fixed window, sliding window, token bucket).
- Rules updatable at runtime without redeploy.

### Non-functional
- < 1 ms p99 added latency — the limiter sits inline on the request path.
- Highly available: define fail-open vs fail-closed **per rule** (fail-open on a login endpoint is a security hole; fail-closed on a read feed is an outage).
- Accurate *enough*: hard accuracy for payments, approximate for public APIs.
- Scales to 100k+ checks/s.

## 2. API Sketch

```go
// Decision is the result of one rate-limit check.
type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

// Limiter checks whether key may proceed under its matching rule.
type Limiter interface {
	Allow(ctx context.Context, key string) (Decision, error)
}
```

Rules live in a config service, are matched by key pattern (e.g. `apikey:*`), and are cached locally in each limiter node with a short TTL.

## 3. High-Level Architecture

```
        ┌──────────┐
        │ clients  │
        └────┬─────┘
             ▼
      ┌─────────────┐      ┌───────────────┐
      │ API servers │─────▶│ rate-limiter  │ (library or sidecar)
      └─────────────┘      │ nodes         │
                           └──────┬────────┘
              ┌───────────────────┼───────────────────┐
              ▼                   ▼                   ▼
        ┌───────────┐       ┌───────────┐       ┌───────────┐
        │ Redis     │       │ gossip    │       │ config    │
        │ (Lua      │       │ protocol  │       │ service   │
        │ counters) │       │ (deltas)  │       │ (rules)   │
        └───────────┘       └───────────┘       └───────────┘
         exact, ~1ms RTT     approximate, local    runtime updates
```

## 4. Data Model & Algorithms

Token bucket state per key in Redis: hash `rl:{key}` = `{tokens, last_refill_ms}` with `EXPIRE` set to the refill period so idle keys self-clean. Refill-and-deduct runs inside a single Lua script so the read-modify-write is atomic across all limiter nodes — no `WATCH`/`MULTI` round trips. Use `redis.call('TIME')` inside the script; never trust client clocks.

| Algorithm | Memory / key | Burst behavior | Notes |
|---|---|---|---|
| Fixed window counter | 1 int | 2× limit at window edges | simplest; edge burst acceptable for coarse limits |
| Sliding window log | 1 timestamp / request | exact | memory blows up at high limits |
| Sliding window counter | 2 ints | approximate | weighted sum of current + previous window |
| Token bucket | 2 values | controlled burst up to capacity | the usual default |
| Leaky bucket | 1 virtual finish time | smooths to constant rate | egress shaping |

## 5. Two Deployment Models

**Centralized (Redis + Lua).** One round trip per check; all nodes share exact counts. Hot-key risk: a viral key pins one Redis shard. Mitigate with token batching — each node pre-fetches 100 tokens and serves locally until exhausted, trading exactness for load.

**In-memory + gossip.** Each node keeps a local sharded map of buckets; nodes exchange counter deltas over a gossip protocol (HashiCorp `memberlist`-style) every ~100 ms and converge on approximate global counts. Zero network latency on the hot path and survives a Redis outage, but over-admission is bounded by gossip interval × aggregate rate. This is roughly how edge networks (e.g. Cloudflare) do it.

## 6. Scaling & Bottlenecks

- **Redis throughput:** a shard sustains ~100k short Lua ops/s; shard by key and pipeline checks from the client.
- **Memory:** ~50 bytes per active key; `EXPIRE` bounds growth to keys active within a window.
- **Failover:** if Redis is unreachable, the per-rule fail-open/fail-closed policy decides — make it explicit in the rule config, not a global accident.

## 7. Go-Specific Deep Dives

- The single-node version is [capstones/rate-limiter](../../capstones/rate-limiter/): a token bucket guarded by `sync.Mutex` with an injectable clock for deterministic tests. Be ready to write it on a whiteboard.
- **Sharded map for in-memory buckets:** one `map[string]*bucket` behind a single `RWMutex` convoys under write contention. Shard by `fnv(key) % nShards` (32–256 shards), one `Mutex` per shard. Maps are never safe for concurrent access — see [core-topics/02](../../core-topics/02-arrays-slices-maps-internals/) and the mutex trade-offs in [core-topics/07](../../core-topics/07-concurrency-patterns-sync-package/).
- **`sync/atomic` for fixed-window counters:** `atomic.Int64.Add(1)` beats a mutex for a single hot counter; know the false-sharing caveat — pad adjacent counters to separate cache lines.
- **Lazy refill, no per-key goroutines:** compute elapsed time on each `Allow` call instead of running a ticker per key; a ticker per key is a goroutine-and-timer leak at scale ([core-topics/06](../../core-topics/06-goroutines-channels-concurrency/)).
- **One janitor, not many timers:** a single background goroutine sweeping shards on an interval beats per-key `time.AfterFunc` when key counts are high — ties into GC pressure from [core-topics/12](../../core-topics/12-garbage-collection-memory-management/).
- **`golang.org/x/time/rate`** is the canonical production token bucket (outside the stdlib); know its `Allow` / `Wait` / `Reserve` semantics even if you implement your own in the interview.

## What the Interviewer Is Listening For

- You state the fail-open/fail-closed policy unprompted.
- You can explain *why* the Lua script must be atomic (two nodes refilling concurrently).
- You pick in-memory + gossip when the requirement is "approximate but never down", and centralized when it is "exact".
- You reach for lazy refill instead of per-key timers without being prompted.
