# Topic 02: Arrays, Slices, and Maps Internals

## 📌 Overview
Understanding arrays, slices, and maps at the runtime data structure level is one of the most heavily tested areas in Go technical interviews.

---

## 🎯 Key Concepts

### 1. Slice Header (Data / Len / Cap)
A slice in Go does NOT store elements directly. It is a 24-byte header pointing to a contiguous backing array:

```go
type SliceHeader struct {
    Data uintptr // Pointer to backing array memory
    Len  int     // Number of elements in the slice
    Cap  int     // Maximum number of elements before reallocation
}
```

> **Modern Go note**: `reflect.SliceHeader` and `reflect.StringHeader` are **deprecated since Go 1.20**. Use the safe `unsafe` package functions instead: `unsafe.SliceData(s)` (get the data pointer), `unsafe.Slice(ptr, len)` (build a slice), and `unsafe.String`/`unsafe.StringData` for strings.

```
Slice Header (24 bytes)           Backing Array in Memory
+-----------------------+        +---+---+---+---+---+---+
| Data | Len=3 | Cap=5  | ------>| 1 | 2 | 3 | 4 | 5 |   |
+-----------------------+        +---+---+---+---+---+---+
```

### 2. Capacity Growth Algorithm (`growslice`)
When `append()` exceeds `Cap`, Go allocates a new larger backing array and copies elements over:
- In Go 1.18+, capacity growth is smooth:
  - If `cap < 256`: doubles capacity ($\times 2$).
  - If `cap >= 256`: grows by $(cap + 3 \times 256) / 4$ (transitioning smoothly to $\approx 1.25\times$ growth).

### 3. Map Internal Structure (`hmap` & `bmap`)
Go maps use a hash table consisting of an array of **buckets** (`bmap`):
- Each bucket holds up to **8 key-value pairs**.
- Load factor threshold: **6.5**. When exceeded, the map allocates $2\times$ buckets and begins incremental evacuation during subsequent writes.
- **Iteration Randomization**: Map iteration order is explicitly randomized by starting at a random bucket offset on every `for range` loop.

> **Go 1.24+ note**: The runtime replaced this classic bucket design with **Swiss Tables** (`internal/runtime/maps`): maps are now built from groups of 8 slots with a per-group control word of metadata bytes (1-byte hash fingerprints) checked with SIMD-friendly bit operations, plus a directory for large maps. The observable contract (randomized iteration, `fatal error: concurrent map read and map write`, O(1) average lookup) is unchanged, and the classic `hmap`/`bmap` model above is still common interview material — but know that it describes pre-1.24 Go.

---

## 🧰 Modern Standard Library Helpers (Go 1.21+)
Interviewers increasingly expect you to know the generic stdlib helpers that replaced hand-written loops:

**`slices` package**
- `slices.Sort(s)` — in-place sort for any `cmp.Ordered` slice (uses pdqsort; `slices.SortFunc` for custom comparators).
- `slices.Contains(s, v)` — linear membership test; also `slices.Index`, `slices.IsSorted`.
- `slices.Compact(s)` — removes **consecutive** duplicates in-place and returns the shortened slice (typically used after `Sort`); the tail of the backing array is zeroed out.
- `slices.Clone(s)` — allocates a **new backing array** and copies. Unlike `copy(dst, src)` — which writes into `dst`'s existing array (shared with whatever backs `dst`) — a clone never aliases the original, so appending to or mutating the clone cannot leak into the source.

**`maps` package**
- `maps.Equal(m1, m2)` — same length and every key maps to an `==`-equal value (`maps.EqualFunc` for custom equality).
- `maps.Clone(m)` — shallow copy: new map, same keys/values (pointer/interface values still point at shared data).
- **Go 1.23+**: `maps.Keys(m)` / `maps.Values(m)` return **iterators** (`iter.Seq[K]` / `iter.Seq[V]`), usable directly in `for k := range maps.Keys(m)`; collect with `slices.Collect` / `slices.Sorted`.

**Built-ins (Go 1.21)**
- `min` / `max` — variadic built-ins over ordered types; at least one argument required, type inferred from arguments.
- `clear(m)` — deletes **all** entries from a map (keeps allocated storage).
- `clear(s)` — zeroes **all elements** of a slice in place; length and capacity are unchanged.

---

## ⚠️ Common Gotchas
1. **Sharing Backing Array**: Slicing an existing slice `s2 := s1[1:3]` shares the underlying array. Mutating `s2` mutates `s1`!
2. **Concurrent Map Panic**: Concurrent read/write to a `map` panics runtime: `fatal error: concurrent map read and map write`.
