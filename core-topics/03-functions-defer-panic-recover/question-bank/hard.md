# Topic 03 Hard Questions

### Q: Explain open-coded defers introduced in Go 1.14.
**Frequency**: rare-senior
**Why interviewers ask**: It gauges runtime-level understanding of why defer is no longer a performance bogeyman.
**Answer**: For functions whose `defer` statements are **not inside loops** (and number at most 8), the compiler eliminates `_defer` struct allocations by emitting the deferred calls inline at each return point, with a bitmask tracking which defers are active. Defers inside loops still go through the runtime (their records are stack-allocated since Go 1.13). Combined with 1.13's stack-allocated defer records, this cut typical defer overhead from roughly 35ns (pre-1.13 heap-allocated path) to about 1ns — which is why modern Go style guides say defer is fine even in hot paths, as long as it is not in a loop.
**Follow-ups**: Why can't defers inside loops be open-coded, and what goes wrong if you `defer rows.Close()` in a loop?

### Q: What changed about `panic(nil)` in Go 1.21, and why?
**Frequency**: rare-senior
**Why interviewers ask**: It tests whether the candidate tracks modern Go changes and understands the classic `if recover() != nil` footgun.
**Answer**: Before Go 1.21, `panic(nil)` caused `recover()` to return `nil` — indistinguishable from "no panic happened", which broke the universal `if r := recover(); r != nil` guard. Since Go 1.21, `panic(nil)` produces a non-nil `*runtime.PanicNilError` (whose message is "panic called with nil argument"), so the guard works. Verified on current Go: `recover()` returns a non-nil `*runtime.PanicNilError`. The old behavior can be restored with `GODEBUG=panicnil=1`.
**Follow-ups**: Name another GODEBUG-controlled behavior transition from recent Go releases (e.g. `http2xconnect`, `randseednop`, `gotypesalias`).

### Q: What happens if a deferred function itself panics while a panic is already unwinding?
**Frequency**: rare-senior
**Why interviewers ask**: Panic-during-panic is the sharp edge that decides whether cleanup code can accidentally destroy diagnostic information.
**Answer**: The new panic starts a fresh unwinding sequence that *replaces* the old one: remaining deferred functions still run in LIFO order, and a later `recover()` sees only the **new** panic value — the original panic value is lost unless you captured it earlier. So in cleanup code that can fail (e.g. `defer f.Close()` where Close can panic), prefer returning errors over panicking, or explicitly chain: recover first, then re-panic with both values wrapped. If no deferred function recovers the replacement panic, the process crashes with the new value.
**Follow-ups**: How would you preserve both panic values — e.g. with `fmt.Errorf("cleanup failed during panic %v: %w", orig, cerr)`?

### Q: What does `runtime.Goexit()` do, and how does it interact with deferred functions and `recover()`?
**Frequency**: rare-senior
**Why interviewers ask**: Goexit sits at the intersection of goroutine lifecycle and defer semantics — almost nobody knows it cold, which makes it a good senior discriminator.
**Answer**: `runtime.Goexit()` terminates the **calling goroutine** only, but unlike a panic it is not an error: all deferred functions still run (verified — deferred calls execute normally), `recover()` inside them returns `nil` because there is no panic, and other goroutines keep running. If the *main* goroutine calls Goexit (or returns while others run Goexit patterns leaving none runnable), the runtime throws a fatal deadlock error. Real use: terminating a worker goroutine cleanly while guaranteeing its defers run; it also underlies how `testing.T.FailNow` works (it calls Goexit, which is why FailNow must be called from the test's own goroutine).
**Follow-ups**: Why must `t.FailNow()` be called from the test goroutine rather than a helper goroutine?
