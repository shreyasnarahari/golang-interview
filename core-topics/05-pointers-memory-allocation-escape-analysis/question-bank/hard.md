# Topic 05 Hard Questions

### Q: What is stack growth in Go and how does it relate to escape analysis?
**Answer**: Every goroutine starts with a small **2 KB** stack. If a variable is allocated on the stack but its size is dynamically computed or exceeds frame boundaries, or if stack copying would invalidate references, the compiler escapes it to the heap.
