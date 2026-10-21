# Topic 13 Easy Questions

### Q: What is the difference between `reflect.TypeOf` and `reflect.ValueOf`?
**Frequency**: often
**Why interviewers ask**: It checks whether the candidate understands the two fundamental entry points of reflection before any deeper discussion.
**Answer**: `reflect.TypeOf(i)` returns a `reflect.Type` describing the **dynamic type** stored in the interface (methods, fields, kind, size); `reflect.ValueOf(i)` returns a `reflect.Value` wrapping the **concrete value** so you can read (and, if addressable, modify) the data. Note that `reflect.TypeOf(nil)` returns `nil` because a nil interface carries no dynamic type. You can go back from a `Value` to an interface with `v.Interface()`, and get the type from a value with `v.Type()`.
**Follow-ups**: What does `reflect.TypeOf(nil)` return and why? How do you get a `reflect.Type` for a nil pointer type without a value (hint: `reflect.TypeOf((*T)(nil))`)?

### Q: What is the third law of reflection in Go?
**Frequency**: often
**Why interviewers ask**: Settability is the most common reflection bug source; interviewers want to see if you've hit it and understood it.
**Answer**: To modify a reflection object, the value must be **settable** — obtained by passing a pointer and calling `.Elem()`:

```go
v := reflect.ValueOf(&x).Elem() // settable
v.SetInt(42)                    // OK
```

`reflect.ValueOf(x)` copies `x`, so the resulting `Value` is unaddressable and any `Set*` call panics with `reflect: reflect.Value.SetInt using unaddressable value`. Check with `v.CanSet()` first.
**Follow-ups**: Why does Go require a pointer here rather than letting you set the copy? What is the difference between `CanSet` and `CanAddr`?

### Q: What does `unsafe.Sizeof` return, and what guarantees does Go make about it?
**Frequency**: sometimes
**Why interviewers ask**: A quick filter for whether the candidate knows `unsafe` basics and Go's memory layout model.
**Answer**: `unsafe.Sizeof(x)` returns the size in bytes of the variable's type as a **`uintptr` compile-time constant** — it does not evaluate `x`. It counts only the fixed-size part: `unsafe.Sizeof("hello")` is 16 (the string header: pointer + length on 64-bit), not 5. Alongside it: `unsafe.Offsetof(structField)` gives a field's byte offset and `unsafe.Alignof(x)` the alignment requirement. Sizes depend on architecture and field ordering/padding, so never hard-code them.
**Follow-ups**: Why can reordering struct fields shrink `unsafe.Sizeof` of a struct? What is the size of an empty struct `struct{}`?

### Q: How do you call a C function from Go?
**Frequency**: sometimes
**Why interviewers ask**: Confirms the candidate has at least seen the cgo mechanism and knows it is not a normal import.
**Answer**: Use cgo: write C declarations in the comment directly above `import "C"`, then call them as members of the pseudo-package `C`:

```go
// #include <stdlib.h>
import "C"

p := C.malloc(C.size_t(64))
defer C.free(p)
```

Types map explicitly (`C.int`, `C.size_t`, `C.CString(s)` — which allocates C memory you must free). cgo requires a C toolchain, disables cross-compilation by default (`CGO_ENABLED=0` for pure-Go builds), and is commonly disabled for static binaries.
**Follow-ups**: Why does `CGO_ENABLED=0` matter for containers like `scratch`/`distroless` images? What happens to goroutine scheduling while a C call runs?
