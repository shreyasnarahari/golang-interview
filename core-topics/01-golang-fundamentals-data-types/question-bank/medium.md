# Topic 01 Medium Questions

### Q: Why does string concatenation `s += "text"` in a loop cause poor performance?
**Answer**: Because strings are immutable in Go, every `+=` operation allocates a new backing memory array and copies the old and new bytes into it, resulting in $O(N^2)$ time complexity.
**Fix**: Use `strings.Builder` or `bytes.Buffer` to efficiently accumulate bytes in $O(N)$ amortized time.
