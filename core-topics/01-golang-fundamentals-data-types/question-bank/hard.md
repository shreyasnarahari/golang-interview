# Topic 01 Hard Questions

### Q: Describe the in-memory layout of a Go string on a 64-bit architecture.
**Frequency**: rare-senior
**Why interviewers ask**: Understanding the string header explains conversion costs, GC behavior, and why strings are cheap to pass around.
**Answer**: A string value is a 16-byte header — a pointer to an immutable byte array (8 bytes) plus a length (8 bytes) — with no capacity field, since strings are immutable and never grow. This is why passing a string copies only the header, and why `s[:5]` is O(1) (a new header sharing the same data). `reflect.StringHeader` exposes this shape but is deprecated as of Go 1.20: its `Data uintptr` field does not keep the backing array alive for the GC. The modern, safe replacements are `unsafe.String(ptr, len)` / `unsafe.StringData(s)` / `unsafe.Slice` / `unsafe.SliceData`, which keep the value typed as a pointer through conversion.

```go
// conceptual runtime layout (not buildable as-is; uintptr is not GC-safe here)
type stringHeader struct {
	data *byte // 8 bytes: pointer to backing array
	len  int   // 8 bytes: length in bytes
}
```

**Follow-ups**: Why can `unsafe.String(&b[0], len(b))` be a data-corruption bug even if the lengths are right? (The returned string aliases the slice; mutating the slice later violates string immutability, breaking map keys, string equality caches, etc.) Given the header layout, why is `len(s)` O(1) but counting runes O(N)?

### Q: Why does `const big = 1 << 100` compile, and when would using `big` fail?
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates who truly understand Go's constant model — arbitrary precision at compile time — from those who think constants are just immutable variables.
**Answer**: Untyped constants are evaluated by the compiler in *arbitrary* precision (the spec mandates at least 256 bits of precision for numeric constants), so `1 << 100` is a perfectly valid untyped integer constant even though no machine integer holds it. It only fails when you force it into a concrete type that cannot represent it: `var i int = 1 << 100` is a compile error ("constant ... overflows int"), while `var f float64 = 1 << 100` compiles (it rounds to the nearest float64). A fun corollary: constant float math is exact, so `const c = 0.1 + 0.2; c == 0.3` is **true**, while the same comparison on runtime `float64` variables is **false** (0.30000000000000004). This also makes `len("abc")` a constant usable in array lengths: `var a [len("abc")]int`.
**Follow-ups**: What is the type of `big` if you write `fmt.Println(big)` and it overflows `int`? (Compile error — the default type `int` cannot hold it; you must convert explicitly.) How would you print `big` exactly? (Convert to `*big.Int` via `new(big.Int).Lsh(big.NewInt(1), 100)` — you cannot name the constant directly into `big.Int`, so the API is used instead.)

### Q: Why does the map lookup `m[string(b)]` not allocate, while `m[string(b)] = v` does?
**Frequency**: rare-senior
**Why interviewers ask**: It tests knowledge of a famous compiler optimization that every performance-sensitive Go engineer leans on when parsing bytes.
**Answer**: Normally `string(b)` allocates and copies. But the compiler recognizes a few conversion patterns where the result cannot escape — most notably a map **lookup** `v := m[string(b)]` (and comma-ok / delete / key-comparison forms) — and rewrites them to hash and compare the bytes directly, with zero allocations (verifiable via `go test -bench -benchmem`: 0 allocs/op for lookups). An **assignment** `m[string(b)] = v` must store the key, so it genuinely materializes a string and allocates (1 alloc/op). Since Go 1.20 you can also hand-roll the zero-copy form with `unsafe.String(&b[0], len(b))`, but then you own the immutability contract: if `b` is later mutated or freed, the map's keys corrupt silently.
**Follow-ups**: How would you prove the allocation difference to a skeptic? (Benchmark with `-benchmem`, or `go build -gcflags=-m` escape analysis.) What other conversion patterns does the compiler optimize? (e.g. ranging over `string(b)` without using the bytes slice, and `string(b) == "..."`-style comparisons in some versions.)

### Q: Go is pass-by-value, so why can a function mutate the *contents* of a slice argument but not reliably change its *length*?
**Frequency**: sometimes
**Why interviewers ask**: It is the single best question for exposing whether a candidate knows what a slice actually is — a header, not the array.
**Answer**: The function receives a **copy of the 24-byte slice header** (pointer, len, cap). Because the pointer field is copied verbatim, both headers alias the same backing array, so `s[0] = x` inside the callee is visible to the caller. But `len` and `cap` are also copies: `s = append(s, x)` in the callee only updates *its* header — the caller still sees the old length. If append grows beyond capacity it allocates a new array, and even element writes stop being visible. Maps and channels work similarly: the copied value contains a pointer to a shared runtime structure (`hmap`), so inserts through a map parameter are visible. The rule of thumb: the *header* is passed by value; the *backing data* is shared by pointer.
**Follow-ups**: Given that, when should a function take `*[]int` instead of `[]int`? (Almost never — return the new slice; `*[]int` is only for rare in-place reslicing APIs.) Does the same aliasing apply to arrays? (No — a `[N]T` array is copied wholesale, including its elements, which is why large arrays are expensive to pass.)
