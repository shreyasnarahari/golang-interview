# Topic 12 Easy Questions

### Q: What kind of garbage collector does Go use?
**Frequency**: often
**Why interviewers ask**: It is the baseline fact every Go engineer must know before any GC tuning discussion.
**Answer**: Go uses a concurrent, non-generational, non-moving **tri-color mark-and-sweep** collector. Marking and sweeping run concurrently with the application (dedicated background workers plus mutator assists), with only two brief stop-the-world phases per cycle. "Non-moving" means heap objects are never relocated (no compaction), though goroutine *stacks* do move when they grow or shrink.
**Follow-ups**: Why does Go not use reference counting like Python or Swift? What do "generational" and "compacting" mean, and what trade-offs did Go make by skipping them?

### Q: How do you manually trigger a garbage collection cycle in code?
**Frequency**: often
**Why interviewers ask**: Checks hands-on familiarity with the `runtime` package and test/benchmark hygiene.
**Answer**: Call `runtime.GC()`. It runs a complete GC cycle and blocks until that cycle finishes. It is mainly useful in tests and benchmarks, or after releasing a large data structure when you want memory returned promptly.
```go
runtime.GC() // blocks until the full cycle completes
```
To disable GC entirely (e.g. in a controlled benchmark), use `debug.SetGCPercent(-1)` from `runtime/debug` — re-enable with `debug.SetGCPercent(100)`.
**Follow-ups**: Is `runtime.GC()` something you'd call in production hot paths? How would you force memory back to the OS instead (`debug.FreeOSMemory`)?

### Q: What is GOGC and what is its default value?
**Frequency**: often
**Why interviewers ask**: It is the single most common GC tuning knob, so interviewers use it to filter candidates who have never profiled a real service.
**Answer**: `GOGC` controls the GC's target heap growth ratio. The next collection is triggered when the heap reaches roughly `live heap × (1 + GOGC/100)`. The default is **100**, meaning "collect when the heap has doubled since the last cycle." Higher values trade more memory for less GC CPU; lower values do the opposite. `GOGC=off` disables ratio-based collection entirely. It can be set via environment variable or `debug.SetGCPercent`.
**Follow-ups**: What happens to GC frequency as live heap grows with GOGC fixed? How does GOGC interact with GOMEMLIMIT?

### Q: What is GOMEMLIMIT (Go 1.19+)?
**Frequency**: sometimes
**Why interviewers ask**: Containerized Go services OOM-killing under load is a common production incident; this separates book knowledge from operational experience.
**Answer**: `GOMEMLIMIT` is a **soft** memory limit for the Go runtime, introduced in Go 1.19 (default: effectively unlimited). When total runtime memory approaches the limit, the GC runs more aggressively to stay under it, overriding the GOGC ratio. "Soft" means the runtime does not guarantee staying under it — a single huge allocation or rapid transient growth can still exceed it. Typical container pattern: `GOMEMLIMIT` slightly below the cgroup limit, often with `GOGC=off` so GC only fires against the limit.
**Follow-ups**: Why did GOMEMLIMIT replace the "memory ballast" trick? Does the limit cover goroutine stacks and cgo allocations? (Stacks/globals count toward the runtime total; cgo memory does not.)
