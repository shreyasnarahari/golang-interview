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

### 2. Hybrid Write Barrier
To run GC concurrently with active user mutator goroutines, Go implements a **Hybrid Write Barrier**:
- Shades any newly created or mutated object pointer **Grey**.
- Eliminates the need to re-scan mutator stacks during STW mark termination, maintaining ultra-low latency.

### 3. Tuning GC Performance
- **`GOGC`** (Default 100): Pacing factor. Triggers GC when heap size increases by 100% over live set after previous collection.
- **`GOMEMLIMIT`** (Go 1.19+): Soft memory cap (e.g. `GOMEMLIMIT=2GiB`). Prevents OOM kills in Kubernetes containers by automatically adjusting `GOGC` target pacing dynamically when memory usage nears the threshold.

---

## ⚠️ Common Gotchas
1. **Memory Ballast Anti-Pattern**: Setting up dummy huge byte arrays (`make([]byte, 10<<30)`) was common in legacy Go to delay GC cycles, but is now **obsolete** due to `GOMEMLIMIT` in Go 1.19+.
