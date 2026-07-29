# Topic 11 Medium Questions

### Q: How does Go 1.14+ implement asynchronous preemption?
**Answer**: By registering a `SIGURG` signal handler on OS threads. When `sysmon` detects a long-running goroutine, it emits `SIGURG` to the thread **M**, triggering signal handler execution which inserts a preemption point onto the goroutine's stack.
