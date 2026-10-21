# Topic 12: Garbage Collection & Memory Management

## 📌 Overview
Go features a concurrent, low-latency **Tri-Color Mark-Sweep Garbage Collector** designed to keep Stop-The-World (STW) pauses below **1 millisecond**.

---

## 🎯 Key Concepts

### 1. Tri-Color Abstraction
1. **White**: Candidate objects for collection. At start of GC, all heap objects are White.
2. **Grey**: Reachable objects visited, but child object pointers are not yet scanned.
3. **Black**: Reachable objects visited along with all their child object pointers.

```
GC Mark Progression:
[White: Unvisited] ----> [Grey: In-Progress] ----> [Black: Confirmed Reachable]
```

### 2. Memory Allocator (`mcache` / `mcentral` / `mheap`)
Go's heap allocator is based on Google's **TCMalloc** (Thread-Caching Malloc) and has three tiers:

```
Goroutine on P
      |
      v
+-----------+     +-----------+     +-----------+      +--------+
|  mcache   |---->|  mcentral |---->|   mheap   |----->|   OS   |
| (per-P,   |     | (67 size  |     | (8 KB     | mmap | (mmap) |
| lock-free)|     | classes)  |     |  pages)   |      +--------+
+-----------+     +-----------+     +-----------+
```

1. **`mcache`**: Per-**P** local cache of spans. Small object allocation is **lock-free** because a P is owned by at most one running M at a time.
2. **`mcentral`**: Shared central cache partitioned into **67 size classes** spanning **8 B to 32 KB**. When an `mcache` span runs dry, it refills from the matching `mcentral` size class (requires a lock).
3. **`mheap`**: Global page allocator managing the heap in **8 KB pages**. Objects **larger than 32 KB** bypass `mcache`/`mcentral` entirely and are allocated directly from `mheap`, which grows the heap from the OS via `mmap`.

### 3. Hybrid Write Barrier
To run GC concurrently with active user mutator goroutines, Go 1.8+ implements a **Hybrid Write Barrier** (a combination of the Dijkstra insertion barrier and the Yuasa deletion barrier). On every pointer write during marking it:
- Shades the **newly written pointer** Grey (insertion barrier).
- Shades the **overwritten pointer** Grey (deletion barrier) — the old value being replaced.

Because **both sides** of every pointer write are shaded, stacks do not have to be re-scanned during marking, which eliminates the stack re-scan at STW mark termination and keeps pauses below **1 ms**.

### 4. Tuning GC Performance
- **`GOGC`** (Default 100): Pacing factor. Triggers GC when heap size increases by 100% over live set after previous collection.
- **`GOMEMLIMIT`** (Go 1.19+): Soft memory cap (e.g. `GOMEMLIMIT=2GiB`). Prevents OOM kills in Kubernetes containers by automatically adjusting `GOGC` target pacing dynamically when memory usage nears the threshold.
- **Programmatic equivalents**: `debug.SetGCPercent(n)` sets `GOGC` at runtime (`debug.SetGCPercent(-1)` disables GC, like `GOGC=off`); `debug.SetMemoryLimit(bytes)` sets `GOMEMLIMIT` at runtime. Both return the previous value.

---

## ⚠️ Common Gotchas
1. **Memory Ballast Anti-Pattern**: Setting up dummy huge byte arrays (`make([]byte, 10<<30)`) was common in legacy Go to delay GC cycles, but is now **obsolete** due to `GOMEMLIMIT` in Go 1.19+.
