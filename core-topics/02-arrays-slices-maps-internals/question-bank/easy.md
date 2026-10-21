# Topic 02 Easy Questions

### Q: What is the zero value of a slice and a map, and what operations are safe on them?
**Frequency**: often
**Why interviewers ask**: It immediately exposes whether you understand nil semantics, which cause some of the most common Go panics in production.
**Answer**: The zero value of both is `nil`. A `nil` slice has `Len=0, Cap=0, Data=nil`; appending to it works and allocates a backing array. A `nil` map is read-only: `v := m[k]`, `len(m)`, `delete(m, k)`, and `for range` are all safe (return zero value / 0 / no-op), but writing `m[k] = v` panics with `assignment to entry in nil map`. Always create maps with `make(map[K]V)` or a literal before writing.
**Follow-ups**: How do you distinguish "key absent" from "key present with zero value"? What does `var m map[K]V` vs `m := map[K]V{}` differ in?

### Q: What does a slice actually consist of under the hood?
**Frequency**: often
**Why interviewers ask**: The slice header explains almost every "surprising" slice behavior — aliasing, append copies, and cheap slicing.
**Answer**: A slice is a 3-word descriptor (24 bytes on 64-bit): a pointer to the backing array (`Data`), the length (`Len`), and the capacity (`Cap`). Slicing (`s[1:3]`) just builds a new header pointing into the same array — O(1), no copy. Two slices can share a backing array, which is why appending through one can be invisible to the other. Note `reflect.SliceHeader` is deprecated; use `unsafe.SliceData`, `unsafe.Len`/`Cap`-style APIs if you ever need the parts.
**Follow-ups**: If slicing is O(1), why can holding a small sub-slice of a huge slice be a memory leak? How does passing a slice to a function differ from passing an array?

### Q: What is the difference between `len` and `cap` of a slice?
**Frequency**: often
**Why interviewers ask**: Confusing the two leads to index-out-of-range panics and incorrect assumptions about when `append` reallocates.
**Answer**: `len` is the number of elements the slice currently exposes (indexes `0..len-1`); `cap` is the number of elements from the start of the slice to the end of the backing array — how far it can grow in place before `append` must reallocate. You can index only up to `len`, but you can reslice up to `cap` (`s = s[:cap(s)]`).

```go
s := make([]int, 3, 8) // len=3, cap=8
s = append(s, 1, 2)    // len=5, cap=8 — no reallocation
```
**Follow-ups**: What are len/cap after `s := make([]int, 0, 100)` and then one append? Can `cap` ever be less than `len`?

### Q: What happens when you read a key that doesn't exist in a map?
**Frequency**: often
**Why interviewers ask**: The comma-ok idiom is everywhere in Go code, and not knowing it produces subtle "zero value treated as real" bugs.
**Answer**: You get the zero value for the value type — no panic, no error. To tell "absent" apart from "present with zero value", use the two-result form:

```go
v, ok := m["key"]
if !ok {
	// key not present; v is the zero value
}
```
This is the comma-ok idiom; the same pattern appears in channel receives and type assertions.
**Follow-ups**: What does `v, ok := <-ch` report on a closed channel? Why can't you rely on `v == zeroValue` alone to detect absence?
