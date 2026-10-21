# System Design Walkthroughs (Go Track)

Three Go-centric system-design walkthroughs for senior backend/infrastructure interviews. Each covers: requirements, API sketch, architecture, data model, scaling, and the exact Go primitives an interviewer is listening for — cross-referenced to the `core-topics/` modules that teach those foundations.

## How to Use These in an Interview

1. **Clarify scope first** (2–3 min): QPS, read/write ratio, consistency requirements, latency budget. Never start drawing boxes before this.
2. **API sketch** (2 min): the interface narrows the design space and shows you think in contracts.
3. **High-level architecture** (10 min): boxes, arrows, data flow. Say what each box owns.
4. **Deep dive** (rest of the time): the interviewer picks a box. This is where the Go-specific sections below pay off — they want to hear `sync.RWMutex` sharding, `context` cancellation trees, GC tuning, not just "we use Redis".

## Walkthroughs

| Design Problem | Core Trade-off | Go Deep Dive Focus | Walkthrough |
|---|---|---|---|
| Distributed Rate Limiter | Exactness (centralized store) vs latency/availability (local + gossip) | Sharded maps, `sync` primitives, atomics, token bucket | [distributed-rate-limiter.md](distributed-rate-limiter.md) |
| Distributed Job Scheduler | At-least-once delivery vs duplicate work; throughput vs ordering | Worker pools, `context` cancellation, graceful shutdown, backoff | [job-scheduler.md](job-scheduler.md) |
| LRU Cache Service | Hit ratio vs memory; lock contention vs consistency | `container/list` LRU, `sync.Map` vs `RWMutex`, GC pressure | [lru-cache-service.md](lru-cache-service.md) |

## Companion Code

Each walkthrough has a buildable single-node companion in `../../capstones/`:

- [capstones/rate-limiter](../../capstones/rate-limiter/) — token bucket + sliding window, the single-node version of the distributed limiter.
- [capstones/job-queue](../../capstones/job-queue/) — worker pool with retries and graceful shutdown, the single-node version of the scheduler.
- [capstones/kv-store](../../capstones/kv-store/) — in-memory store with TTL, the storage half of the cache service.

Implement the capstone first, then use the walkthrough to practice scaling it out loud.
