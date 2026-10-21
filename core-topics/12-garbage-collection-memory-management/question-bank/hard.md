# Topic 12 Hard Questions

### Q: Explain the hybrid write barrier introduced in Go 1.8.
**Frequency**: sometimes
**Why interviewers ask**: It separates candidates who recite "tri-color GC" from those who understand *why* marking can run concurrently without losing objects.
**Answer**: The barrier is Clements/Hudson's hybrid of Dijkstra's *insertion* barrier and Yuasa's *deletion* barrier. While marking is active, every pointer write `*slot = ptr` shades **both** pointers grey:
```go
if writeBarrier.enabled {
    shade(ptr)     // the newly written pointer (Dijkstra insertion side)
    shade(*slot)   // the pointer being overwritten (Yuasa deletion side)
}
*slot = ptr
```
Shading the new pointer prevents a reachable object from being hidden behind an already-scanned (black) object; shading the *overwritten* pointer protects the old value, because stacks are never re-scanned after their initial scan — a stack frame's locals are treated as fixed roots. The payoff: **no stack re-scan during mark termination**, which was the dominant STW cost pre-1.8, bringing pauses to the sub-millisecond range. The barrier compiles to a flag check plus a write-buffer push, so it costs a few instructions on every pointer write even when GC is idle.
**Follow-ups**: Why does the Yuasa half let stacks avoid re-scanning? What is the cost of the barrier when GC is off, and how does the compiler keep it cheap?

### Q: Walk through the GC pacer: when exactly does a cycle trigger, and how do GOGC and GOMEMLIMIT interact?
**Frequency**: sometimes
**Why interviewers ask**: Pacing is where GC theory meets production symptoms (CPU spikes, OOM kills); senior candidates should be able to reason about the trigger math.
**Answer**: After each cycle the runtime measures the **live heap** `L` and computes the next cycle's **heap goal**: `L × (1 + GOGC/100)` (with a hard "overshoot" bound about 1.1× above it, beyond which allocation stalls for marking). The cycle doesn't fire at the goal; the pacer aims to *complete* marking by the goal, so it starts marking early at a **trigger** point below the goal, computed from measured marking speed vs. allocation rate. Two pressure valves keep this safe: **mutator assists** draft fast-allocating goroutines into marking work, and dedicated background mark workers consume up to ~25% of GOMAXPROCS. `GOMEMLIMIT` clamps the heap goal downward: when total runtime memory nears the limit the goal shrinks and GC runs more often, and if the app stays pinned at the limit the runtime caps GC CPU at 50% to avoid a thrashing death spiral. With `GOGC=off`, only the memory limit (and `runtime.GC()`) triggers cycles.
**Follow-ups**: What happens if marking is projected to finish late (assists ramp up)? Why is the memory limit *soft* even though it clamps the goal?

### Q: Describe Go's memory allocator: mcache, mcentral, mheap.
**Frequency**: often
**Why interviewers ask**: It's the classic "tcmalloc-derived" internals question that tests whether you understand allocation cost and fragmentation, not just GC.
**Answer**: A tcmalloc-style three-tier allocator over ~67 size classes (8 B up to 32 KiB) carved out of 64 MiB **arenas** on 64-bit platforms:
- **mcache**: per-P cache of spans, one span per size class. Allocations from mcache need **no lock** — this is why Go's allocator scales with cores. The **tiny allocator** (objects < 16 B with no pointers) packs several objects into one 16 B block.
- **mcentral**: per-size-class central free lists (partial/full span sets). When an mcache span empties, it refills from the matching mcentral under that mcentral's lock.
- **mheap**: the global heap. It owns all spans (runs of contiguous 8 KiB pages), refills mcentrals, and grows by requesting new arenas from the OS. Objects **larger than 32 KiB** bypass size classes and are allocated directly from mheap.

Sweeping returns free slots to spans; spans that stay empty can have their pages released back to the OS (scavenger, paced by `debug.SetMemoryLimit`/heap goal history).
**Follow-ups**: Why per-P instead of per-goroutine caches? How does this design cause RSS to stay high after a spike (span fragmentation / lazy scavenging)?

### Q: What is the Green Tea GC and how does it change the classic tri-color marking loop?
**Frequency**: rare-senior
**Why interviewers ask**: It probes whether a senior candidate follows current runtime development — Green Tea shipped as an experiment in Go 1.25 and is the default marking algorithm from Go 1.26 onward.
**Answer**: Classic tri-color marking scans objects one at a time as they are popped from a LIFO work buffer, so consecutive scans jump randomly around the heap — terrible cache locality on large heaps. **Green Tea** keeps tri-color invariants but makes marking **span-centric**: instead of enqueueing each grey object, it sets the object's mark bit and enqueues its *span* on a FIFO queue. When a span is dequeued, the marker scans all newly-marked objects on that span together (tracked with a second "scanned" bitset; the intersection of mark and scan bits identifies work, keeping the GC precise). Batching objects within a span improves locality, enables prefetching, and amortizes metadata lookups; the Go team reports roughly 10–40% reductions in GC CPU overhead on allocation-heavy, large-heap workloads. Write barriers, the pacer, and STW structure are unchanged.
**Follow-ups**: Why does a FIFO span queue accumulate more work per span than the old LIFO object buffer? Which workloads benefit least (pointer-light heaps, tiny heaps that fit in cache)?
