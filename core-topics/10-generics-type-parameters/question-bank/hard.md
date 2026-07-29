# Topic 10 Hard Questions

### Q: Explain the GCShape stencil compilation model used in Go generics.
**Answer**: GCShape (Garbage Collector Shape) groups generic types that share identical memory layout and GC pointer maps (e.g. all pointer types `*int`, `*string`). The compiler generates a single shared function body for all types sharing the same GCShape, passing a runtime dictionary for type-specific method calls, reducing binary code duplication.
