# Topic 04: Structs, Interfaces, and Methods

## 📌 Overview
Go replaces traditional class inheritance with struct composition and implicit interfaces. Understanding method receivers (value vs pointer) and interface runtime structures (`iface` / `eface`) is vital for Go design rounds.

---

## 🎯 Key Concepts

### 1. Value Receiver vs Pointer Receiver Rule Matrix
- **Value Receiver (`func (s Struct) Method()`)**: Works on a copy of the struct. Modifications do not affect the original caller.
- **Pointer Receiver (`func (s *Struct) Method()`)**: Works directly on the pointer. Modifications mutate the caller.
- **Interface Satisfaction Rule**:
  - A **Pointer (`*T`)** has both value and pointer receiver methods in its method set.
  - A **Value (`T`)** ONLY has value receiver methods in its method set!

| Method Receiver Type | Called on Value `T` | Called on Pointer `*T` | Satisfies Interface requiring `*T` method? |
|---|---|---|---|
| `(s Struct)` | Yes | Yes (Go auto-dereferences) | Yes |
| `(s *Struct)` | Yes (Go auto-takes address if addressable) | Yes | **Only `*T` satisfies interface!** |

### 2. Interface Representation (`iface` vs `eface`)
Interfaces are implemented at runtime by two key C structures in the Go runtime:
1. **`iface`**: Interface with methods (e.g. `io.Reader`).
   - `tab *itab`: Pointer to interface table holding type descriptor + method pointers.
   - `data unsafe.Pointer`: Pointer to concrete data object.
2. **`eface`**: Empty interface (`interface{}` / `any`).
   - `_type *_type`: Pointer to runtime type information.
   - `data unsafe.Pointer`: Pointer to concrete data object.

```
iface (16 bytes)
+-------------------------+
| tab  *itab              | ---> [Type Info | Method Function Pointers]
+-------------------------+
| data unsafe.Pointer     | ---> [Concrete Value Memory Address]
+-------------------------+
```

### 3. The `nil` Interface Trap
An interface variable is `nil` **ONLY IF both `tab`/`_type` AND `data` are `nil`**.
If you assign a typed `nil` pointer (e.g., `var p *MyStruct = nil`) to an interface variable, the interface has a valid type descriptor, so `iface != nil` evaluates to `true`!

---

## ⚠️ Common Gotchas
1. **Interface Nil Trap**: Returning a concrete `nil` pointer cast to an `error` interface causes `err != nil` checks to pass incorrectly. Always return explicit untyped `nil`.
