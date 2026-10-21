# Topic 02 Hard Questions

### Q: What is full 3-index slicing `s[low:high:max]` and why is it useful?
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates who have been bitten by backing-array aliasing from those who only know the happy path.
**Answer**: The third index sets the capacity of the result explicitly to `max - low`, bounding how far the sub-slice can grow in place. This prevents a later `append` on the sub-slice from silently overwriting elements of the parent backing array — a nasty aliasing bug when handing out sub-slices from a shared buffer.

```go
s := []int{1, 2, 3, 4, 5}
sub := s[1:3:3] // len=2, cap=2. append(sub, 99) forces reallocation,
                // so s[3] is NOT clobbered (a plain s[1:3] would allow it)
```
Typical uses: parsers/tokenizers handing out slices of a shared input, and defensive APIs that must guarantee callers can't mutate the original buffer via append.
**Follow-ups**: What is the runtime cost of the third index? How does this compare with `slices.Clone` as an aliasing fix?

### Q: Describe exactly how `append` grows a slice — what is the growth formula in current Go?
**Frequency**: rare-senior
**Why interviewers ask**: "It doubles" is the Go 1.17-era answer; the runtime switched to a smooth formula in Go 1.18, and seniors should know the real behavior and its consequences.
**Answer**: Since Go 1.18 (`runtime.nextslicecap`): if the requested length already exceeds double the old capacity, the new capacity is the requested length. Otherwise capacity **doubles while below 256 elements**, then transitions smoothly toward **1.25x** growth via `newcap += (newcap + 3*256) >> 2` until it fits. The result is then rounded up to the next memory size class, so observed capacities aren't the "clean" 1,2,4,8 sequence — e.g. a 256→512→848→1280→1792→2560 progression for `[]int`. Two practical corollaries: (1) growth is still amortized O(1) with bounded waste, but the waste factor shrinks for large slices; (2) never hard-code expected capacities — preallocate with `make([]T, 0, n)` when the final size is known to skip all intermediate copies. The compiler may also pre-grow slices when it can prove consecutive appends, so microbenchmarks can show even larger first capacities.
**Follow-ups**: Why does growth need to be superlinear at all (amortized analysis)? What happens to the old backing array, and when does the GC reclaim it?

### Q: Explain the pre-Go 1.24 map implementation: hmap, bmap, and how collisions and growth worked.
**Frequency**: rare-senior
**Why interviewers ask**: This was the canonical Go-internals question for a decade, and you still need it to read older literature, profiles, and to appreciate why Go 1.24 changed it.
**Answer**: A map was a pointer to `hmap` (count, bucket-array pointer, hash seed, growth state). The bucket array held `bmap` buckets of exactly **8 key/value slots** each, with an 8-byte `tophash` array caching the top byte of each key's hash for fast miss rejection. Collisions overflowed into chained **overflow buckets**. Keys were stored together, then values (`k0k1...k7 v0v1...v7`), avoiding padding waste. Growth triggered when the **load factor exceeded 6.5** (or too many overflow buckets): the bucket count doubled, but rehashing was **incremental** — `oldbuckets` entries were evacuated a few at a time on subsequent writes/reads (`growing` flag, `evacuate`), so no single insert paid a huge latency spike. Deletion cleared slots but never shrank the bucket array.
**Follow-ups**: Why evacuate incrementally instead of all at once — what latency guarantee does it buy? Why store keys and values in separate arrays within a bucket rather than interleaved?

### Q: How did Go 1.24's Swiss-table map redesign change all of this?
**Frequency**: rare-senior
**Why interviewers ask**: Modern Go teams expect you to know the current implementation — it changes performance reasoning, memory footprints, and which "map facts" are still true.
**Answer**: Go 1.24 made Swiss tables (`internal/runtime/maps`, formerly `GOEXPERIMENT=swissmap`) the default, replacing hmap/bmap. Key mechanics, verified against go1.27.2 sources: storage is **groups of 8 slots**; each group has an 8-byte **control word** where each byte marks empty/deleted or holds **H2 — the low 7 bits** of the slot's hash. Lookup compares all 8 control bytes **in parallel** with bitwise tricks (SIMD-accelerated on amd64), so one probe effectively checks 8 slots at once; only on an H2 match (~0.7% false-positive rate) is the full key compared. The upper 57 bits (H1) pick the starting group, and collisions resolve by **quadratic probing across groups** — no overflow buckets. The load-factor ceiling rose from 6.5 to **7/8 (87.5%)**. Deleting from a full group leaves a **tombstone** (`ctrlDeleted`) so probe chains stay intact; inserts reuse tombstones first. Incremental growth is preserved via **extendible hashing**: a directory of tables, each covering part of the hash space, so only one small table rehashes at a time (up to `maxTableCapacity`). Net effects vs. pre-1.24: ~30% faster lookups/inserts on large maps, less memory overhead at high load, and no long overflow chains.
**Follow-ups**: Why can a Swiss table never be 100% full (probe-sequence invariant)? Which pre-1.24 facts remain true — iteration randomization, concurrent-access fatal errors, no shrink on delete?
