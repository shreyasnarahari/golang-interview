# Topic 03 Real Interview Q&A

### Q: Can a `recover()` in a parent goroutine catch a panic from a spawned child goroutine?
**Answer**: No. Panics unwind the stack of the specific goroutine where the panic occurred. If a child goroutine panics without its own `deferred recover()`, the entire Go process crashes.

### Q: When are deferred function arguments evaluated?
**Answer**: Arguments are evaluated immediately when the `defer` statement is executed.
