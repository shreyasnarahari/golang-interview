# Topic 13 Hard Questions

### Q: Explain the valid conversions allowed for `unsafe.Pointer`.
**Answer**:
1. A pointer of any type to `unsafe.Pointer`.
2. `unsafe.Pointer` to a pointer of any type.
3. `uintptr` to `unsafe.Pointer`.
4. `unsafe.Pointer` to `uintptr`.
Pointer arithmetic MUST be performed by converting `unsafe.Pointer` $\rightarrow$ `uintptr` $\rightarrow$ Arithmetic $\rightarrow$ `unsafe.Pointer` within a single expression.
