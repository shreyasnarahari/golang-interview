# Topic 14 Real Interview Q&A

### Q: (Asked at a cloud infrastructure company) How do you enable live HTTP pprof profiling endpoints in a Go web service, and what precautions do you take before shipping it?
**Frequency**: often
**Why interviewers ask**: Production profiling is standard at infra companies, but naively exposing it is a security and reliability footgun.
**Answer**: Import for side effects, which registers handlers under `/debug/pprof/` on `http.DefaultServeMux`:
```go
import _ "net/http/pprof"
```
Caveats: (1) it binds to the **default** mux — if your server uses a custom mux, register the handlers explicitly; (2) never expose `/debug/pprof/` on the public internet — goroutine dumps leak stack traces, source paths, and parameter values, and `/debug/pprof/cmdline` reveals flags; serve it on an internal-only port or behind auth; (3) a CPU profile (`/debug/pprof/profile?seconds=30`) adds load, so keep durations bounded on hot instances. Available endpoints include `heap`, `goroutine`, `allocs`, `block`, `mutex`, `threadcreate`, `trace`, plus `symbol` and `cmdline`.
**Follow-ups**: How would you profile a fleet of hundreds of instances — per-instance scraping or continuous profiling? How do you enable mutex/block profiling on these endpoints?

### Q: (Asked at a gaming company) What flag displays memory allocation bytes in `go test -bench`, and why do allocations matter as much as ns/op?
**Frequency**: often
**Why interviewers ask**: Game and low-latency code lives and dies by GC pressure; they want candidates who read all benchmark columns.
**Answer**: `-benchmem` (or `b.ReportAllocs()` inside the benchmark) adds two columns: **B/op** (bytes allocated per operation) and **allocs/op** (number of heap allocations per operation). Allocations drive GC frequency, and GC costs CPU (mark/assist) and causes STW pauses, so a function with better ns/op but higher allocs/op can be worse at scale. Example: `go test -bench=. -benchmem -count=5`. To attribute allocations, capture `-memprofile=mem.out` and inspect `go tool pprof -sample_index=alloc_space mem.out`. Common fixes: `sync.Pool` for temporary buffers, preallocating slices/maps with capacity hints, and avoiding interface boxing that escapes to the heap.
**Follow-ups**: A benchmark shows 0 B/op for one implementation — how is that possible, and how do escape analysis decisions factor in? How do you compare two benchmark runs statistically?

### Q: (Asked at a fintech company) Your Go payment service's p99 latency doubled after a deploy while CPU usage stayed flat. Walk me through your diagnosis.
**Frequency**: often
**Why interviewers ask**: This open-ended scenario tests whether you can sequence the profiling toolbox against a real symptom, not just name tools.
**Answer**: Flat CPU + worse p99 means goroutines are **waiting**, not computing — a CPU profile will show nothing, so start with off-CPU evidence. Steps: (1) grab a **goroutine profile** (`/debug/pprof/goroutine?debug=1` or `debug=2` for full stacks) to see where goroutines pile up — look for thousands blocked on the same channel, lock, or DB call; (2) enable and fetch **mutex** (`runtime.SetMutexProfileFraction`) and **block** (`runtime.SetBlockProfileRate`) profiles to quantify contention introduced by the deploy; (3) capture a 5-second **execution trace** (`/debug/pprof/trace?seconds=5`) to inspect scheduler latency, GC STW events, and per-goroutine blocking timelines; (4) diff a **heap/allocs** profile against the pre-deploy baseline in case new allocations increased GC pressure; (5) correlate with deploy diff — typical culprits are a new mutex around a hot path, a shared resource with reduced concurrency (smaller connection pool), or an added synchronous downstream call. Fix, then verify with before/after benchmarks using `-count=10` + `benchstat`, and consider committing a refreshed `default.pgo` so PGO keeps matching production behavior.
**Follow-ups**: Why not start with the CPU profile here? How would you set up continuous profiling so next time you have pre-incident baselines automatically?
