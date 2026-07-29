# Topic 14 Hard Questions

### Q: What is false sharing in multi-core CPU cache lines and how is it prevented in Go?
**Answer**: False sharing occurs when two goroutines running on separate CPU cores modify independent fields that reside within the same 64-byte L1 cache line, causing cache invalidation ping-pong. It is prevented by padding struct fields with explicit byte arrays (e.g. `_ [64]byte`) to force cache-line alignment.
