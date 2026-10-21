# Topic 07 Real Interview Q&A

### Q: (At a high-frequency trading firm) What happens if you pass a `sync.Mutex` by value to a function, and how would you catch it?
**Frequency**: often
**Why interviewers ask**: Latency-critical systems live and die by lock correctness, and this bug compiles cleanly.
**Answer**: Go copies the mutex struct — including its internal state (the locked flag and waiter count). The function now locks and unlocks an independent copy, so mutual exclusion is silently broken: two goroutines can be "inside the lock" simultaneously, producing data races that corrupt shared state. The same applies to copying any struct that embeds a `Mutex`, `RWMutex`, `WaitGroup`, or `Cond` after first use. Detection: `go vet` flags passing lock values by copy ("passes lock by value"), the `go test -race` race detector surfaces the resulting races at runtime, and the standard library embeds a `noCopy` sentinel in these types so linters can catch it. Fix: always pass pointers, and make the mutex an unexported field of the struct it protects.
**Follow-ups**: Why is copying a mutex that has never been used still risky? Where in the standard library is intentional mutex copying allowed (hint: before first use)?

### Q: (At a cloud infrastructure company) When would you use `sync.Map` over `sync.Mutex` + a plain map on a hot metadata path?
**Frequency**: often
**Why interviewers ask**: Picking the wrong one on a hot path shows up as p99 latency regressions; they want measured reasoning, not folklore.
**Answer**: `sync.Map` is specialized, not a drop-in replacement. It wins in two cases: **(1) write-once/read-many** — a key is stored once and read many times (e.g., a growing cache of service descriptors), because reads are lock-free; and **(2) disjoint key sets** — many goroutines each updating their own keys, avoiding contention on one big lock. It loses for general read/write workloads: keys/values are boxed into `any` (allocations, lost type safety, `interface{}` casts at call sites), Range is not a snapshot, and churning keys defeat its read-path optimizations. For most services, `RWMutex + map[K]V` (typed, since generics) is faster, type-safe, and easier to reason about; for heavy writer workloads, sharded maps (N striped locks) beat both. Since Go 1.24 `sync.Map` is backed by a lock-free hash trie, which improved write-heavy cases, but the API-level costs remain: erased types, and still no cheap way to count entries (no `Len` — you must `Range`). Answer with a benchmark on the actual workload shape.
**Follow-ups**: How would you shard a map to reduce lock contention, and what hash would you shard on? What does `sync.Map.Range` guarantee about consistency?

### Q: (At a data infrastructure company) Design a worker pool that streams results back and shuts down cleanly on cancellation or the first error. Walk me through the coordination.
**Frequency**: often
**Why interviewers ask**: This is a live-coding staple that tests channels, WaitGroup, context, and goroutine-leak awareness in one design.
**Answer**: Core shape: one `jobs` channel, N workers fanning out from it, one `results` channel fanning back in, and a dedicated closer.
```go
jobs := make(chan Job)
results := make(chan Result)

var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs {
            select {
            case results <- process(j):
            case <-ctx.Done():
                return
            }
        }
    }()
}
go func() { wg.Wait(); close(results) }()
```
Key decisions to narrate:
1. **Closing `jobs`**: the producer closes it after enqueueing (or on `ctx.Done()`), and workers drain with `range` — no "poison pill" values needed.
2. **Closing `results`**: never closed by workers; a separate goroutine does `wg.Wait(); close(results)` so the consumer's `range results` terminates.
3. **Cancellation**: `ctx` is checked on *both* send and receive sides so a consumer that stops reading doesn't strand workers (the classic leak); use `context.WithCancelCause` (Go 1.20) so the first error can be stored as the cancel cause via `cancel(err)`.
4. **First-error semantics**: on error, record it once (`sync.Once` or `errgroup` from `golang.org/x/sync`), cancel the context so all workers and the producer stop, and let `results` still be drained/closed so nothing leaks.
5. **Backpressure**: channel buffer sizes and worker count are the knobs; an unbuffered `results` channel naturally throttles workers to consumer speed.
Trade-offs worth mentioning: errgroup does exactly this but hides the mechanics — in an interview, write it by hand first, then say you'd ship errgroup.
**Follow-ups**: How do you propagate the *first* error while letting in-flight work finish? How would you add per-job timeouts without spawning a timer per job (hint: `time.AfterFunc`, Go 1.21+)? 
