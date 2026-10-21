# Topic 13 Real Interview Q&A

### Q: (Asked at a storage-engine / infrastructure company) "We cache object addresses as `uintptr` values in a table for fast lookup. Any problem with that?"
**Frequency**: sometimes
**Why interviewers ask**: A classic trap: the design looks like a performance win but fundamentally misunderstands Go's memory model.
**Answer**: Yes — `uintptr` is an integer holding an address, **not a pointer**. The garbage collector does not treat `uintptr` values as live references: if no real pointer to the object remains, the GC will free it, leaving your table of addresses dangling. Even if another reference keeps the object alive, Go reserves the right for a future compacting collector to **move** objects, invalidating every cached address. The fix depends on intent: keep real pointers (`*T`) or `unsafe.Pointer` (which the GC does track and update) for identity; if C code or a stable integer ID is needed, use `runtime/cgo.Handle`. The only legitimate `uintptr` use is transient arithmetic inside a single expression, converted back to `unsafe.Pointer` immediately.
**Follow-ups**: How would you detect this class of bug in CI before it corrupts data in production? What would `GODEBUG=cgocheck` / `-d=checkptr` catch here, and what would it miss?

### Q: (Asked at a database company using C libraries for compression/crypto) "Why is cgo significantly slower than pure Go function calls, and how would you design around it?"
**Frequency**: often
**Why interviewers ask**: cgo is unavoidable in real systems (rocksdb, openssl, zstd); the interviewer wants cost-aware design, not dogma.
**Answer**: A cgo call must transition the goroutine out of the Go scheduler (the M enters a syscall-like state, similar to `entersyscall`), switch from the goroutine's **contiguous growable stack** to the C/OS stack, fix up arguments to the C calling convention, and reverse all of it on return — roughly ~50–100 ns versus ~1 ns for an inlinable pure-Go call. While C code runs, that OS thread cannot run other goroutines, so a long C call that doesn't block in a syscall still consumes an M. GC is not "disabled"; the goroutine is simply in a state where its Go stack needn't be scanned until it returns. Additionally, passing Go pointers into C is restricted by the cgo pointer-passing rules, adding `cgocheck` validation cost. Design mitigations: **batch work per crossing** (compress a whole block, not per record), keep hot small helpers in pure Go or assembly, copy data into C-allocated buffers (`C.malloc`/arena) to avoid pointer-rule overhead, and never call into cgo inside a per-goroutine hot loop — pool C contexts instead.
**Follow-ups**: How would you benchmark whether a cgo call is worth it for a given payload size? What happens to `GOMAXPROCS` and tail latency when many goroutines sit in long C calls simultaneously?

### Q: (Asked at a low-latency trading firm) "When, if ever, would you approve `unsafe` in production code? Give a concrete example and how you'd keep it safe."
**Frequency**: sometimes
**Why interviewers ask**: They want judgment: knowing the escape hatch, its measurable payoff, and the safety harness around it.
**Answer**: Approve it only behind a measured bottleneck, wrapped in a small audited API. A concrete example: a market-data parser allocating millions of `string(b)` conversions per second from a reused byte buffer. Each conversion copies; replacing it with a zero-copy view for the lifetime of one message —

```go
s := unsafe.String(&buf[off], n) // buf is reset before the next message
```

— removes the allocation and copy entirely. Guardrails: (1) the returned string must not outlive the buffer (never store it in maps/structs that escape the parse loop); (2) never mutate through a `[]byte` view of a real string — strings are immutable and literals may sit in read-only pages; (3) isolate `unsafe` in one file with fuzz + unit tests; (4) run CI with `-race` and `-gcflags=all=-d=checkptr=2`, plus `-asan` if cgo is involved; (5) benchmark first — Go 1.24's Swiss-table maps and improved allocator already removed many historical motivations for unsafe tricks.
**Follow-ups**: How would you fuzz a zero-copy parser to prove the view never escapes? What Go 1.20+ APIs replaced the old `reflect.SliceHeader` forging pattern, and why was the old pattern dangerous?
