# Topic 07 Hard Questions

### Q: How does `sync.Map` work internally to avoid lock contention?
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates who have read the standard library source from those who only know the API.
**Answer**: The answer depends on the Go version — a strong candidate names both designs.

**Since Go 1.24**, `sync.Map` is a thin wrapper over `internal/sync.HashTrieMap`: a lock-free concurrent hash trie. Keys are hashed and placed in a tree of nodes; loads walk the trie using only atomic pointer loads (no locks at all), while stores rebuild only the small path from the root to the affected node using compare-and-swap (copy-on-write), so readers never block and writers only conflict on the same trie path.

**Before Go 1.24** (the classic interview answer), `sync.Map` used two maps: an immutable `read` map accessed locklessly via atomic loads, and a mutex-protected `dirty` map holding new entries. A read missing from `read` fell back to `dirty` under the mutex and incremented a `misses` counter; once `misses >= len(dirty)`, `dirty` was *promoted* to become the new `read`, amortizing lock cost. Deletes of keys in `read` were just an atomic CAS on the entry pointer using an `expunged` sentinel — no map mutation.

Either way, the design goal is the same: make reads of existing keys lock-free, which is why `sync.Map` wins for write-once/read-many caches and disjoint key sets, and loses to `RWMutex + map[K]V` for write-heavy or churning workloads.
**Follow-ups**: Why does the classic read/dirty design degrade when keys churn? What did Go 1.24's hash trie gain besides removing the mutex from reads?

### Q: How does `sync.RWMutex` prevent writer starvation, and why can recursive read-locking deadlock?
**Frequency**: rare-senior
**Why interviewers ask**: Writer-reader fairness is where naive `RWMutex` usage breaks under load; the deadlock case bites real code.
**Answer**: Internally an `RWMutex` keeps an atomic reader count plus a writer semaphore. `RLock` increments the count; a writer calling `Lock` makes the count **negative** (subtracting `rwmutexMaxReaders`), so every *subsequent* `RLock` sees the negative count and queues behind the pending writer — a pending writer blocks new readers. The writer then waits until the readers that were already inside drain (`RUnlock` wakes the writer when the count returns to the negative floor), and the writer's `Unlock` wakes all queued readers. So a writer waits at most for the current generation of readers, never indefinitely.

The deadlock: if goroutine A holds `RLock` and calls `RLock` again (directly or via a helper), and a writer arrives between the two read-locks, the second `RLock` blocks behind the writer, while the writer waits for A's first `RLock` to be released — a cycle. Go forbids "upgrading" reads to writes for the same reason; recursive read-locking is only safe if no writer can ever arrive between the two acquisitions (in practice: never rely on it).
**Follow-ups**: Why is a plain `Mutex` often faster than `RWMutex` even in read-heavy workloads? How would you detect this deadlock in production?

### Q: How does `sync.Pool` work internally, and why can items disappear at any time?
**Frequency**: rare-senior
**Why interviewers ask**: It tests understanding of Go's scheduler (P's) and GC interaction, and when *not* to use a pool.
**Answer**: `sync.Pool` keeps a per-P (per-logical-processor) private pool plus a per-P shared slice that other P's may steal from, so the common `Get`/`Put` path on the local P takes no lock at all. To survive GC without pinning memory, each P also keeps a **victim cache**: at every GC, the primary pool is demoted to "victim" and the previous victim cache is dropped entirely. A `Get` that misses the primary pool tries the victim cache before allocating fresh. Net effect: pooled objects survive at most roughly one GC cycle, and the pool is emptied at GC time — that is why the docs say any item may be removed automatically without notice. Consequences: use `Pool` for reusable scratch memory (buffers, temporary structs) to reduce allocation pressure, never for things whose lifetime matters (connections, sessions), and always reset state before `Put` to avoid leaking data between users.
**Follow-ups**: Why does per-P storage scale better than a single mutex-protected freelist? What kind of workload actually gets slower with `sync.Pool`?

### Q: In a fan-in/pipeline design, who should close the output channel, and what leaks if you get shutdown wrong?
**Frequency**: sometimes
**Why interviewers ask**: Goroutine leaks from mis-closed pipelines are among the most common production Go bugs.
**Answer**: The convention: the **sender closes the channel**, never the receiver, and when there are multiple senders (fan-in), none of them close it directly — a dedicated closer waits on a `sync.WaitGroup` and closes once all senders are done:
```go
var wg sync.WaitGroup
for _, in := range inputs {
    wg.Add(1)
    go func(ch <-chan int) {
        defer wg.Done()
        for v := range ch {
            out <- v // blocks forever if consumer is gone
        }
    }(in)
}
go func() { wg.Wait(); close(out) }()
```
Failure modes: a producer that sends on a closed channel panics; a consumer that ranges over a channel nobody closes blocks forever; and the fan-in goroutine above leaks if `out` has no reader — every stage must also respect cancellation (`select { case out <- v: case <-ctx.Done(): return }`) so a downstream exit propagates upstream instead of stranding goroutines. Rule of thumb: every goroutine needs a guaranteed path to exit, and channel close signals "no more values" downstream while `context.Context` signals "stop working" upstream.
**Follow-ups**: Why not just close every input channel from the receiver when it is done reading? How does this change if stages must preserve ordering?
