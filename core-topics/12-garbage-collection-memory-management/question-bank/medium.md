# Topic 12 Medium Questions

### Q: What are the STW (stop-the-world) phases in Go's concurrent GC?
**Frequency**: often
**Why interviewers ask**: Candidates who claim "Go GC has no pauses" reveal they have never measured tail latency.
**Answer**: Each GC cycle has **two** brief STW phases:
1. **Sweep termination / mark setup**: finishes the previous sweep, enables the write barrier, and marks all goroutine stacks as needing scan.
2. **Mark termination**: re-scans anything still dirty, disables the write barrier, and computes the next heap goal.

Each pause is typically tens to a few hundred microseconds (the runtime targets sub-millisecond), dominated by the number of goroutines and Ps, not heap size. Everything between them — the bulk of marking and all sweeping — runs concurrently.
**Follow-ups**: What can make these pauses long in practice? (Huge goroutine counts, goroutines stuck in tight loops without preemption points delaying the stop.) How would you observe them (`runtime.MemStats.PauseNs`, `GODEBUG=gctrace=1`, execution traces)?

### Q: What happens when application goroutines allocate faster than the background markers can keep up?
**Frequency**: sometimes
**Why interviewers ask**: It tests understanding of how "concurrent GC" stays safe under adversarial allocation rates — a real cause of GC CPU spikes.
**Answer**: The **mutator assist** system kicks in. A goroutine that allocates during an active mark phase must first perform a proportional amount of marking work (an "assist debt") before its allocation proceeds. The faster it allocates, the more marking it is drafted into doing. This back-pressure guarantees marking finishes before the heap blows through its goal, but it means GC CPU cost lands on your application goroutines, inflating tail latency for allocation-heavy request handlers.
**Follow-ups**: How do mutator assists show up in an execution trace? What code changes reduce assist load (fewer/smaller allocations, `sync.Pool`, raising GOGC)?

### Q: How would you configure GC for a memory-constrained container running an allocation-heavy service?
**Frequency**: often
**Why interviewers ask**: This is the practical k8s deployment question — it tests whether you can combine GOGC, GOMEMLIMIT, and profiling into a strategy.
**Answer**: Set `GOMEMLIMIT` a bit *below* the container's cgroup limit (leaving headroom for non-heap memory: stacks, OS overhead, cgo) so the runtime self-throttles before the OOM killer does. A common pattern is `GOGC=off` plus `GOMEMLIMIT`, so collection is driven purely by the memory limit — the runtime also caps GC CPU usage (~50%) to avoid thrashing when pinned at the limit. Then attack the root cause: reduce allocation rate via escape analysis, buffer reuse, and `sync.Pool`, since no GC setting fixes a leaky allocation profile. Verify with `GODEBUG=gctrace=1` and heap profiles.
**Follow-ups**: Why is GOMEMLIMIT a *soft* limit, and what can still get you OOM-killed? How do you detect that GC is thrashing at the limit?

### Q: `runtime.SetFinalizer` vs `runtime.AddCleanup` (Go 1.24) — what's the difference and which should you use?
**Frequency**: sometimes
**Why interviewers ask**: Modern Go replaced a famously error-prone API; interviewers want to know if you track the current runtime.
**Answer**: Prefer `runtime.AddCleanup` (Go 1.24). `SetFinalizer` attaches the cleanup to the object itself, which has classic pitfalls: the finalizer must not resurrect the object's fields into new cycles, it blocks collection of cyclic structures containing the object (finalizers prevent cycle collection), and it constrains the object layout. `AddCleanup` takes a separate `arg`, decoupling the cleanup from the object's lifetime graph:
```go
runtime.AddCleanup(f, func(fd int) { syscall.Close(fd) }, f.fd)
```
It allows collection of cycles, supports multiple cleanups per object, and permits interior-pointer attachment. Go 1.24 also added the `weak` package for weak pointers, replacing another common finalizer misuse (caches).
**Follow-ups**: Why are finalizers/cleanups unsuitable for prompt resource release (timing is non-deterministic — use explicit `Close`)? What does the `weak` package enable that finalizers couldn't?
