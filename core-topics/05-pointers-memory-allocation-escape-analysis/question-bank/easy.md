# Topic 05 Easy Questions

### Q: What is the primary performance benefit of stack allocation over heap allocation?
**Frequency**: often
**Why interviewers ask**: It reveals whether you understand why escape analysis exists and what heap allocation actually costs.
**Answer**: Stack allocation is just a stack-pointer bump and cleanup is free — the frame is reclaimed when the function returns. Heap allocation goes through the runtime's allocator (`mcache`/`mcentral`/`mheap`) and the object must later be found and freed by the garbage collector, adding both allocation latency and GC pressure.
**Follow-ups**: When must a variable go to the heap instead? How would you check where a specific variable was allocated?

### Q: What is the difference between `new` and `make` in Go?
**Frequency**: often
**Why interviewers ask**: Confusing the two is a classic beginner trap; the answer shows whether you know Go's built-in reference types.
**Answer**: `new(T)` allocates zeroed memory for any type `T` and returns a `*T`; it does no initialization beyond zeroing. `make` works only on slices, maps, and channels, and returns an initialized (ready-to-use) value of that type — e.g. a slice with a backing array or a map with buckets.
```go
p := new(int)        // *int pointing at a zeroed int
s := make([]int, 10) // []int with len 10 and a live backing array
```
A `new([]int)` result is a pointer to a nil slice — you can't index it without further initialization.
**Follow-ups**: Does `new(T)` always allocate on the heap? Why does `make` not exist for structs?

### Q: What is escape analysis?
**Frequency**: often
**Why interviewers ask**: It is the central concept of this topic and separates candidates who have read Go docs from those who have profiled real code.
**Answer**: A compile-time analysis the Go compiler performs to decide whether a variable can live on the stack or must "escape" to the heap. If the compiler can prove a reference never outlives its function's stack frame, it stack-allocates it; otherwise it heap-allocates it. You inspect the decisions with `go build -gcflags="-m"`.
**Follow-ups**: Name two constructs that force an escape. Is escape analysis done at run time or compile time?

### Q: What is the zero value of a pointer, and what happens when you dereference it?
**Frequency**: often
**Why interviewers ask**: Nil-pointer panics are the most common crash in Go services; interviewers want to know you recognize and guard against them.
**Answer**: The zero value of any pointer type is `nil`. Dereferencing a nil pointer (reading or writing through it) causes a run-time panic: `invalid memory address or nil pointer dereference`. Calling a *method* with a nil pointer receiver is fine as long as the method doesn't dereference it — a common idiom is to handle nil receivers explicitly.
**Follow-ups**: Is a nil pointer the same as a pointer to a zero value? How does a nil pointer inside a non-nil interface cause a bug?
