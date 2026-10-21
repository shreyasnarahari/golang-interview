# Go (Golang) Interview Prep 🐹

A structured, 80/20 study system for Go software engineering, backend, distributed systems, cloud-native (Kubernetes/Docker), and infrastructure developer interviews.

---

## 🎯 What's Here

```
golang-interview/
├── README.md                      ← Master guide, 80/20 roadmap & learning path
├── syllabus.txt                   ← Complete 20-topic Go reference map (Basics to Runtime & GC)
├── REVIEW.md                      ← Critical review & prioritized rebuild roadmap
├── Makefile                       ← build / test / race / vet / fmt / exercise / mdcheck / ci targets
├── .github/workflows/ci.yml       ← CI: fmt, vet, build, race tests, markdown snippet compile check
├── tools/mdcheck/                 ← Compile-checks complete Go programs embedded in markdown
├── real-interview-q/              ← Curated company-tagged question files (index: real-interview-q/README.md)
│   └── system-design/             ← Go-centric system-design walkthroughs (rate limiter, job scheduler, cache service)
├── capstones/                     ← Level 4 Applied: portfolio-grade projects (spec + stubs + acceptance tests + reference solution)
│   ├── rate-limiter/              ← Token bucket + sliding window, thread-safe, with benchmarks
│   ├── kv-store/                  ← In-memory KV store: TTL heap, gob snapshots, sweeper
│   ├── job-queue/                 ← Worker pool: backpressure, graceful shutdown, panic recovery
│   └── mini-redis/                ← TCP server: RESP-ish protocol, goroutine-per-conn, pprof-instrumented
├── practice-go/                   ← Sandbox environment configured with Go module
│   ├── go.mod
│   └── main.go
└── core-topics/                   ← Deep-dive study modules (Beginner to Advanced)
    ├── 01-golang-fundamentals-data-types/
    ├── 02-arrays-slices-maps-internals/
    ├── 03-functions-defer-panic-recover/
    ├── 04-structs-interfaces-methods/
    ├── 05-pointers-memory-allocation-escape-analysis/
    ├── 06-goroutines-channels-concurrency/
    ├── 07-concurrency-patterns-sync-package/
    ├── 08-context-cancellation-timeouts/
    ├── 09-error-handling-custom-errors/
    ├── 10-generics-type-parameters/
    ├── 11-go-runtime-scheduler-gmp/
    ├── 12-garbage-collection-memory-management/
    ├── 13-reflection-unsafe-cgo/
    ├── 14-performance-optimization-profiling/
    └── 15-testing-tooling-build-systems/
```

Each `core-topics/NN-*/` module contains:
- `README.md` — the study notes
- `exercises/` — **stub implementations** (`panic("TODO: implement ...")`) plus the full test suite; red until *you* implement them
- `solutions/` — reference implementations with the same test suite; always green
- `question-bank/` — 15 questions per module (easy / medium / hard / real-interview), each with frequency tag, why-it's-asked, model answer, and follow-ups; exportable to Anki via `make anki`
- `resources.md` — curated links

Modules 05, 11, 12, and 14 additionally contain a `lab/` — guided hands-on experiments (escape-analysis scavenger hunt, `GODEBUG=schedtrace`, GC tuning under `GOGC`/`GOMEMLIMIT`, live `pprof` profiling of a deliberately slow server) with exact commands and expected observations.

---

## 🚀 How to Use This Repository

### 1. 📖 The 80/20 Study Plan
Don't try to memorize everything at once. Focus on high-impact areas first:
- **Phase 1 (Week 1 — Core Language Fundamentals)**: Topics 01 – 05 (Data Types, Slices/Maps Internals, Defer/Panic, Structs & Interfaces, Pointers & Escape Analysis).
- **Phase 2 (Week 2 — Systems & Concurrency)**: Topics 06 – 10 (Goroutines & Channels, `sync` Package & Patterns, Context Cancellation, Errors, Generics).
- **Phase 3 (Week 3 — Runtime Internals & Performance)**: Topics 11 – 15 (GMP Scheduler Model, Tri-Color GC, Reflection & `unsafe`, `pprof` Profiling, Testing/Tooling).
- **Phase 4 (Week 4 — Applied Synthesis)**: The `capstones/` projects + `real-interview-q/system-design/` walkthroughs.

