# Topic 13 Real Interview Q&A

### Q: Why is `uintptr` dangerous to store as a long-lived local variable?
**Answer**: `uintptr` is an integer representation of an address, not a pointer. The Go Garbage Collector does NOT treat `uintptr` as an active reference. If the object has no other pointers, GC will collect it, making the `uintptr` address dangling!

### Q: Why is CGO significantly slower than pure Go function calls?
**Answer**: CGO requires switching stack environments (Go segmented stack to C OS stack), disabling GC preemption, and executing calling conventions, taking ~100ns vs ~1ns for pure Go calls.
