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

### 3. Compilation Model: GCShape Stenciling (not full monomorphization)
Unlike Rust or C++, Go does **not** generate a fully specialized code body per type argument. Instead the compiler uses **gcshape-based stenciling** for *all* instantiations:
- Types are grouped by their **gcshape** — size, alignment, and pointer/has-pointer layout. All types sharing a gcshape share **one compiled code body**.
- e.g., `int` and `float64` share a gcshape (8 bytes, pointer-free) → one instantiation; `*int`, `*string`, `unsafe.Pointer` share another (pointer-shaped) → one instantiation.
- A **dictionary** carrying type-specific metadata (type descriptors, itabs for interface calls) is passed so the shared code body can handle the differences.

The result: far less binary bloat than full monomorphization, at the cost of a dictionary lookup for some type-specific operations (and interface method calls staying indirect).

### 4. Range-Over-Func Iterators (Go 1.23+)
Go 1.23 lets `for ... range` iterate over **functions**, and the `iter` package defines the two canonical signatures as **generic function types** — generics are what make the feature type-safe:

```go
type Seq[V any]     func(yield func(V) bool)      // single values: slices, lists
type Seq2[K, V any] func(yield func(K, V) bool)   // pairs: maps, indexed data
```

`for v := range someSeq` compiles into a call to `someSeq` with an implicit `yield`; the loop body runs per call, and `break`/`return` make `yield` return `false` so the iterator stops early. Iterating a `Seq2` gives `for k, v := range seq2`.

```go
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

// Backward is a generic iterator yielding a slice's elements in reverse.
func Backward[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(s[i]) {
				return // consumer broke out of the loop
			}
		}
	}
}

func main() {
	for v := range Backward([]string{"a", "b", "c"}) {
		fmt.Print(v, " ") // c b a
	}
	fmt.Println()
	fmt.Println(slices.Sorted(maps.Keys(map[string]int{"x": 1, "y": 2})))
}
```

Stdlib producers/consumers to know: `maps.Keys` / `maps.Values` (return `iter.Seq` in 1.23+), `slices.All` / `slices.Backward`, and collectors `slices.Collect(seq)` (`Seq[V] -> []V`) and `slices.Sorted(seq)`.

---

## ⚠️ Common Limitations of Go Generics
1. **No Generic Methods on Non-Generic Structs**: Methods on a struct cannot introduce NEW type parameters not declared on the parent struct definition.
2. **No Operator Constraints Without Type Sets**: You cannot constrain `T` to "types supporting `+`" directly — you must enumerate a type union (`~int | ~float64 | ...`) or use `cmp.Ordered` / `golang.org/x/exp/constraints`.
3. **Type Parameters Are Not Covariant**: `[]MyType` does not satisfy `[]any`, and generics do not add covariance — a `List[Derived]` is not a `List[Base]`.
4. **Interface Constraints Can't Be Ordinary Interfaces Everywhere**: A constraint interface containing type elements (`~int | string`) can only be used as a constraint, not as a variable type.
5. **Generic Code Still Boxes at Runtime Boundaries**: Passing a generic value into an `any` parameter still incurs interface boxing; generics remove boxing only inside statically-instantiated generic code.
