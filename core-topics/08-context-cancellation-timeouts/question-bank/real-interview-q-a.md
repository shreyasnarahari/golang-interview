# Topic 08 Real Interview Q&A

### Q: (Asked at a high-throughput payments company) "Why do you insist that `cancel()` always be called after `context.WithTimeout`, even on success paths?"
**Frequency**: often
**Why interviewers ask**: Skipped cancels are a real memory-leak pattern they have debugged in production under load.
**Answer**: `WithTimeout` allocates a `timerCtx` holding a `time.Timer` and registers the child in the parent's `children` map. If `cancel()` is never called: (1) the timer lives in the runtime's timer heap until it fires — at millions of requests with long timeouts that is real memory and timer-wheel churn; (2) the child stays linked in the parent's map, so a long-lived parent accumulates dead children. Calling `cancel()` stops the timer immediately and removes the child from the parent. The idiom is:
```go
ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
defer cancel()
```
`defer` covers early returns and panics too. Note `cancel` is idempotent — only the first call does anything, so double-cancel is harmless.
**Follow-ups**: How would you detect this leak in a running service (pprof, timer counts)? Does the leak also occur with plain `WithCancel`?

### Q: (Asked at a cloud infra company) "An operation failed and you have a context — how do you distinguish 'our caller gave up' from 'our own timeout fired' from 'we were explicitly canceled'?"
**Frequency**: often
**Why interviewers ask**: Retry policy, metrics attribution, and user-facing errors all hinge on classifying the stop reason correctly.
**Answer**: `ctx.Err()` returns exactly one of two sentinels — `context.Canceled` (explicit cancel, or parent canceled) or `context.DeadlineExceeded` (a deadline in the chain fired) — and is nil while the context is live. Check with `errors.Is` since wrapped errors propagate. To distinguish *whose* deadline fired, compare the operation's own derived context against the parent: if `opCtx.Err() == context.DeadlineExceeded` but `parentCtx.Err() == nil`, it was our local budget; if the parent is also done, the caller gave up or a shared upstream deadline fired. Since Go 1.20, `context.Cause(ctx)` can carry the domain reason when cancellation came from `WithCancelCause` (first cause wins), falling back to `Err()` otherwise.
**Follow-ups**: Should a canceled-context error map to HTTP 499, 504, or 500 at the edge? Why must you check `ctx.Err()` before interpreting a partial result?

### Q: (Asked at a platform engineering team) "Walk me through graceful shutdown of an HTTP server using context."
**Frequency**: often
**Why interviewers ask**: It combines signal handling, deadline budgeting, and in-flight request draining — a complete context story in one design.
**Answer**: Three contexts work together:
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
go srv.ListenAndServe()
<-ctx.Done()                          // signal arrived (or was canceled)
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)             // stop accepting, drain in-flight requests
```
Key decisions: `signal.NotifyContext` (Go 1.16+) replaces manual `signal.Notify` + channel plumbing and lets `stop()` unregister the signal; `Shutdown` needs a **fresh** context — not the canceled signal context — with its own deadline bounding how long we wait for in-flight requests. Note that `Shutdown` does **not** cancel in-flight request contexts: it closes listeners and idle connections, then waits for active connections to return to idle, so handlers must either finish on their own or watch their own `r.Context()`/`ctx.Done()` to abort early (the shutdownCtx deadline is the backstop — if it expires, `Shutdown` returns the context error). Anything that must finish after shutdown starts (audit logs, metrics flush) uses `context.WithoutCancel` plus its own short timeout.
**Follow-ups**: What happens to handlers that ignore their request context during shutdown? How would you also shut down background workers using the same signal context?
