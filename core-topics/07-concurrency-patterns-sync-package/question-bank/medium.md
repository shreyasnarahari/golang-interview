# Topic 07 Medium Questions

### Q: What is the difference between starvation mode and normal mode in `sync.Mutex`?
**Answer**:
- **Normal Mode**: Waiters join a FIFO queue, but newly arrived goroutines have an advantage to acquire the lock since they are already running on the CPU.
- **Starvation Mode**: Triggered if a waiter fails to acquire the lock for $>1\text{ ms}$. Ownership is handed directly from the unlocking goroutine to the front of the queue, preventing tail latency starvation.
