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

---

## ⚠️ Common Gotchas
1. **`b.N` Misuse**: Never hardcode loop bounds inside benchmarks! Always run `for i := 0; i < b.N; i++` so Go benchmark runner can dynamically scale `b.N` until execution time stabilizes.
