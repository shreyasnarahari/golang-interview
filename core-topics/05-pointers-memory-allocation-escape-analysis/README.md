# Topic 05: Pointers, Memory Allocation, and Escape Analysis

## 📌 Overview
Unlike C/C++ where developers explicitly choose stack (`stack`) vs heap (`malloc`) allocation, the Go compiler automatically determines whether a variable resides on the stack or heap via **Escape Analysis**.

---

## 🎯 Key Concepts

### 1. Stack vs Heap Allocations
- **Stack Allocation**:
  - Extremely fast (pointer increment).
  - Automatically freed when function call frame finishes.
  - No Garbage Collection (GC) overhead.
- **Heap Allocation**:
  - Requires memory allocation through Go memory manager (`mcache`/`mheap`).
  - Managed and cleaned up by the concurrent GC.
  - Introduces allocation latency and GC work.

### 2. Escape Analysis Rules
Go compiler determines if a variable can be safely allocated on the stack. If a reference to the variable outlives the current stack frame, it **escapes to the heap**.

**Common Escape Conditions**:
1. **Returning a Pointer**: Returning a pointer or reference to a local variable created inside a function.
2. **Interface Parameters**: Passing variables to functions taking `interface{}` parameters (e.g. `fmt.Println(x)`).
3. **Dynamic / Large Sizes**: Allocating slices or buffers whose sizes are unknown at compile time or exceed stack limits.
4. **Channel / Map / Closure Storage**: Storing pointers in channels, maps, or closed-over variable references.

### 3. Inspecting Escape Analysis
Run the Go compiler with memory analysis flags:
```bash
go build -gcflags="-m"
```
Output like `moved to heap: x` or `escapes to heap` explicitly reveals allocation choices made by the compiler.

---

## ⚠️ Common Gotchas
1. **Premature Pointer Optimization**: Passing small structs by pointer (`*SmallStruct`) to "save memory" can actually degrade performance if it forces heap escaping instead of remaining on stack!
