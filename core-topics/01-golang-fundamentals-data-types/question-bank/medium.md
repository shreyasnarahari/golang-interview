# Topic 01 Medium Questions

### Q: Why does string concatenation `s += "text"` in a loop cause poor performance?
**Frequency**: sometimes
**Why interviewers ask**: It checks whether you connect string immutability to allocation behavior and know the idiomatic fix.
**Answer**: Strings are immutable in Go, so every `+=` allocates a new backing array and copies all existing bytes plus the new ones — building an N-byte string this way copies O(N²) bytes in total. The fix is `strings.Builder` (or `bytes.Buffer`), which grows one buffer in amortized O(N):

```go
var b strings.Builder
for _, part := range parts {
	b.WriteString(part)
}
s := b.String() // zero-copy view of the accumulated buffer
```

Note the compiler does optimize one special case: a single expression like `a + b + c` allocates only once.
**Follow-ups**: Why is `Builder.String()` zero-copy while `bytes.Buffer.String()` historically copied? Does the answer change if the parts are known up front? (Yes — `strings.Join` precomputes the total length and allocates once.)

### Q: What do `len("你好")` and `utf8.RuneCountInString("你好")` return, and why do they differ?
**Frequency**: sometimes
**Why interviewers ask**: It probes whether you internalized that Go strings are UTF-8 byte sequences, not character arrays.
**Answer**: `len("你好")` returns **6** — the number of bytes (each CJK character takes 3 UTF-8 bytes) — while `utf8.RuneCountInString` returns **2**, the number of code points. A `for range` loop over a string decodes UTF-8 on the fly: the index advances by byte offset and the value is a `rune`. Slicing by bytes (`s[:2]`) can split a multi-byte rune and produce invalid UTF-8, which decoders render as the Unicode replacement character (U+FFFD).
**Follow-ups**: How would you reverse a string containing multi-byte characters without corrupting it? (Convert to `[]rune`, reverse, convert back.) What is `[]rune(s)` of an invalid UTF-8 string?

### Q: What are untyped constants in Go, and why does `var f float32 = 3.5` compile while `var i int = 3.5` does not?
**Frequency**: sometimes
**Why interviewers ask**: Untyped constants are a uniquely Go concept that explains a lot of "why does this compile?" confusion in everyday code.
**Answer**: A constant declared without an explicit type (like `const x = 3.5` or the literal `42`) is *untyped*: it carries a kind (integer, float, rune, string, bool) and arbitrary precision, but no concrete type. It only acquires a type at the point of use — it is converted to whatever type the context demands, if representable. So `3.5` can become a `float32` but not an `int`. When context forces a type without giving one (e.g. passing `42` to `fmt.Println`), the *default type* kicks in: `int`, `float64`, `rune`, `string`, or `bool`.

```go
const x = 3.5        // untyped float constant
var f32 float32 = x  // OK: representable as float32
// var i int = x     // compile error: constant 3.5 truncated to integer
```

**Follow-ups**: What type does `fmt.Println(42)` see — and what about `fmt.Println('a')`? (Default types: `int` and `rune`/`int32`, so it prints `97`.) Why is this mechanism preferable to C-style implicit conversions?

### Q: What happens when you convert between `string` and `[]byte`?
**Frequency**: sometimes
**Why interviewers ask**: Conversion costs are a classic hidden-allocation bug in Go hot paths; interviewers want candidates who know when the copy happens.
**Answer**: Each direction allocates and copies: `[]byte(s)` copies the string's bytes into a new mutable slice, and `string(b)` copies the slice's bytes into a new immutable string. Neither is a "cast" — you get an independent copy, so mutating the slice afterward never affects the string. Two consequences: converting a large byte slice to string just to compare or store it is a real cost, and the copy is exactly what guarantees string immutability. If you only need a read-only view, prefer keeping `[]byte` and using `bytes.Equal` etc. (The compiler does elide the copy in a few provably safe spots, e.g. a `m[string(b)]` map *lookup*.)
**Follow-ups**: Why must `string(b)` copy at all — what invariant would break otherwise? (String immutability: the slice could be mutated later.) How would you do a zero-copy conversion, and why is it dangerous? (`unsafe.String`/`unsafe.Slice` since Go 1.20 — you take ownership of the immutability contract yourself.)
