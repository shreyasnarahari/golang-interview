# Topic 07 Hard Questions

### Q: How does `sync.Map` work internally to avoid lock contention?
**Answer**: `sync.Map` uses two internal maps:
1. `read atomic.Value` (`readOnly` struct holding `map[any]*entry`): Read-only map accessed locklessly using atomic loads.
2. `dirty map[any]*entry`: Map protected by a `Mutex` holding newly written keys or keys marked amended.
Reads first check `read`. If missing and `amended==true`, it falls back to `dirty` under mutex.
