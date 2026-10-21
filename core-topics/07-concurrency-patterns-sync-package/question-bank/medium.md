# Topic 07 Medium Questions

### Q: What is the difference between starvation mode and normal mode in `sync.Mutex`?
**Frequency**: sometimes
**Why interviewers ask**: It tests whether you understand Go's fairness trade-off rather than treating the mutex as a black box.
**Answer**: **Normal mode**: blocked waiters join a FIFO queue, but when the lock is released, a *newly arriving* goroutine competes directly with the head of the queue — and usually wins because it is already running on a CPU while the queued waiter still has to be woken. This maximizes throughput but can starve queued waiters. **Starvation mode**: if a waiter has been queued for more than 1 ms (`starvationThresholdNs` in the runtime), the mutex switches to starvation mode: `Unlock` hands ownership directly to the front of the queue and new arrivals must queue even if the lock appears free. This bounds tail latency. The mutex leaves starvation mode when a waiter is the last in the queue or its wait was under 1 ms. Bottom line: normal mode favors throughput, starvation mode bounds tail latency.
**Follow-ups**: Why does letting newcomers barge in improve throughput? Does `sync.RWMutex` have the same two modes?

### Q: What is wrong with calling `wg.Add(1)` inside the goroutine instead of before `go`?
**Frequency**: often
**Why interviewers ask**: This is a classic code-review bug that separates people who have debugged real races from those who have not.
**Answer**: It races with `wg.Wait()`: the parent may reach `Wait` before the child goroutine ever gets scheduled and calls `Add`, so `Wait` sees a zero counter and returns immediately — the parent proceeds (or exits) while children are still running. It can also trigger the runtime panic "sync: WaitGroup is reused before previous Wait has returned" on repeated runs.
```go
// BUG: Wait can return before Add executes.
go func() {
    wg.Add(1) // too late
    defer wg.Done()
    work()
}()

// Correct: counter incremented before the goroutine starts.
wg.Add(1)
go func() {
    defer wg.Done()
    work()
}()
```
The memory model only guarantees ordering from `Add` to `Wait` when `Add` happens (or is sequenced) before the goroutine launch.
**Follow-ups**: How would the race detector catch this? How does `wg.Go` (Go 1.25) eliminate this class of bug?

### Q: If the function passed to `sync.Once.Do` panics, does the `Once` consider itself done? What about `sync.OnceFunc`?
**Frequency**: sometimes
**Why interviewers ask**: It probes a subtle semantic difference that matters for retry-on-failure initialization.
**Answer**: With plain `Once.Do`, a panic **counts as done** — `Do` defers marking the `Once` complete before running `f`, so subsequent calls to `Do` return immediately without retrying, even though initialization failed (and the panic still propagates to the caller). This is a common footgun for fallible initialization. The Go 1.21 wrappers behave differently: `sync.OnceFunc` / `OnceValue` / `OnceValues` only mark themselves done when `f` returns *normally*; if `f` panics, the next call to the wrapped function invokes `f` again, giving you retry semantics. For fallible setup, prefer `OnceValues` returning `(T, error)` and handle the error, or manage retries explicitly.
**Follow-ups**: How would you implement retryable one-time initialization before Go 1.21? Is `sync.Once` safe to use concurrently from many goroutines on the first call?

### Q: Why is the condition of a `sync.Cond` checked in a `for` loop, and when do you use `Signal` versus `Broadcast`?
**Frequency**: sometimes
**Why interviewers ask**: `Cond` misuse is invisible until production; the loop idiom reveals whether you understand wakeup semantics.
**Answer**: `Wait` must be called with `c.L` held; it atomically unlocks, suspends, and re-locks before returning. The predicate is checked in a `for` loop because waking up does **not** guarantee the condition still holds — another goroutine may have consumed the resource between `Signal` and your `Wait` returning, or `Broadcast` may wake more waiters than can proceed.
```go
c.L.Lock()
for !ready() {
    c.Wait()
}
// ... consume; condition guaranteed held here ...
c.L.Unlock()
```
Use `Signal` when exactly one waiter can make progress (e.g., one item added to a queue); use `Broadcast` when the state change may unblock many waiters (e.g., shutdown flag flipped, configuration reloaded). In modern Go, channels, `context`, or a plain mutex around a counter often express the same logic more clearly — `Cond` is mostly for low-level library code.
**Follow-ups**: Why must the lock be held when calling `Signal`/`Broadcast` in practice, even though the docs do not strictly require it? What happens if you call `Wait` without holding the lock?
