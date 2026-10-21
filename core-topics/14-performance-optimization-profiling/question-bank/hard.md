# Topic 14 Hard Questions

### Q: What is false sharing in multi-core CPU cache lines and how is it prevented in Go?
**Frequency**: rare-senior
**Why interviewers ask**: Separates candidates who have debugged real multi-core scalability walls from those who only know the profiler.
**Answer**: Cores cache memory in cache lines — 64 bytes on amd64, 128 on arm64 (see `internal/cpu.CacheLinePadSize` in the runtime). If two goroutines on different cores write to *different* fields that share one line, every write invalidates the other core's copy, causing coherence traffic ("cache-line ping-pong") that can make scaling negative. Prevention: pad hot, independently-written fields to separate lines, or shard counters:
```go
type paddedCounter struct {
	v atomic.Int64
	_ [56]byte // pad so each counter owns a full 64-byte line (128 on arm64)
}
```
The Go runtime itself uses `cpu.CacheLinePad` in per-P structures. Cheaper structural fixes: per-P sharding (like `sync.Pool` and the runtime's own stat counters) or `atomic.Int64` already helps correctness but not false sharing.
**Follow-ups**: How would you detect false sharing in practice (CPU profile hints, `perf c2c`)? Why does increasing GOMAXPROCS sometimes make throughput worse?

### Q: What does an execution trace (`go tool trace`) show that a CPU profile cannot?
**Frequency**: sometimes
**Why interviewers ask**: Traces are the tool for latency/scheduler problems; many candidates stop at CPU profiles.
**Answer**: A CPU profile is a statistical aggregate of on-CPU samples; an execution trace is a timestamped event log over a window: goroutine create/start/stop/block/unblock, which P and M each goroutine ran on, network/syscall blocking, channel and lock waits, GC phases including STW pauses, and user-defined regions/tasks (`runtime/trace`). This answers questions CPU profiles cannot: why a goroutine waited 5 ms before running, scheduler latency under high GOMAXPROCS, GC assist and STW stalls. Collect via `curl 'localhost:6060/debug/pprof/trace?seconds=5' > trace.out`, `go test -trace=trace.out`, or the `runtime/trace` API, then view with `go tool trace trace.out`. Since Go 1.21 the tracer was rewritten, cutting overhead roughly an order of magnitude, so tracing in production-like environments is realistic.
**Follow-ups**: When would you choose a trace over the block profile? What trace view would you open first for a p99 latency spike?

### Q: How does profile-guided optimization (PGO) work in Go, and what is the modern workflow?
**Frequency**: rare-senior
**Why interviewers ask**: PGO is the flagship recent compiler feature; seniors should know how to deploy it safely.
**Answer**: The compiler uses a CPU pprof profile from production to make better optimization decisions: it raises the inlining budget for hot call sites and (since Go 1.22) **devirtualizes** hot interface calls into conditional direct calls. Workflow: capture a profile from a representative production workload (`/debug/pprof/profile?seconds=30`), commit it as `default.pgo` in the main package directory, and build — since Go 1.21 `-pgo=auto` is the **default**, so `go build` picks up `default.pgo` automatically (`-pgo=off` disables, `-pgo=<file>` overrides). Reported gains are typically 2–7% CPU, more on interface-heavy code with devirtualization. Practical notes: keep the profile representative and refreshed as the code changes; PGO applies transitively to dependencies of the main package; stale profiles degrade gracefully rather than miscompile.
**Follow-ups**: What happens if the committed profile no longer matches the code? Why commit the profile into the repo instead of generating it in CI?

### Q: How does Go's CPU profiler actually collect samples, and what are its blind spots?
**Frequency**: rare-senior
**Why interviewers ask**: Understanding the mechanism reveals why CPU profiles can mislead on I/O-bound or lock-bound services.
**Answer**: The runtime sets a per-thread interval timer (`setitimer(ITIMER_PROF)`) delivering `SIGPROF` at 100 Hz; the signal handler asynchronously records the interrupted goroutine's stack into a lock-free profile buffer. Because sampling only fires on threads executing Go code, the profile measures **on-CPU** time only. Blind spots: goroutines blocked on I/O, locks, or channels are invisible (use goroutine/block/mutex profiles or a trace); the signal-based mechanism can skew against short syscall-heavy functions; and flat-time aggregation hides latency distribution. Overhead is low (a few percent), which is why continuous production profiling (e.g. via `/debug/pprof/profile` scrapers) is standard practice.
**Follow-ups**: A service uses 15% CPU but has terrible p99 — why is the CPU profile unhelpful here, and what do you grab instead? Why is sampling preferred over instrumentation for CPU profiling?
