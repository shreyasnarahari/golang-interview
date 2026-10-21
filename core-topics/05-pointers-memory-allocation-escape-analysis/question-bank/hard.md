# Topic 05 Hard Questions

### Q: How do goroutine stacks grow in Go, and how does that interact with escape analysis?
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates who know the runtime from those who only know the language surface.
**Answer**: Every goroutine starts with a small stack (2 KiB today) that grows dynamically by allocating a larger segment and copying the old frames over, up to a default maximum of 1 GB (64-bit). This copying is why Go can run millions of goroutines, and it is also why the compiler must know every pointer into a stack frame precisely (pointer maps in the stack maps) — a pointer that cannot be tracked safely is forced to the heap. Escape analysis is the compile-time side of the same story: objects too large for a stack slot (over 128 KiB) are moved to the heap outright, reported as `too large for stack` by `-gcflags="-m -m"`.
**Follow-ups**: Why can't Go move a stack if an untracked pointer (e.g. from `unsafe.Pointer` misuse) points into it? What happens when a goroutine exceeds the 1 GB stack limit?

### Q: How does inlining interact with escape analysis? Give a concrete example.
**Frequency**: rare-senior
**Why interviewers ask**: It tests whether you understand that escape decisions are not fixed per function but depend on the call site after optimization.
**Answer**: Escape analysis runs after inlining, so an inlinable function's "escaping" local may not escape at a particular call site once the body is spliced into the caller's frame. Verified behavior:
```go
func mk() *Point { return &Point{1, 2} } // inlinable (cost < 80)

a := mk().X // inlined: "&Point{...} does not escape" — no allocation
```
The same call to a `//go:noinline` version reports `&Point{...} escapes to heap`. Practical consequence: keeping hot helper functions under the inliner budget (cost 80) can eliminate heap allocations; conversely, `//go:noinline` in benchmarks can hide allocations your production binary actually performs.
**Follow-ups**: Why does `-gcflags="-m"` sometimes report an escape inside `mk` yet no allocation at the call site? How do you check whether a function was inlined?

### Q: `var a [131073]byte` as a local variable escapes, but `[131072]byte` does not. Why?
**Frequency**: rare-senior
**Why interviewers ask**: It checks whether you know there is a hard size cutoff for stack slots, not just the qualitative "big things escape."
**Answer**: The compiler caps a single non-escaping stack variable at 128 KiB (131072 bytes). One byte over and escape analysis reports `too large for stack` and moves the object to the heap — regardless of whether any reference leaves the frame. The same 128 KiB limit governs the variable-size stack allocation of `make([]byte, n)`: above the limit the generated code takes the heap path at run time. The rationale is bounding worst-case stack growth: a 128 KiB cap keeps frame sizes compatible with the small starting stacks goroutines have.
**Follow-ups**: How would you confirm the exact cutoff on your toolchain? Why is a per-object cap needed at all, given stacks already grow dynamically?

### Q: How does the compiler reason about escapes across function boundaries, and where does that reasoning break down?
**Frequency**: rare-senior
**Why interviewers ask**: It probes the actual mechanism (interprocedural analysis) rather than memorized escape rules.
**Answer**: Escape analysis is interprocedural: the compiler computes an escape "summary" per parameter — visible in `-m` output as `p does not escape` or `leaking param: p` — and callers use those summaries to decide whether an argument may stay on the stack. So passing `&x` to a plain function whose parameter is proven non-escaping keeps `x` local, no inlining required. The analysis breaks down where the callee is unknown at compile time: interface method calls, function values, reflection, and values sent through channels or stored into maps/globals — in those cases the compiler must assume the reference leaks and escapes it to the heap. Use `-gcflags="-m -m"` to see the escape *flow* (`flow: {heap} ← &x:`) naming the exact assignment chain responsible.
**Follow-ups**: Why does storing a pointer into a channel force an escape? How do generics (gcshape stenciling) change this picture, if at all?
