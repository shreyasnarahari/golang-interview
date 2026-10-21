# Topic 10 Hard Questions

### Q: Explain the GCShape stenciling model the Go compiler uses for generics, and why it is not C++-style monomorphization.
**Frequency**: rare-senior
**Why interviewers ask**: Senior candidates are expected to reason about the runtime cost model of generics, not just the syntax.
**Answer**: Go compiles each generic function/type once per *GCShape* — an equivalence class of types with identical memory layout and GC pointer/bitmap structure — rather than once per concrete type. All pointer types (`*int`, `*string`, `*Node`) share one GCShape and thus one stenciled function body; `int`, `string`, and `struct{...}` shapes each get their own. Type-specific information (sizes, how to call methods, how to compare) is passed at runtime via a *dictionary* argument, so a single stencil serves many types. This is deliberately not full monomorphization: binaries stay far smaller (no exponential instantiation bloat) at the cost of some indirection — method calls and interface conversions on type parameters go through the dictionary and are harder to inline/devirtualize. Go 1.18+ uses this "dictionaries and gcshape stenciling" design from the proposal by the Go team.
**Follow-ups**: Which operations on a type-parameter value force dictionary lookups? How does this model affect inlining and benchmark numbers versus hand-written per-type code?

### Q: Why can't a method declare its own type parameters, e.g. `func (s Stack[T]) Map[U any](f func(T) U) Stack[U]`?
**Frequency**: rare-senior
**Why interviewers ask**: It reveals whether the candidate understands an actual implementation-driven language limitation rather than cargo-culting "Go generics are limited".
**Answer**: The Go spec forbids type parameters on methods: a method may only use the type parameters declared on its *receiver's* type. The reason is implementation strategy: Go instantiates generic code at compile time via gcshape stenciling, but interface method dispatch happens at runtime through interface tables. If methods could be generic, the set of methods a generic type has would not be knowable until a caller instantiates the method — the compiler would have to generate code for method instantiations it never saw, or build interface tables at runtime, which the design explicitly excludes. Workaround: write a plain generic *function* (`func Map[S ~[]T, T, U any](s S, f func(T) U) []U`), which is exactly what `slices`/`maps` and iterator helpers do.
**Follow-ups**: How do interface method tables interact with generic instantiations? Is this restriction likely to change, and what would it cost?

### Q: How do range-over-func iterators (Go 1.23) work mechanically — what is `iter.Seq`, what does `yield` returning `false` mean, and what happens on `break`?
**Frequency**: sometimes
**Why interviewers ask**: Iterators are the newest major language feature and interviewers use them to check whether a candidate keeps current with Go's evolution.
**Answer**: Since Go 1.23, `for range` accepts a function value. `iter.Seq[V]` is just `func(yield func(V) bool)`: the compiler lowers the loop into a call of your function, passing the loop body as `yield`. Each iteration invokes `yield(v)`; if the loop body finishes normally it returns `true` and you continue, but if the body executes `break` (or the loop otherwise terminates early), the compiler-generated `yield` returns `false` and your iterator must stop producing and return promptly — ignoring it is a bug that can leak resources or loop forever. The `iter` package provides `Seq`, `Seq2` (key/value), and adapters like `iter.Pull` (converting a push iterator into next/stop calls). The design lets any container expose iteration without allocating a separate iterator object.
**Follow-ups**: How would you write an iterator that must release a resource even when the caller breaks early (defer inside the Seq function)? What is the difference between `iter.Seq` and `iter.Pull`-style pull iterators?

### Q: What is the status of generic type aliases in current Go, and why were they harder than generic defined types?
**Frequency**: rare-senior
**Why interviewers ask**: It separates candidates tracking the actual language roadmap from those repeating Go 1.18-era limitations.
**Answer**: As of Go 1.24, generic type aliases are fully supported (they shipped experimentally behind `GOEXPERIMENT=aliastypeparams` in Go 1.23):
```go
type Set[T comparable] = map[T]struct{}
```
Unlike a generic *defined* type (`type Set[T comparable] map[T]struct{}`), an alias introduces no new type: `Set[int]` *is* `map[int]struct{}`, with identical identity, method set, and assignability. They were harder because aliases only create names, not types — a parameterized alias must be expanded at every use site while still recording the instantiation for type identity, reflection (`reflect` must not see a fake type), and export data across package boundaries, which required compiler and exporter work.
**Follow-ups**: When is an alias preferable to a defined generic type (hint: gradual refactoring, re-exporting another package's instantiation)? Can an alias declaration add methods? (No.)
