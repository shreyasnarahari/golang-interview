# Topic 14 Easy Questions

### Q: What command launches the interactive web GUI for pprof visualization?
**Frequency**: often
**Why interviewers ask**: Confirms hands-on familiarity with the standard Go profiling workflow, not just theory.
**Answer**: `go tool pprof -http=:8080 profile.pprof` opens a browser UI with a flame graph, graph view, top functions, and annotated source. It can also attach directly to a live server: `go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30`.
**Follow-ups**: What are the most useful views in the web UI (flame graph vs. graph vs. peek)? How do you profile a remote production server with it?

### Q: What are the main profile types available in pprof, and what does each one show?
**Frequency**: often
**Why interviewers ask**: Knowing which profile answers which question is the foundation of Go performance debugging.
**Answer**: The main types are: **CPU (profile)** — where on-CPU time is spent; **heap** — live in-use memory by allocation stack; **allocs** — cumulative allocations since start (finds churn/GC pressure); **goroutine** — stack traces of all live goroutines (leaks, blocked goroutines); **block** — time blocked on synchronization primitives (channels, mutexes via `select`/receive); **mutex** — lock contention; **threadcreate** — OS thread creation stacks. All are exposed under `/debug/pprof/` when `net/http/pprof` is imported.
**Follow-ups**: Which profile would you grab first for a memory leak vs. a latency spike? What is the difference between `heap` and `allocs`?

### Q: How do you collect a CPU profile from a Go program or test?
**Frequency**: often
**Why interviewers ask**: Collecting the profile is the first concrete step of any optimization task.
**Answer**: Three common ways: (1) `go test -bench=. -cpuprofile=cpu.out` for benchmarks; (2) hit `/debug/pprof/profile?seconds=30` on a service importing `net/http/pprof`; (3) programmatically with `runtime/pprof`:
```go
f, _ := os.Create("cpu.out")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```
Then analyze with `go tool pprof cpu.out`.
**Follow-ups**: How long should a CPU profile run to be meaningful? How would you profile a short-lived CLI tool where an HTTP endpoint isn't practical?

### Q: In the `go tool pprof` interactive shell, what do `top`, `list`, and `peek` show?
**Frequency**: sometimes
**Why interviewers ask**: Interviewers want to see you can navigate a profile without the web UI, e.g. on a production box.
**Answer**: `top` lists the functions with the highest **flat** (self) sample count; `top -cum` sorts by **cumulative** time including callees. `list FuncName` shows the annotated source of a function with per-line sample counts. `peek regex` lists callers and callees of matching functions, useful for understanding who is responsible for a hot callee. Use `-nodecount`, `-edgefraction`, and flags like `-sample_index=alloc_space` to refine.
**Follow-ups**: What is the difference between flat and cumulative time? If `top` shows `runtime.mallocgc` on top, what does that tell you?
