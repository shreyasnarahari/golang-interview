# Topic 01 Hard Questions

### Q: Explain the memory layout of string header (`reflect.StringHeader`) in 64-bit architecture.
**Answer**: In Go runtime, a string header is represented by a 16-byte struct:
```go
type StringHeader struct {
    Data uintptr // 8 bytes pointer to underlying byte array
    Len  int     // 8 bytes length of string in bytes
}
```
Because `StringHeader` contains no capacity field (`Cap`), string memory boundaries are strictly fixed.
