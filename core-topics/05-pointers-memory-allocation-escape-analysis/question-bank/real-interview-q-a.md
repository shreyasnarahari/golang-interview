# Topic 05 Real Interview Q&A

### Q: Does `new(T)` always allocate on the heap?
**Answer**: No. `new(T)` simply allocates zeroed memory for type `T`. If the returned pointer does not escape the function scope, the Go compiler allocates `new(T)` directly on the stack.

### Q: How do you inspect escape analysis decisions made by the Go compiler?
**Answer**: By passing `-gcflags="-m"` to `go build` or `go test`.
