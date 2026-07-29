# Topic 09 Hard Questions

### Q: How does `errors.As` inspect custom error types through pointer reflection?
**Answer**: `errors.As(err, target)` requires `target` to be a non-nil pointer to a type that implements `error` or is an interface. It walks the `Unwrap()` chain using reflection until it finds a type assignable to `*target`, then populates `target`'s underlying pointer memory directly.