### 2. 💻 Hands-On Exercise-First Practice
Every module inside `core-topics/` contains an `exercises/` directory with **stub implementations** (`panic("TODO: implement ...")`) and a full test suite. The tests fail until *you* write the implementation — that is the point. Reference implementations live in `solutions/`; peek only after your own version passes.

Run a module's exercise suite:
```bash
make exercise T=06
# or directly:
go test -v ./core-topics/06-goroutines-channels-concurrency/exercises/
```

Verify the reference solutions (always green):
```bash
make test        # all modules' solutions
make race        # solutions under the race detector
```

Full local CI (gofmt + vet + build + solution tests + markdown snippet compile check):
```bash
make ci
```

### 3. 🏢 Real-World Company Focus (`real-interview-q/`)
Check `real-interview-q/README.md` for the index of company-tagged question files (**Google**, **Uber**, **Meta**, **Amazon**, **Docker/Kubernetes firms**, **HashiCorp**, **CockroachDB**, **Cloudflare**).

### 4. ⚡ High Performance & Runtime Internals (Modules 11–14 + Labs)
For senior roles, distributed systems engineering, or low-latency infrastructure interviews, the deep runtime content lives directly in the modules (the former `high-performance-go.md` was merged into them):
- **Module 11** — GMP scheduler: work-stealing order, `runnext`, syscall handoff, `sysmon`, SIGURG async preemption, plus a `lab/` with `GODEBUG=schedtrace` and `GOMAXPROCS` scaling experiments.
- **Module 12** — Tri-color GC with the hybrid write barrier, the TCMalloc `mcache`/`mcentral`/`mheap` allocator, `GOGC`/`GOMEMLIMIT` tuning, plus a `lab/` allocation-churn experiment under `GODEBUG=gctrace=1`.
- **Module 14** — Full `pprof` workflow (`net/http/pprof`, `go tool pprof top/peek/list`), execution tracer, PGO, `sync/atomic` vs `sync.Mutex`, plus a `lab/` that profiles a deliberately slow server.
- **Module 05** — Complete escape-analysis rules with a predict-then-verify scavenger-hunt `lab/`.
- **Module 13** — Zero-copy string/byte conversions via `unsafe.StringData`/`unsafe.Slice`.

### 5. 🏗️ Capstone Projects (`capstones/` — Level 4: Applied)
After the 15 modules, build four portfolio-grade systems that force concurrency + errors + context + performance together. Each capstone has a spec `README.md` (requirements, API contract, trade-offs, "what the interviewer is listening for"), a stubbed package with failing acceptance tests, and a reference `solution/`:

1. **`rate-limiter/`** — thread-safe per-key token bucket + sliding window, lazy refill, idle-key eviction, 0-alloc hot path benchmarks.
2. **`kv-store/`** — in-memory KV with TTL (`container/heap` deadlines), gob snapshot/restore, sweeper goroutine.
3. **`job-queue/`** — worker pool with bounded-queue backpressure, graceful drain on `Shutdown(ctx)`, per-job panic recovery, leak-free.
4. **`mini-redis/`** — TCP server speaking a RESP-inspired protocol, goroutine-per-connection, context-driven shutdown, `net/http/pprof` observability.

Run a capstone's acceptance suite (red until you implement the stubs):
```bash
make capstone C=rate-limiter
```

Pair each capstone with its walkthrough in `real-interview-q/system-design/` to practice scaling the single-node design into a distributed one.

---

## 🎓 Difficulty Progression

- 🟢 **Beginner**: Variables, zero values, slice headers, map usage, structs, interfaces, value vs pointer receivers, basic error handling.
- 🟡 **Intermediate**: `defer` stack semantics, `panic`/`recover`, channels (buffered vs unbuffered), `select` statement, `sync.WaitGroup`/`Mutex`/`RWMutex`, `context.Context` propagation, error wrapping (`%w`).
- 🔴 **Advanced**: Go Runtime Scheduler (GMP model, work stealing, async preemption), Tri-color concurrent GC & write barriers, `GOMEMLIMIT` / `GOGC` tuning, Escape Analysis (`-gcflags="-m"`), `reflect`, `unsafe.Pointer`, `pprof` memory/CPU analysis.
- 🟣 **Expert / Applied**: Build the four `capstones/` projects unaided and defend the design trade-offs (lock granularity, backpressure, allocation strategy, graceful shutdown); walk through the `real-interview-q/system-design/` designs.
