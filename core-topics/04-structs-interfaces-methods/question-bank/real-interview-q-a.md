# Topic 04 Real Interview Q&A

### Q: What is compile-time interface compliance check in Go?
**Answer**: A pattern used to verify at compile time that a struct type implements an interface:
```go
var _ MyInterface = (*MyStruct)(nil)
```
If `*MyStruct` fails to implement `MyInterface`, compilation halts with a clear error.

### Q: What is struct embedding vs inheritance?
**Answer**: Struct embedding promotes embedded fields and methods to the outer struct for convenience, but it is composition, not subtyping/inheritance.
