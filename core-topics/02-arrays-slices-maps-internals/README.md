# Topic 02: Arrays, Slices, and Maps Internals

## 📌 Overview
Understanding arrays, slices, and maps at the runtime data structure level is one of the most heavily tested areas in Go technical interviews.

---

## 🎯 Key Concepts

### 1. Slice Header (`reflect.SliceHeader`)
A slice in Go does NOT store elements directly. It is a 24-byte header pointing to a contiguous backing array:

```go
type SliceHeader struct {
    Data uintptr // Pointer to backing array memory
    Len  int     // Number of elements in the slice
    Cap  int     // Maximum number of elements before reallocation
}
```

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

---

## ⚠️ Common Gotchas
1. **Sharing Backing Array**: Slicing an existing slice `s2 := s1[1:3]` shares the underlying array. Mutating `s2` mutates `s1`!
2. **Concurrent Map Panic**: Concurrent read/write to a `map` panics runtime: `fatal error: concurrent map read and map write`.
