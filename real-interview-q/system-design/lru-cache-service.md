# Design: LRU Cache Service

Design an in-memory cache service with TTL and bounded memory, then scale it across nodes. Go focus: `container/list` LRU internals, `sync.Map` vs sharded `RWMutex`, and GC pressure at millions of entries.

---

## 1. Requirements

### Functional
- `Get` / `Set` / `Delete` with per-entry TTL.
- Bounded memory: evict least-recently-used (or least-frequently-used) entries when full.
- Metrics: hit ratio, evictions, per-shard sizes.
- (Distributed variant) shard keys across a fleet of cache nodes with minimal reshuffling when nodes join/leave.

### Non-functional
- O(1) get/put, sub-millisecond p99.
- Survives node loss (cache misses, not data loss — the database remains the source of truth).
- Predictable memory ceiling per node.

## 2. API Sketch

```go
// Cache is a bounded, expiring key-value store.
type Cache interface {
	Get(key string) (val []byte, ok bool)
	Set(key string, val []byte, ttl time.Duration)
	Delete(key string)
}
```

For the distributed variant, routing lives in a client-side library: `node := ring.Lookup(key)` — servers stay dumb, the ring is shared via config.

## 3. High-Level Architecture

```
 ┌────────┐   ┌────────┐        ┌─────────────────────────┐
 │client A│   │client B│        │ config svc (ring state) │
 └───┬────┘   └───┬────┘        └─────────────────────────┘
     │  consistent-hash ring (virtual nodes)
     ▼            ▼
 ┌────────┐  ┌────────┐  ┌────────┐
 │cache   │  │cache   │  │cache   │   each node: sharded LRU
 │node 1  │  │node 2  │  │node 3  │   + TTL janitor + metrics
 └───┬────┘  └───┬────┘  └───┬────┘
     └───────────┴───────────┘
              miss → read-through from database
```

## 4. Data Model

**LRU = map + doubly linked list.** `map[string]*list.Element` for O(1) lookup, `container/list` for O(1) move-to-front and evict-from-back. On `Get`: move element to front. On `Set` when full: evict `list.Back()`.

```go
type entry struct {
	key      string
	val      []byte
	deadline time.Time // zero means no expiry
}

type shard struct {
	mu    sync.Mutex
	items map[string]*list.Element
	lru   *list.List // front = most recently used
}
```

- **TTL:** store a deadline per entry; check lazily on `Get`, plus one janitor goroutine sweeping shards periodically so dead entries of never-read keys don't pin memory.
- **LFU alternative:** `map[key]*node` + `map[freq]*list` of nodes per frequency; evict from the lowest non-empty frequency list. Better hit ratios on skewed workloads, more pointer churn.
- **Sharding:** 64–256 shards by `fnv(key) % n` so one hot key never blocks the whole cache.

## 5. Consistent Hashing for the Distributed Variant

Hash each node into ~150 virtual points on a `uint32` ring (sorted slice, binary search for the successor of `hash(key)`). Adding or removing a node moves only ~1/N of keys, versus nearly all keys with `hash(key) % n`. Virtual nodes smooth out the variance from uneven hash placement.

## 6. Scaling & Bottlenecks

- **Hot keys:** one key can overload one node. Replicate hot keys across nodes with client-side random selection, or cache the hottest entries in-process (L1) with a short TTL.
- **Cache stampede:** on a hot-key miss, hundreds of goroutines hit the database simultaneously. Use singleflight semantics — one goroutine fetches, the rest wait on the shared result.
- **Lock contention:** a single mutex serializes all cores; sharded mutexes scale roughly linearly to shard count. `sync.Map` wins only for read-mostly or disjoint-key workloads and gives no room for per-entry TTL bookkeeping — the sharded `map` + `Mutex` is the right default here.
- **GC pressure:** millions of small entries (map nodes, list elements, `[]byte` values) mean millions of pointers for the GC to scan every cycle. Mitigations: `GOMEMLIMIT` to cap the heap before the OOM killer does, `GOGC` tuning to trade CPU for memory, and slab/arena allocation (pack values into large `[]byte` slabs, store offsets) to cut pointer count. See [core-topics/12](../../core-topics/12-garbage-collection-memory-management/).

## 7. Go-Specific Deep Dives

- The storage half of this design is [capstones/kv-store](../../capstones/kv-store/): in-memory maps with TTL and benchmarks proving the memory ceiling.
- **`container/list` mechanics:** be able to write the `Get` (lookup, `MoveToFront`) and `Set` (insert or update, then evict `Back` while over capacity) paths without notes. Doubly-linked lists trade CPU cache locality for O(1) reorder — fine at cache-node scale.
- **`sync.Map` vs sharded `RWMutex`:** know the two cases `sync.Map` wins (write-once-read-many, disjoint keys) and why eviction metadata disqualifies it — [core-topics/07](../../core-topics/07-concurrency-patterns-sync-package/).
- **GC and escape analysis:** keep values as `[]byte` and avoid interface-typed values in the hot path; every interface box is a potential allocation — [core-topics/05](../../core-topics/05-pointers-memory-allocation-escape-analysis/).
- **Prove it:** benchmarks and `pprof` heap profiles to show hit ratio vs memory, and where allocations come from — [core-topics/14](../../core-topics/14-performance-optimization-profiling/).

## What the Interviewer Is Listening For

- LRU via map + linked list, with the exact operations on each path.
- Sharded locks chosen over `sync.Map`, with the reason.
- TTL handled lazily + janitor, not a timer per key.
- Consistent hashing with virtual nodes, and why `mod n` fails.
- GC pressure named as the real scaling limit of an in-memory Go cache, with `GOMEMLIMIT`/slab mitigation.
