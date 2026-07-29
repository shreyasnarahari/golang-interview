# Topic 07 Real Interview Q&A

### Q: What happens if you pass a `sync.Mutex` by value to a function?
**Answer**: Go copies the mutex struct, including its lock state. The function operates on a separate copy, breaking mutual exclusion and leading to data races.

### Q: When should you use `sync.Map` over `sync.Mutex` + standard `map`?
**Answer**: `sync.Map` is optimized for write-once/read-many workloads or disjoint key spaces across goroutines. For general read/write maps, `sync.RWMutex` + `map[K]V` offers better performance and type safety.
