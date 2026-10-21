# Topic 04 Hard Questions

### Q: Explain the runtime structure of `itab` in `iface`.
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates who have read the runtime from those who only use interfaces.
**Answer**: A non-empty interface value is an `iface`: `{tab *itab, data unsafe.Pointer}`. The `itab` stores:
1. `inter *interfacetype`: description of the interface type.
2. `_type *_type`: description of the concrete type stored in the value.
3. `hash uint32`: copy of `_type.hash` for fast type-switch/assertion comparisons.
4. `fun [1]uintptr`: variable-length array (declared size 1, allocated to fit) of code pointers to the concrete methods that implement the interface's methods.

`itab` pairs are computed once per (interface, concrete type) combination and cached in a global hash table, so conversions after the first are cheap lookups. A dynamic call is essentially `itab.fun[i](data)` — one extra indirection versus a static call, which is why the compiler devirtualizes hot calls when it can prove the concrete type (and PGO profiles help it do so).
**Follow-ups**: Why can't `fun` pointers point directly at value-receiver methods when the interface holds a pointer? (Method wrappers.) What does a type switch compile down to? (Comparing `itab`/`hash`, or sequential type comparisons.)

### Q: What is the difference between `iface` and `eface` in the runtime, and why does the distinction exist?
**Frequency**: rare-senior
**Why interviewers ask**: It probes whether the candidate understands the actual memory layout of interface values and the nil check.
**Answer**: `eface` backs the empty interface `any` and is `{_type *_type, data unsafe.Pointer}` — no `itab` is needed because there are no methods to dispatch. `iface` backs non-empty interfaces and carries the `itab`. Both are two words wide (16 bytes on 64-bit), which is why assigning any value to `any` may cause the *data* to be boxed: if the concrete value doesn't fit in one pointer word, the runtime heap-allocates a copy and `data` points to it (a hidden allocation visible in escape analysis). `iface` converts to `eface` cheaply; the reverse requires a type assertion that rebuilds/looks up an `itab`.
**Follow-ups**: Why does `fmt.Println(intSlice)` copy the slice header but not the backing array? Why can storing a small non-pointer value in `any` still allocate?

### Q: How does struct field alignment affect memory usage, and how do you optimize it?
**Frequency**: sometimes
**Why interviewers ask**: Memory layout awareness matters for high-throughput services and cache-sensitive code.
**Answer**: Each field is placed at an offset that is a multiple of its alignment (its size for basic types, capped at the word size), and the struct's total size is padded to a multiple of its largest field's alignment. Ordering fields largest-to-smallest minimizes padding. Verified on amd64:
```go
type Bad struct {           // 24 bytes
	a bool  // +7 padding
	b int64
	c bool
	d int32 // +3 trailing padding
}
type Good struct {          // 16 bytes
	b int64
	d int32
	a, c bool
}
```
Savings multiply across slices/arrays of structs (24→16 bytes per element is a 33% reduction and better cache density). Tools: `unsafe.Sizeof`/`unsafe.Alignof` to check, and the `fieldalignment` analyzer (from `golang.org/x/tools`, available via golangci-lint's `govet`) to find waste automatically. Caveat: for false-sharing-sensitive concurrent code you sometimes *add* padding deliberately (e.g. cache-line padding).
**Follow-ups**: Why is a struct's size rounded up to its largest alignment? When would you deliberately insert padding instead of removing it?

### Q: How do interfaces and generics differ in how the compiler generates code, and what are the performance trade-offs?
**Frequency**: rare-senior
**Why interviewers ask**: Go 1.18+ generics changed the polymorphism story; interviewers want to know you can choose between them on evidence.
**Answer**: Interfaces are fully dynamic: one compiled function works for all types, with dispatch through the `itab` and values boxed behind pointers (extra indirection, possible heap allocation, calls usually not inlinable unless devirtualized). Generics are *stenciled* at compile time, but Go deduplicates instantiations by **GC shape**: types with identical underlying memory layouts (e.g. all pointer types, or `int`/`uint`) share one generated function body parameterized by a shape dictionary. So `Max[int]` and `Max[uint]` share code, while `Max[string]` gets its own. Consequences: generic calls on scalars are static, inlinable, and allocation-free — often faster than the interface version — but code size grows with the number of distinct shapes, and generic code operating on interface-like constraints still boxes where it must treat values uniformly.
**Follow-ups**: When would an interface still beat generics (e.g. heterogeneous collections, plugin boundaries, `io.Reader` chains)? How would you measure the difference? (Benchmark both, check `-gcflags=-m` inlining/escape decisions.)
