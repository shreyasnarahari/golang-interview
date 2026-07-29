# Topic 08 Real Interview Q&A

### Q: Why should `cancel()` always be called when creating a `WithTimeout` context?
**Answer**: `WithTimeout` attaches a timer to the parent context. If `cancel()` is not called, the timer remains active in memory until the timeout fires, leaking resources even if the work finishes early.

### Q: What does `ctx.Err()` return after cancellation vs timeout?
**Answer**:
- Cancellation: `context.Canceled`
- Timeout: `context.DeadlineExceeded`
