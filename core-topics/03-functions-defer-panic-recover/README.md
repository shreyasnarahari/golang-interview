# Topic 03: Functions, Defer, Panic, and Recover

## 📌 Overview
Functions in Go support multiple return values, named returns, variadic parameters, and closures. Go uses `defer`, `panic`, and `recover` for resource cleanup and exception-like unwinding.

---

## 🎯 Key Concepts

### 1. Named Return Values & Defer Modification
- Named return variables are initialized to their zero values at function entry.
- A deferred closure can inspect and modify named return values before the function yields control to the caller:

```go
func triple(x int) (result int) {
    defer func() {
        result += x // Accesses named return variable 'result'!
    }()
    return x * 2 // result = 2 * x, then defer runs -> result = 3 * x
}
```

### 2. Defer Mechanics & Performance Evolution
- `defer` calls execute in **LIFO** (Last-In, First-Out) order.
- **Argument Evaluation**: Arguments to a deferred function are evaluated **immediately** when the `defer` statement is executed, NOT when the deferred call runs!
- **Compiler Optimization Evolution**:
  - *Go 1.13 and earlier*: Defer calls allocated `_defer` structs on heap (~35ns per defer).
  - *Go 1.14+*: **Open-coded defers**. Inlines defers directly into the function return path for loops with fixed iterations, reducing defer overhead to near **zero extra cost** (~1–3ns).

### 3. Panic & Recover Constraints
- `panic` stops normal execution, unwinds the stack of the current goroutine, and runs all deferred functions.
- `recover()` intercepts a panicking goroutine and restores normal execution.
- **Rule**: `recover()` ONLY works when called **directly inside a deferred function**. Calling `recover()` in a helper function invoked by `defer` returns `nil` and fails to stop the panic.

---

## ⚠️ Common Gotchas
1. **Goroutine Boundary**: A `recover()` in one goroutine CANNOT catch a panic originating in a different goroutine. Every goroutine must handle its own panics!
2. **Defer in Infinite Loop**: Calling `defer file.Close()` inside a `for` loop defers cleanup until the entire enclosing function returns, potentially exhausting file descriptors.
