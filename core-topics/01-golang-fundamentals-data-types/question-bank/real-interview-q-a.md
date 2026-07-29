# Topic 01 Real Interview Questions & Answers

### Q1: Is Go pass-by-reference or pass-by-value?
**Answer**: Go is strictly **pass-by-value**. Every variable passed to a function is copied. When passing a pointer, a copy of the memory address is passed. When passing a slice, a copy of the 24-byte `SliceHeader` struct is passed.

### Q2: What is a `rune` in Go?
**Answer**: A `rune` is an alias for `int32`. It represents a single UTF-8 encoded Unicode code point.

### Q3: What happens when you index a string `str[0]`?
**Answer**: It returns the raw `byte` (`uint8`) at position 0, not a character or rune.
