# Topic 02 Easy Questions

### Q: What is the zero value of a slice and a map?
**Answer**: `nil`. A `nil` slice has `Len=0, Cap=0, Data=nil`. Appending to a `nil` slice works and allocates memory. However, writing to a `nil` map (`m["key"] = 1`) panics (`panic: assignment to entry in nil map`).
