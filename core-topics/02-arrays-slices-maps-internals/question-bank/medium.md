# Topic 02 Medium Questions

### Q: Predict the output: a function receives a slice, mutates `s[0]`, then appends — which changes does the caller see?
**Frequency**: often
**Why interviewers ask**: It tests whether you truly understand that a slice header is copied by value while the backing array is shared.
**Answer**: The caller sees element mutations (`s[0] = x`) because both headers point at the same backing array, but does **not** see appended elements: `append` may grow the caller's local header (or reallocate), and the caller's `len` is unchanged. Even if no reallocation happens, the new element sits beyond the caller's `len`.

```go
func mutate(s []int) {
	s[0] = 99        // visible to caller
	s = append(s, 7) // invisible: caller's len/cap unchanged
}
```
If the callee needs to return a grown slice, it must `return s`. (For the rare in-place case there is the go1.21+ `slices.Concat`-style approach, but the idiom is: return the new slice.)
**Follow-ups**: When does `append` write into the shared backing array vs. allocate a new one — and why is that an aliasing hazard? How would you defensively copy a slice (`slices.Clone`)?

### Q: Why does deleting keys from a map not shrink its memory usage?
**Frequency**: sometimes
**Why interviewers ask**: Long-lived caches that "leak" memory despite `delete` calls are a classic production incident, and the fix reveals map-internals knowledge.
**Answer**: `delete` marks slots empty (pre-1.24: clears bucket entries; go1.24+ Swiss tables: marks the slot empty or leaves a tombstone), but the allocated bucket/group storage stays so future inserts don't pay reallocation. Shrinking only happens via rehash triggered by growth heuristics, never by deletion alone. Verified behavior: a map that grew to millions of entries keeps most of that footprint after deleting everything. Fixes: periodically rebuild the map (copy live entries into a fresh map), shard it so individual shards can be dropped, or store pointers so the per-slot overhead stays small.
**Follow-ups**: Does the same apply to slices — how do you release a backing array? How do tombstones in Swiss tables interact with this?

### Q: Why is map iteration order randomized in Go?
**Frequency**: often
**Why interviewers ask**: It checks whether you've internalized that map order is deliberately unspecified, not just "implementation-defined".
**Answer**: The runtime seeds each iteration with a random start (pre-1.24: random bucket + offset; go1.24+: random group/offset in the Swiss table), so two `for range` loops over the same map usually produce different orders. This is intentional: it prevents developers from depending on hash or insertion order, which would break whenever the map grows and rehashes. Consequences: never serialize a map by ranging if you need deterministic output — collect keys, `slices.Sort` them, then emit. Note also that adding/removing entries during iteration may or may not surface those entries in that same loop.
**Follow-ups**: How would you dump a map in sorted key order? Is it safe to delete keys while ranging over the map?

### Q: What happens when two goroutines access a map concurrently, and how do you detect it?
**Frequency**: often
**Why interviewers ask**: Concurrent map misuse is one of the few data races Go refuses to tolerate silently, and knowing the failure mode (and that you can't recover from it) matters.
**Answer**: A concurrent write with any other access triggers a **fatal error** — `fatal error: concurrent map writes` or `concurrent map read and map write`. This is deliberately not an ordinary panic: `recover()` cannot catch it; the process dies, because the runtime detected a corrupted hash table it cannot safely continue with. Concurrent reads alone are fine. Detect races in tests with `go test -race`. Fixes: guard with `sync.Mutex`/`sync.RWMutex`, or use `sync.Map` for the narrow cases it wins (write-once/read-many caches, disjoint key sets per goroutine).
**Follow-ups**: When is `sync.Map` actually faster than a mutex-guarded map, and when is it slower? Why does the runtime make this fatal instead of returning an error?
