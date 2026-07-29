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
