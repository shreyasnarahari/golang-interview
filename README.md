# Go (Golang) Interview Prep 🐹

A structured, 80/20 study system for Go software engineering, backend, distributed systems, cloud-native (Kubernetes/Docker), and infrastructure developer interviews.

---

## 🎯 What's Here

```
golang-interview/
├── README.md                      ← Master guide, 80/20 roadmap & learning path
├── syllabus.txt                   ← Complete 20-topic Go reference map (Basics to Runtime & GC)
├── high-performance-go.md         ← Specialized guide (GMP scheduler, Tri-color GC, pprof, zero-copy, sync.Pool)
├── real-interview-q/              ← Curated questions actually asked at top tech companies
│   └── interview-questions.md
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

---

## 🚀 How to Use This Repository

### 1. 📖 The 80/20 Study Plan
Don't try to memorize everything at once. Focus on high-impact areas first:
- **Phase 1 (Week 1 — Core Language Fundamentals)**: Topics 01 – 05 (Data Types, Slices/Maps Internals, Defer/Panic, Structs & Interfaces, Pointers & Escape Analysis).
- **Phase 2 (Week 2 — Systems & Concurrency)**: Topics 06 – 10 (Goroutines & Channels, `sync` Package & Patterns, Context Cancellation, Errors, Generics).
- **Phase 3 (Week 3 — Runtime Internals & Performance)**: Topics 11 – 15 (GMP Scheduler Model, Tri-Color GC, Reflection & `unsafe`, `pprof` Profiling, Testing/Tooling).

### 2. 💻 Hands-On Interactive Practice
Every module inside `core-topics/` contains a `practice/` directory with starter code and matching `go test` unit test suites (`p0*_test.go`).

Run tests across all topics using `go test`:
```bash
go test ./...
```
Or run tests for a specific topic:
```bash
cd core-topics/04-structs-interfaces-methods/practice
go test -v
```

### 3. 🏢 Real-World Company Focus (`real-interview-q/`)
Check `real-interview-q/interview-questions.md` for company-tagged questions (**Google**, **Uber**, **Meta**, **Amazon**, **Docker/Kubernetes firms**, **HashiCorp**, **CockroachDB**, **Cloudflare**).

### 4. ⚡ High Performance & Runtime Internals (`high-performance-go.md`)
For senior roles, distributed systems engineering, or low-latency infrastructure interviews, study `high-performance-go.md` for Go scheduler (GMP) mechanics, Tri-color mark-sweep GC with write barriers, memory allocation (`mcache`/`mcentral`/`mheap`), zero-copy string/slice conversions (`unsafe.StringData`), and allocation elimination via `sync.Pool`.

---

## 🎓 Difficulty Progression

- 🟢 **Beginner**: Variables, zero values, slice headers, map usage, structs, interfaces, value vs pointer receivers, basic error handling.
- 🟡 **Intermediate**: `defer` stack semantics, `panic`/`recover`, channels (buffered vs unbuffered), `select` statement, `sync.WaitGroup`/`Mutex`/`RWMutex`, `context.Context` propagation, error wrapping (`%w`).
- 🔴 **Advanced**: Go Runtime Scheduler (GMP model, work stealing, async preemption), Tri-color concurrent GC & write barriers, `GOMEMLIMIT` / `GOGC` tuning, Escape Analysis (`-gcflags="-m"`), `reflect`, `unsafe.Pointer`, `pprof` memory/CPU analysis.
