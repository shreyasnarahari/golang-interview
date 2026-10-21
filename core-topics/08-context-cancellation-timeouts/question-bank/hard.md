# Topic 08 Hard Questions

### Q: Explain how parent-child cancellation propagation works in `cancelCtx`.
**Frequency**: rare-senior
**Why interviewers ask**: It probes whether the candidate understands the runtime cost and lifecycle of the context tree, not just the API.
**Answer**: `WithCancel` builds a `cancelCtx` and `propagateCancel` registers it in the nearest parent cancelCtx's `children` map (under the parent's mutex). Calling `cancel()` on a node: stores `err` atomically (first call wins, later calls are no-ops), closes the `done` channel, then recursively cancels every child with `removeFromParent=false`, and finally removes itself from its parent's `children` map. Details worth knowing:
- The `done` channel is created **lazily** on first `Done()` call and stored in an `atomic.Value`, because an atomic load is much cheaper than a mutex in hot paths; canceling a context nobody watches stores a pre-closed channel sentinel instead.
- Because a child unregisters itself when canceled, `defer cancel()` in a call chain also keeps the parent's `children` map from growing.
- Value lookup (`Value`) is a **linear walk up the tree** — O(depth) per call — so deep `WithValue` chains are a hidden cost.
**Follow-ups**: What happens to goroutines blocked on `ctx.Done()` if `cancel` is never called and the parent is long-lived? Why is the first `cancel` call the one that sticks?

### Q: How do `context.WithCancelCause` and `context.Cause` (Go 1.20) work?
**Frequency**: sometimes
**Why interviewers ask**: Cause propagation is the modern way to answer "why did this operation stop" without string-matching sentinel errors.
**Answer**: `WithCancelCause` returns a `CancelCauseFunc` accepting an error that explains the cancellation; `context.Cause(ctx)` retrieves it. Semantics: the **first** cancellation of the context (or a parent) sets the cause, later causes are ignored. If canceled with a plain `CancelFunc` (no cause), `Cause(ctx)` falls back to `ctx.Err()`. So `ctx.Err()` is always `Canceled`/`DeadlineExceeded` (the *mechanism*), while `Cause(ctx)` can carry a domain error (the *reason*):
```go
ctx, cancel := context.WithCancelCause(parent)
cancel(errShardUnavailable) // from another goroutine on failure
// elsewhere:
if err := context.Cause(ctx); errors.Is(err, errShardUnavailable) { ... }
```
There are also `WithDeadlineCause`/`WithTimeoutCause` variants. Cause is stored on the cancelCtx and discovered by walking to the cancelCtx that actually fired.
**Follow-ups**: If the parent is canceled after the child already set a cause, which does `Cause` return on the child? Why keep `Err()` limited to the two sentinel values?

### Q: What does `context.AfterFunc` (Go 1.21) do and why was it added?
**Frequency**: sometimes
**Why interviewers ask**: It reveals whether the candidate tracks the modern API for cleanup-on-cancel and its goroutine semantics.
**Answer**: `AfterFunc(ctx, f)` arranges for `f` to run **in its own goroutine** after `ctx` is canceled; if already canceled, `f` starts immediately. It returns a `stop func() bool`: calling it unregisters `f` and returns true if it prevented the run; false means `f` already started (or was already stopped). Critically, `stop` does **not** wait for a running `f` — if you need to know `f` finished, coordinate explicitly (channel/`WaitGroup`). It replaced the idiomatic-but-leaky pattern:
```go
// old: one goroutine per registration, blocked until ctx done
go func() { <-ctx.Done(); cleanup() }()

// new: no goroutine until cancellation actually happens
stop := context.AfterFunc(ctx, cleanup)
```
Multiple `AfterFunc` calls on one context are independent. If the context type has an `AfterFunc` method, the custom implementation is used.
**Follow-ups**: Why does running `f` in a fresh goroutine (rather than inline in the canceler) matter? What happens if `stop` is called concurrently with cancellation?

### Q: If a parent context has a 2-second deadline and a child is derived with `WithTimeout(parent, 10*time.Second)`, what is the child's effective deadline and how is it implemented?
**Frequency**: sometimes
**Why interviewers ask**: Deadline min-propagation is the mechanic behind cascading RPC timeouts, and the implementation detail exposes shallow understanding.
**Answer**: The effective deadline is the **minimum** — 2 seconds. `WithDeadline` checks `parent.Deadline()` first: if the parent's deadline is already earlier than the requested one, it skips creating a `timerCtx` entirely and returns a plain `WithCancel(parent)` (no new timer is allocated, since the parent's timer will fire first). Otherwise a `timerCtx` is created holding a `time.Timer` set to the child's deadline; canceling the child early stops that timer (which is why omitting `defer cancel()` leaks the timer until it fires). So per-RPC budgets can only **tighten**, never extend, the caller's remaining budget — a deliberate design that prevents a callee from outliving its caller's deadline.
**Follow-ups**: How should a client split a remaining deadline across sequential RPC calls? What does `Deadline()` return for a context derived from `Background()`?
