# Topic 10: Generics & Type Parameters (Go 1.18+)

## 📌 Overview
Introduced in Go 1.18, Generics add type parameters to functions and types, enabling compile-time type-safe abstractions without runtime interface boxing costs or reflection.

---

## 🎯 Key Concepts

### 1. Generic Functions & Type Parameters
Type parameters are specified in square brackets `[T Constraint]` before function parameters:

```go
func Map[T any, R any](slice []T, transform func(T) R) []R {
    res := make([]R, len(slice))
    for i, v := range slice {
        res[i] = transform(v)
    }
    return res
}
```

### 2. Constraints & The `~` Operator
- **`any`**: Constraint permitting any Go type (alias for `interface{}`).
- **`comparable`**: Built-in constraint matching types supporting `==` and `!=` (primitives, pointers, structs with comparable fields).
- **Underlying Type Operator `~T`**: Matches any type whose **underlying type** is `T`.
  - e.g., `type MyInt int` satisfies `~int`, but DOES NOT satisfy `int`!

```go
type Number interface {
    ~int | ~int64 | ~float64 // Set element interface constraint
}
```

### 3. Compilation Model: Monomorphization vs GCShape Dictionary
Go uses a hybrid compilation implementation:
- **Monomorphization** for value types (generates specialized machine code for basic types like `int`, `float64`).
- **GCShape Dictionary** for pointer/reference types (shares a single compiled binary code body across pointer types using dictionary pass to save binary size bloat).

---

## ⚠️ Common Limitations of Go Generics
1. **No Generic Methods on Non-Generic Structs**: Methods on a struct cannot introduce NEW type parameters not declared on the parent struct definition.
