# Topic 06 Easy Questions

### Q: How do you perform a non-blocking channel read?
**Frequency**: often
**Why interviewers ask**: It is the most common `select` idiom and reveals whether you know `default`.
**Answer**: Use a `select` with a `default` case; if no case is ready, `default` runs instead of blocking:
```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("Channel not ready")
}
```
The same pattern works for a non-blocking send (`case ch <- v:`). Note there is no way to "peek" — a successful receive consumes the value.
**Follow-ups**: How would you do a non-blocking send? What happens if you remove the `default` case?

### Q: What is the difference between a goroutine and an OS thread?
**Frequency**: often
**Why interviewers ask**: It is the foundational Go concurrency question and tests whether you understand why Go scales to millions of concurrent tasks.
**Answer**: An OS thread has a large fixed stack (typically 1–8 MB), is scheduled preemptively by the kernel, and context switches involve kernel trap and register/stack swapping. A goroutine starts with a tiny growable stack (2 KB, grows/shrinks as needed), is multiplexed onto a small pool of OS threads by the Go runtime scheduler (GOMAXPROCS threads executing user code), and switches in user space — orders of magnitude cheaper. You can comfortably run hundreds of thousands of goroutines; that many OS threads would exhaust memory.
**Follow-ups**: What does GOMAXPROCS control? When does a goroutine block its underlying OS thread?

### Q: What is the difference between a buffered and an unbuffered channel?
**Frequency**: often
**Why interviewers ask**: Choosing capacity wrongly is a classic source of deadlocks and lost backpressure.
**Answer**: An unbuffered channel (`make(chan T)`) synchronizes sender and receiver: the send blocks until a receiver is ready and vice versa — a direct handoff (rendezvous). A buffered channel (`make(chan T, n)`) lets sends complete without a receiver until the buffer is full; receives block only when the buffer is empty. Unbuffered channels give you a happens-before guarantee between the send and the receive completing; buffered channels trade that synchronization for decoupling.
**Follow-ups**: Does a buffered channel of capacity 1 behave like an unbuffered channel? Which one would you pick for a worker pool, and why?

### Q: What does `for range` over a channel do, and when does the loop end?
**Frequency**: sometimes
**Why interviewers ask**: It tests whether you know the only clean exit condition — and the classic deadlock it hides.
**Answer**: `for v := range ch` receives values until the channel is closed and drained, then the loop terminates. If the channel is never closed, the loop blocks forever (a leak; the runtime flags a fatal deadlock only if *all* goroutines are blocked). Receiving the zero value is not a stop signal — only closure ends the range. Common bug: the sender forgets `close(ch)`, or there are multiple senders and nobody coordinates the close.
**Follow-ups**: Who should close the channel when there are multiple senders? What does the receive `v, ok := <-ch` form give you inside such a loop?
