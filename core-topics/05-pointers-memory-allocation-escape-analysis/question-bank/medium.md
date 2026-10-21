# Topic 05 Medium Questions

### Q: Why does passing a local struct to `fmt.Println(x)` cause it to escape to the heap?
**Frequency**: often
**Why interviewers ask**: It tests whether you can connect the abstract idea of interface boxing to a concrete, everyday allocation.
**Answer**: `fmt.Println(a ...any)` takes variadic `any` (`interface{}`). Storing a concrete value in an interface requires a boxed copy of the value that the interface's data word points to, and because the compiler cannot prove that boxed copy stays local (the call is opaque and may retain it), the value is moved to the heap. Even plain constants show this: `go build -gcflags="-m"` reports `42 escapes to heap` for `fmt.Println(42)`. Hot paths should use typed logging (e.g. `slog.Int`) or avoid printing to keep values on the stack.
**Follow-ups**: Does this apply to small ints too, or only structs? How could you verify the escape yourself?

### Q: Is it safe to return a pointer to a local variable in Go? What happens?
**Frequency**: often
**Why interviewers ask**: C/C++ programmers expect a dangling-pointer bug; the question checks you know Go's answer and its cost.
**Answer**: It is completely safe — Go has no dangling pointers to locals. Escape analysis detects that the variable outlives the frame and silently allocates it on the heap, so the returned pointer stays valid. The cost is a heap allocation plus future GC work, which is why returning a value instead of a pointer is often faster for small structs.
```go
func mk() *Point {
	p := Point{1, 2}
	return &p // compiler reports: moved to heap: p
}
```
**Follow-ups**: If returning a pointer is safe, when would you choose to return a value instead? What does the `-gcflags="-m"` output look like for this function?

### Q: Does `make([]byte, n)` with a non-constant `n` always allocate on the heap?
**Frequency**: sometimes
**Why interviewers ask**: It probes whether your mental model is current — older Go always heap-allocated such slices, modern Go does not.
**Answer**: Not since Go 1.20. The compiler can stack-allocate a variable-size slice's backing array when the slice does not escape: it emits code that stack-allocates when the size is small enough and falls back to the heap at run time when it is not. You can see `make([]byte, n) does not escape` in `-gcflags="-m"` output. The escape decision still dominates — if the slice is returned or stored somewhere long-lived, it goes to the heap regardless of size.
**Follow-ups**: What is the compile-time-size cutoff above which even a local array always escapes? How would you prove the stack allocation happens?

### Q: When can passing a struct by pointer be *slower* than passing it by value?
**Frequency**: sometimes
**Why interviewers ask**: It defeats the C-derived instinct that "pointers are always cheaper" and checks real understanding of allocation costs.
**Answer**: When the pointer forces the struct to escape to the heap. Copying a small struct (a few words) within the stack is nearly free, while taking its address may push it through the heap allocator and make it GC-tracked. Pointers also add indirection (cache misses) and make the aliasing analysis harder for the compiler. Rule of thumb: pass small, immutable structs by value; use pointers for large structs, when mutation must be visible to the caller, or when the type carries a no-copy field like a mutex.
**Follow-ups**: How does this advice interact with method receiver choice? How would you measure which is faster for a specific struct?
