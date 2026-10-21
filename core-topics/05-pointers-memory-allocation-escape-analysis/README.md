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
1. **Returned Pointer**: A pointer to a local variable is returned from the function.
2. **`interface{}` Boxing**: A variable is assigned to an `interface{}` parameter (e.g. `fmt.Println(x)`).
3. **Dynamic / Unknown Sizes**: A slice or buffer whose size is not known at compile time outlives the frame (note: current Go stack-allocates non-constant-size slices that do *not* escape, with a runtime size check falling back to the heap for large sizes).
4. **Stack Frame Bounds**: A variable exceeds the stack frame size allocation bounds (e.g. a very large local array).
5. **Closure Capture by Reference**: A pointer to a variable is captured by a closure that outlives the frame.
6. **Channel / Map Storage**: Storing pointers in channels or maps, where the compiler cannot prove the referent stays local.

### 3. Inspecting Escape Analysis
Run the Go compiler with memory analysis flags:
```bash
go build -gcflags="-m"
```
Output like `moved to heap: x` or `escapes to heap` explicitly reveals allocation choices made by the compiler.

For more detail, use double verbosity:
```bash
go build -gcflags="-m -m"
```
The second `-m` adds the escape *flow*: the chain of assignments (`flow: ~r0 ← &x: ...`) showing exactly why a variable was moved to the heap.

Inlining interacts with escape analysis: inlinable functions may avoid escape because their locals are analyzed in the caller's frame — check for `can inline` lines in the `-m` output.

**Hands-on**: the [lab/escapehunt](lab/escapehunt/) scavenger hunt lets you predict stack vs heap for 10 scenarios and verify against real `-m` output.

---

## ⚠️ Common Gotchas
1. **Premature Pointer Optimization**: Passing small structs by pointer (`*SmallStruct`) to "save memory" can actually degrade performance if it forces heap escaping instead of remaining on stack!
