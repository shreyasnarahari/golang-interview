# Topic 01 Easy Questions

### Q: How do you declare constants in Go using `iota`?
**Answer**: `iota` is an integer constant generator that increments by 1 for each line in a `const` block.

```go
const (
	Low = iota // 0
	Medium     // 1
	High       // 2
)
```
