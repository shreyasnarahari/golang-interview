# Topic 05 Real Interview Q&A

### Q: (Cloud infra company) Does `new(T)` always allocate on the heap? Walk me through what actually happens.
**Frequency**: often
**Why interviewers ask**: Infra teams care about allocation discipline; this question separates candidates who repeat "new = heap" folklore from those who know the compiler's actual behavior.
**Answer**: No. `new(T)` only guarantees zeroed memory for a `T` and returns a `*T`; where that memory lives is an escape-analysis decision. If the pointer never leaves the function's frame, the allocation is folded into the stack frame and costs nothing at cleanup:
```go
func f() int {
	p := new(int) // "new(int) does not escape" — stack-allocated
	*p = 5
	return *p
}
```
It goes to the heap only when the pointer outlives the frame — returned, stored in a heap object, captured by an escaping closure, or boxed into an interface. So "use `new` to force a heap allocation" is a myth; the way to force the heap is to let the pointer escape.
**Follow-ups**: How would you demonstrate this with the compiler? Is there any case where you'd *want* to force an escape?

### Q: (High-frequency trading / latency-sensitive shop) How do you find and eliminate unwanted heap allocations in a hot path?
**Frequency**: often
**Why interviewers ask**: Latency-critical Go lives and dies by allocation rate; interviewers want a concrete, tool-driven workflow rather than vague "avoid pointers" advice.
**Answer**: First measure: `go test -bench=. -benchmem` shows allocs/op per benchmark, and `pprof` with `-sample_index=alloc_space` on a CPU/memory profile shows where they come from. Then attribute each allocation with `go build -gcflags="-m"` (add a second `-m` for the escape flow showing *why* each variable moved). Common fixes: return values instead of pointers for small structs; avoid interface boxing in hot calls (e.g. replace `fmt.Sprintf`/`Println` with typed alternatives like `slog.Int`); preallocate slices with known capacity; keep helper functions under the inliner budget so their locals stay in the caller's frame; and reuse scratch buffers with `sync.Pool` only after profiling proves reuse wins. Re-run `-benchmem` after each change — allocations you "fixed" without measurement often weren't the bottleneck.
**Follow-ups**: Why is `sync.Pool` a last resort rather than a first tool? What allocation sources can't escape analysis remove?

### Q: (Database/storage engine company) How do you inspect escape analysis decisions, and what do the cryptic `-gcflags="-m"` messages actually mean?
**Frequency**: sometimes
**Why interviewers ask**: Storage engines have tight per-record budgets; knowing whether a candidate can read compiler diagnostics is directly job-relevant.
**Answer**: Run `go build -gcflags="-m"` (works with `go test` too). Key messages: `moved to heap: x` means local `x` is heap-allocated because its address outlives the frame; `x escapes to heap` means a value (often a parameter or boxed interface argument) is allocated on the heap; `does not escape` is the good outcome; `leaking param: p` marks a parameter whose summary says it flows to the result or heap, which callers then use. `can inline f with cost 14` plus `inlining call to f` tell you inlining happened — which can retroactively un-escape a callee's locals. Add a second `-m` (`-gcflags="-m -m"`) to get the escape *flow*: a chain like `flow: {heap} ← &x: from &x (returned)` naming the exact statement responsible. For whole-module noise control, scope it: `go build -gcflags="mypkg=-m -m" ./...`.
**Follow-ups**: Why might `-m` output differ between two builds of identical source (hint: inlining and compiler version)? How would you keep a critical allocation from regressing in CI?
