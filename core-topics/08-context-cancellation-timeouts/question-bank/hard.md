# Topic 08 Hard Questions

### Q: Explain how parent-child cancellation propagation works in `cancelCtx`.
**Answer**: When a `cancelCtx` is created via `WithCancel`, it attempts to locate a parent `cancelCtx` up the tree. If found, it registers itself in the parent's `children` map. When `parent.cancel()` is called, it locks its mutex, iterates over its `children` map, and recursively cancels every child goroutine context.
