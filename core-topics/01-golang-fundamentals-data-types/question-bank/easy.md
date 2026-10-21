# Topic 01 Easy Questions

### Q: How do you declare constants in Go using `iota`?
**Frequency**: often
**Why interviewers ask**: It tests whether you can read and write idiomatic Go enums, which appear in almost every codebase.
**Answer**: `iota` is a constant generator that equals the index of its `ConstSpec` line within a `const` block, starting at 0. Omitting the expression repeats the previous one, and blank lines do not consume an index.

```go
const (
	Low = iota // 0
	Medium     // 1 (repeats the iota expression)
	High       // 2
)
```

`iota` resets to 0 in each new `const` block.
**Follow-ups**: How would you use `iota` to declare bit-flag constants like `1 << iota`? Does a blank line inside the block increment `iota`? (No — it counts constant specifications, not physical lines.)

### Q: What are the zero values of Go's basic types?
**Frequency**: often
**Why interviewers ask**: Zero values are a core Go design principle; interviewers check you rely on them instead of defensive nil/zero initialization.
**Answer**: Every declared-but-uninitialized variable gets a well-defined zero value: `0` for all numeric types, `false` for `bool`, `""` for `string`, and `nil` for pointers, slices, maps, channels, functions, and interfaces. Because of this, `var s strings.Builder` or `var mu sync.Mutex` is immediately usable without a constructor.
**Follow-ups**: Which of these types can be used right after `var` declaration without initialization — a slice, a map, a `sync.Mutex`? (Slice: yes for reads/append; map: no, writes panic; Mutex: yes.)

### Q: What is the difference between `byte` and `rune`?
**Frequency**: often
**Why interviewers ask**: It reveals whether you understand how Go represents text — bytes versus Unicode code points — which trips up candidates on string handling.
**Answer**: `byte` is an alias for `uint8` (a raw unit of memory/UTF-8 data); `rune` is an alias for `int32` (a single Unicode code point). A rune literal uses single quotes (`'世'`); its value is the code point number. One rune may occupy 1–4 bytes when UTF-8 encoded.
**Follow-ups**: What does `"héllo"[1]` return — a byte or a rune? (A byte: indexing a string yields `uint8`.) How do you count characters instead of bytes in a string? (`utf8.RuneCountInString` or `len([]rune(s))`.)

### Q: What integer types does Go have, and which should you use by default?
**Frequency**: often
**Why interviewers ask**: Choosing `int` versus sized types signals whether you think about portability and when explicit sizing actually matters.
**Answer**: Go offers `int8/16/32/64`, their unsigned counterparts, plus `int` and `uint` whose size matches the platform's word size (32 or 64 bits), and `uintptr` for pointer arithmetic with `unsafe`. Use plain `int` unless you have a specific reason: wire formats and binary protocols (explicit size), memory-sensitive huge slices, or bit manipulation requiring a known width. Never use sized ints just to "save memory" in ordinary code — and never assume `int` is 64-bit.
**Follow-ups**: What happens when integer arithmetic overflows in Go? (Both signed and unsigned overflow wrap around deterministically modulo 2^n per the spec — there is no panic.) How do you detect the platform's int size at runtime? (`strconv.IntSize` or `unsafe.Sizeof(int(0))`.)
