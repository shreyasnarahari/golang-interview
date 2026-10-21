# Topic 14: Performance Optimization & Profiling (`pprof`)

## 📌 Overview
Go ships with world-class profiling instrumentation built directly into the standard library (`net/http/pprof` and `runtime/pprof`).

---

## 🎯 Key Concepts

### 1. `pprof` Profiling Types
- **CPU Profile**: Samples stack traces every 10ms to identify CPU hot paths.
- **Heap Profile**: Samples active memory allocations to detect memory leaks and high allocation rates.
- **Goroutine Profile**: Captures stack traces of all live goroutines (essential for identifying goroutine leaks and deadlocks).
- **Block Profile**: Measures time goroutines spend waiting on unbuffered channels and mutexes.
- **Mutex Profile**: Measures lock contention on `sync.Mutex` and `sync.RWMutex`.

```bash
# Analyze CPU profile interactively via web UI
go tool pprof -http=:8080 cpu.pprof

# Analyze Memory allocations
go tool pprof -alloc_space heap.pprof
```

### 2. Benchmarking in Go (`testing.B`)
- Write benchmarks using `BenchmarkXxx(b *testing.B)`.
- Use `b.ReportAllocs()` to track byte and number of heap allocations per iteration.
- Always reset timer `b.ResetTimer()` after setup operations!

```go
func BenchmarkProcess(b *testing.B) {
    data := prepareData()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}
```

### 3. Allocation Recycling with `sync.Pool`
Thread-safe cache managed by Go runtime GC. Reduces allocation churn in high-throughput servers.

### 4. Live Profiling Workflow (`net/http/pprof`)

**Step 1 — Instrument the server.** Importing `net/http/pprof` for side effects registers profiling handlers on `http.DefaultServeMux`:

```go
import (
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
)
```

Endpoints exposed under `/debug/pprof/`: `profile` (CPU), `heap`, `allocs`, `goroutine`, `block`, `mutex`, `trace`, `threadcreate`. With a custom `http.ServeMux`, the blank import has no effect — register the handlers from `net/http/pprof` explicitly (or serve them on a separate admin listener).

**Step 2 — Capture a profile under load.** `go tool pprof` fetches profiles directly over HTTP:

```bash
# 30s CPU profile (server must be handling representative traffic meanwhile)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Heap profile (live/in-use allocations by default)
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Step 3 — Read the output.** Inside the interactive pprof shell:
- `top` — functions sorted by **flat** samples (time spent in the function itself); `top -cum` sorts by cumulative (function + everything it calls).
- `list <func>` — source annotated with per-line sample counts; pinpoints the exact hot line.
- `peek <regex>` — the callers above and callees below matching functions, with flat/cum breakdown.
- `web` — call-graph view (needs Graphviz); `-http=:8081` serves the same views (flame graph included) in a browser.

For heap profiles, compare **`-inuse_space`** (default: memory still live — leak hunting) against **`-alloc_space`** (everything ever allocated — churn hunting).

**Step 4 — Execution tracer, when sampling isn't enough.** `runtime/trace` records an event timeline: goroutine state transitions, blocking on channels/locks/syscalls, network waits, GC phases.

```bash
curl -o trace.out "http://localhost:8080/debug/pprof/trace?seconds=5"
go tool trace trace.out
```

**pprof vs trace**: pprof answers "where does aggregate time/memory go" — statistical sampling, cheap enough to run continuously in production. The execution tracer answers "what happened, in what order, and why did this goroutine stall" — use it for latency, scheduling, and serialization questions that pprof's averages hide.

### 5. Profile-Guided Optimization (PGO)
Introduced in Go 1.20 and **enabled by default since Go 1.21**: the compiler automatically picks up a `default.pgo` file placed in the main package directory, or you pass a profile explicitly:

```bash
# Capture a pprof CPU profile from production/staging traffic, then either:
cp cpu.pprof ./cmd/server/default.pgo
go build ./cmd/server          # auto-detected default.pgo

# ... or point at it explicitly:
go build -pgo=cpu.pprof ./cmd/server

# Opt out:
go build -pgo=off ./cmd/server
```

What the profile guides:
- **Hot-path inlining**: functions the profile marks as hot are inlined even when they exceed the default cost budget, cutting call overhead where it actually matters.
- **Devirtualization**: hot interface calls are rewritten into a type check against the dominant concrete type plus a direct call, which can then be inlined further.

The Go team measures roughly 2–7% CPU improvement across representative workloads. The profile must reflect real traffic — a stale or unrepresentative profile simply yields little gain (it never miscompiles).

### 6. Lockless Concurrency: `sync/atomic` vs `sync.Mutex`

| Aspect | `sync.Mutex` | `sync/atomic` / `atomic.Pointer[T]` |
|---|---|---|
| **Mechanism** | Uncontended fast path is a single CAS; on contention the goroutine blocks on a runtime semaphore — the Go scheduler deschedules it (`gopark`) and wakes it when the lock is released (`goready`) | Hardware atomic instructions: CAS / bus-locked loads and stores |
| **Contention Overhead** | Goroutine park + wake round trip through the Go scheduler (user-space scheduling, not an OS-thread context switch) | Hardware instruction latency (~5–10 ns) |
| **Best For** | Multi-field invariants and multi-step critical sections | Single counters, flags, atomic pointer swaps |

Rule of thumb: if the critical section is "update one word" (a counter, a flag, a pointer swap), use atomics. If several fields must stay consistent with each other, use a mutex — a pile of atomics holding a cross-field invariant is a subtle bug factory.

---

## ⚠️ Common Gotchas
1. **`b.N` Misuse**: Never hardcode loop bounds inside benchmarks! Always run `for i := 0; i < b.N; i++` so Go benchmark runner can dynamically scale `b.N` until execution time stabilizes.
2. **Exposing `/debug/pprof/` Publicly**: The endpoints leak goroutine stacks and heap contents and enable CPU-profile DoS. Bind them to an internal/admin listener, never a public port.
3. **Profiling an Idle Server**: CPU profiles are statistical samples of whatever is running — under no load they show GC and background noise. Generate representative traffic before capturing (and before training a PGO profile).
