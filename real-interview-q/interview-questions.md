# Real Company Go Interview Questions & Solutions

This file serves as a master index of company-specific Go interview question banks. View the full company-specific question files below for detailed solutions, theory breakdowns, and runnable Go code implementations:

## 🏢 Company Question Files Index

- 🔵 **Google**: [google.md](google.md) — `iface` nil interface trap, distributed context timeout propagation, goroutine leak detection, concurrent LRU cache.
- 🔴 **Uber**: [uber.md](uber.md) — Token Bucket rate limiter, `sync/atomic` vs `sync.Mutex` contention, worker pools with graceful context cancellation.
- ☁️ **Cloudflare**: [cloudflare.md](cloudflare.md) — Zero-copy `string`/`[]byte` conversions via `unsafe`, `sync.Pool` byte buffer recycling, Fan-In event aggregators.
- ⚙️ **HashiCorp**: [hashicorp.md](hashicorp.md) — State machine broadcasts via `sync.Cond`, in-memory WAL key-value store, Raft consensus mechanics.
- ☸️ **Kubernetes / CNCF**: [kubernetes-cncf.md](kubernetes-cncf.md) — Reconciliation loops, workqueues with backoff retries, `GOMEMLIMIT` container memory tuning.
- 🪳 **CockroachDB / TiDB**: [cockroachdb.md](cockroachdb.md) — Zero-allocation binary key encoders, concurrent MVCC (Multi-Version Concurrency Control) engines.
- 🐶 **Datadog**: [datadog.md](datadog.md) — Lockless metric aggregators, heap memory allocation tracking via `runtime.MemStats`.
- 💳 **Stripe**: [stripe.md](stripe.md) — Idempotency key payment middleware, error wrapping with `fmt.Errorf("%w")` and `errors.As`.
- 🎵 **ByteDance / TikTok**: [bytedance.md](bytedance.md) — Fixed capacity goroutine pools, `hmap` bucket evacuation mechanics.
- 🔷 **Meta**: [meta.md](meta.md) — Struct field alignment, memory padding rules, offset optimization.
- 📦 **Amazon / AWS**: [amazon.md](amazon.md) — Exponential backoff with full jitter retry algorithms, AWS SDK context propagation.
- 💻 **Microsoft / Azure**: [microsoft.md](microsoft.md) — Table-driven unit testing with subtests (`t.Run`), mock interfaces.
- 🪙 **Coinbase**: [coinbase.md](coinbase.md) — Limit order book matching engines, concurrent order execution.
- 🟣 **Twitch**: [twitch.md](twitch.md) — Real-time chat room broadcasters, slow-subscriber non-blocking channels.
- 🚗 **DoorDash**: [doordash.md](doordash.md) — Circuit breaker state machine pattern (`Closed` $\rightarrow$ `Open` $\rightarrow$ `Half-Open`).
