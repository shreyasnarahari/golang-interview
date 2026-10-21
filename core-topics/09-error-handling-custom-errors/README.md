# Topic 09: Error Handling & Custom Errors

## 📌 Overview
Go treats errors as explicit values rather than control-flow exceptions. Go 1.13 introduced formal error wrapping, unwrapping, and inspection tools (`errors.Is` and `errors.As`).

---

## 🎯 Key Concepts

### 1. The `error` Interface
In Go, an error is any type implementing the single-method `error` interface:
```go
type error interface {
    Error() string
}
```

### 2. Error Wrapping & Unwrapping (Go 1.13+)
- **Wrapping (`fmt.Errorf("... %w", err)`)**: Uses the `%w` verb to wrap an underlying target error inside a new error instance.
- **`errors.Is(err, target)`**: Reports whether any error in `err`'s chain matches `target`. Replaces direct equality comparisons (`err == ErrNotFound`).
- **`errors.As(err, &targetStruct)`**: Searches `err`'s chain for an error that can be assigned to `targetStruct`. Replaces type assertions (`err.(*MyError)`).

```
Error Chain Wrapping
+-------------------------------------------------------+
|  fmt.Errorf("database query failed: %w", errNotFound) |
|      |                                                |
|      +---> Wrapped target: ErrNotFound                |
+-------------------------------------------------------+
```

### 3. Sentinel Errors vs Custom Error Structs
- **Sentinel Errors**: Package-level exported constants/variables (e.g. `io.EOF`, `sql.ErrNoRows`).
- **Custom Error Structs**: Structs holding extra diagnostic context (e.g. HTTP status codes, SQL query text).

---

## ⚠️ Common Gotchas
1. **Using `%v` instead of `%w`**: `fmt.Errorf("err: %v", err)` formats the error message as a string but breaks the unwrap chain! `errors.Is` will fail to match wrapped errors.
2. **Wrapping `nil` errors**: `fmt.Errorf("op: %w", nil)` returns a **non-nil** error wrapping nil. Always check `err != nil` before wrapping.
3. **Comparing wrapped errors with `==`**: After wrapping, `err == io.EOF` fails even if the chain contains `io.EOF` — use `errors.Is`/`errors.As`.
4. **Multi-wrapping (Go 1.20+)**: `fmt.Errorf` accepts multiple `%w` verbs, producing an error whose `Unwrap() []error` exposes a **tree**; `errors.Is`/`As` traverse the whole tree. `errors.Join(err1, err2)` builds one directly.
