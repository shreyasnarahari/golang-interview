# Topic 02 Hard Questions

### Q: What is full 3-index slicing `s[low:high:max]` and why is it useful?
**Answer**: Full 3-index slicing sets the capacity of the resulting slice explicitly to `max - low`. This prevents sub-slices from accidentally overwriting elements in the parent backing array during subsequent `append()` operations.

```go
s := []int{1, 2, 3, 4, 5}
sub := s[1:3:3] // len=2, cap=2. append(sub, 99) WILL force reallocation!
```
