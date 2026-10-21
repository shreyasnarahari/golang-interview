# Topic 14 Medium Questions

### Q: Why should `b.ResetTimer()` be called in Go benchmarks?
**Frequency**: often
**Why interviewers ask**: Benchmark hygiene mistakes produce misleading numbers, and interviewers want to know you can produce trustworthy measurements.
**Answer**: To discard setup time (reading fixtures, building mock datasets) so measured time reflects only the code under test. Related: `b.StopTimer()`/`b.StartTimer()` to exclude per-iteration setup, and `b.ReportAllocs()` (or `-benchmem`) to include B/op and allocs/op. Also note the `b.N` loop should not hoist work the compiler can eliminate — assign results to package-level sinks to defeat dead-code elimination:
```go
var result int

func BenchmarkWork(b *testing.B) {
	b.ResetTimer()
	var r int
	for i := 0; i < b.N; i++ {
		r = Work()
	}
	result = r
}
```
**Follow-ups**: What happens if you forget `ResetTimer` after expensive setup? How can the compiler invalidate a microbenchmark?

### Q: How does `sync.Pool` work, and why can an object you `Put` disappear?
**Frequency**: often
**Why interviewers ask**: Tests whether you understand Pool's semantics rather than treating it as a general cache.
**Answer**: `sync.Pool` keeps per-P (per-logical-processor) free lists of reusable objects, avoiding allocations and lock contention on the hot path. Items have **no lifetime guarantee**: on every GC cycle, pools are drained (items survive at most one GC via a victim cache, then are dropped), so `Get` may return nothing and `New` is called. Consequences: never use Pool for scarce resources with invariants (DB connections with open sockets); always reset returned objects; Pool shines for high-churn temporary objects like `bytes.Buffer`:
```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)
```
**Follow-ups**: Why is Pool per-P rather than a single mutex-protected slice? When does Pool hurt performance instead of helping?

### Q: How do you rigorously compare benchmark results before and after an optimization?
**Frequency**: sometimes
**Why interviewers ask**: A single `go test -bench` run is noisy; seniors expect statistical comparison.
**Answer**: Run multiple iterations with `-count` and compare distributions with `benchstat` (`golang.org/x/perf`):
```text
go test -bench=. -benchmem -count=10 > old.txt
# make the change
go test -bench=. -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```
`benchstat` reports the delta with a confidence interval and flags statistically insignificant results ("~ no change"). Also pin the environment: same machine, disable turbo/CPU scaling if possible, avoid background load, and check both ns/op and allocs/op.
**Follow-ups**: Why is `-count=1` insufficient? What can make benchmark results differ between your laptop and CI?

### Q: How do you enable mutex and block profiling, and what do they tell you?
**Frequency**: sometimes
**Why interviewers ask**: Latency problems in concurrent services are usually contention, not CPU — these profiles are how you find it.
**Answer**: They are off by default and must be enabled:
```go
runtime.SetMutexProfileFraction(rate) // rate>0: ~1/rate of contention events sampled
runtime.SetBlockProfileRate(rate)     // rate in ns: blocking events longer than rate sampled (1 = all)
```
Then fetch `/debug/pprof/mutex` and `/debug/pprof/block`. The **mutex profile** shows where goroutines contend on locks (`sync.Mutex.Unlock` stacks with accumulated delay); the **block profile** shows time spent waiting on channel ops, `select`, `time.Sleep`, etc. Typical use: high tail latency with low CPU utilization → check goroutine dump, then mutex/block profiles for the contended lock.
**Follow-ups**: Why are these sampled rather than recorded unconditionally? How do you distinguish lock contention from channel blocking in the profiles?
