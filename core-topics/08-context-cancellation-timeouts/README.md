# Topic 08: Context, Cancellation, & Deadlines

## 📌 Overview
The `context.Context` interface is the standard Go idiom for carrying deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines.

---

## 🎯 Key Concepts

### 1. `context.Context` Interface
```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

### 2. Derived Context Factory Functions
1. **`context.Background()` & `context.TODO()`**: Non-nil empty root contexts that are never canceled.
2. **`context.WithCancel(parent)`**: Returns a child context and a `CancelFunc`. Calling `cancel()` closes the `Done()` channel for the child and all its descendants.
3. **`context.WithTimeout(parent, duration)`**: Automatically triggers cancellation when the specified duration elapses.
4. **`context.WithDeadline(parent, time)`**: Automatically triggers cancellation when a specific clock deadline is reached.
5. **`context.WithValue(parent, key, val)`**: Returns a child context carrying a request-scoped key-value pair.

### 3. Modern Context APIs (Go 1.20 / 1.21)
- **`context.WithCancelCause(parent)`** (1.20): like `WithCancel`, but `cancel(err)` records a *cause*. Children see `ctx.Err() == context.Canceled`, while the canceler can retrieve the real reason via **`context.Cause(ctx)`**. `WithDeadlineCause`/`WithTimeoutCause` do the same for deadlines.
- **`context.WithoutCancel(parent)`** (1.21): returns a context that inherits the parent's *values* but is never canceled by it — essential for async work (e.g., logging, cleanup) that must outlive the request.
- **`context.AfterFunc(ctx, f)`** (1.21): arranges for `f` to run in its own goroutine after `ctx` is done; the returned stop function unregisters it. This replaces manual `go func() { <-ctx.Done(); f() }()` plumbing and is how 1.21+ context propagation itself is implemented.

```
Context Tree Hierarchy
+-------------------------------+
|     context.Background()      |
+-------------------------------+
                |
+-------------------------------+
| WithValue("requestID", "123") |
+-------------------------------+
                |
+-------------------------------+
| WithTimeout(parent, 500ms)    |  ===> Triggers Done() on timeout or explicit cancel()
+-------------------------------+
```

---

## ⚠️ Common Gotchas
1. **Always Defer `cancel()`**: Failing to call `cancel()` when using `WithTimeout` or `WithCancel` causes the parent context to retain references to the child, causing memory leaks!
2. **Context Value Overuse**: `WithValue` should ONLY be used for request-scoped metadata (trace IDs, auth tokens), NEVER for passing optional function arguments!
