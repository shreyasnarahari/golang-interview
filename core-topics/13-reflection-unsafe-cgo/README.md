# Topic 13: Reflection, Unsafe, and CGO

## 📌 Overview
Go offers powerful low-level escapes from the standard type system: runtime type inspection via `reflect`, memory manipulation via `unsafe`, and C interop via `cgo`.

---

## 🎯 Key Concepts

### 1. The `reflect` Package
- **`reflect.TypeOf(i)`**: Returns `reflect.Type` representing concrete runtime type info.
- **`reflect.ValueOf(i)`**: Returns `reflect.Value` holding concrete runtime data value.
- **Laws of Reflection**:
  1. Reflection goes from interface value to reflection object.
  2. Reflection goes from reflection object to interface value (`v.Interface()`).
  3. To modify a reflection object, the value must be **settable** (must pass a pointer `reflect.ValueOf(&x).Elem()`).

### 2. The `unsafe` Package
Allows bypassing Go's type safety and memory safety guarantees:
- **`unsafe.Pointer`**: Arbitrary pointer type that can convert to/from any pointer type or `uintptr`.
- **`uintptr`**: Integer representation of memory address (used for pointer arithmetic, but NOT tracked by Garbage Collector!).
- **Go 1.20+ Zero-Copy Functions**:
  - `unsafe.StringData(str)`: Returns pointer to string's byte array.
  - `unsafe.Slice(ptr, len)`: Returns slice referencing memory pointer.

### 3. CGO Mechanics & Overhead
- `import "C"` allows calling C code directly from Go.
- **Performance Overhead**: CGO calls require stack switching between Go contiguous stack and OS thread C stack, context switching costs (~100–200 ns vs ~1–2 ns for pure Go call).

---

## ⚠️ Common Gotchas
1. **Holding `uintptr` Across Allocations**: Storing a memory address in `uintptr` does NOT update if GC moves memory or sweeps the object. Convert back to `unsafe.Pointer` immediately in a single expression!
