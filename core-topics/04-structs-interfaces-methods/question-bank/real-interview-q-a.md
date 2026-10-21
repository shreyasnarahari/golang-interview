# Topic 04 Real Interview Q&A

### Q: (Asked at a fintech infrastructure company) "Your team keeps shipping bugs where a function returns a concrete error type. Show us the compile-time trick to keep a type honest against an interface, and explain why the typed-nil bug still slipped through."
**Frequency**: sometimes
**Why interviewers ask**: Fintechs care about error-handling correctness; this tests both interface mechanics and production debugging experience.
**Answer**: The compile-time compliance check is a package-level declaration with no runtime cost:
```go
var _ MyInterface = (*MyStruct)(nil)
```
If `*MyStruct` ever stops implementing `MyInterface` (e.g. a renamed method), the build fails immediately with a clear message. But it does **not** protect against the typed-nil bug: `var e *MyErr = nil; return e` from a function returning `error` compiles fine (the method set matches) yet produces a non-nil interface wrapping a nil pointer, so `err != nil` checks pass and `err.Error()` can panic. Defenses: always return the literal `nil` from error paths, have constructors return the interface type directly, and lint for returning concrete nils.
**Follow-ups**: Why does `var _ MyInterface = MyStruct{}` (value form) sometimes fail where the pointer form passes? How would you write a test that catches a typed-nil return?

### Q: (Asked at a cloud infra company) "We're designing a plugin system for a storage backend. Would you model backends with embedding or interfaces, and walk us through the receiver and nil-handling decisions?"
**Frequency**: sometimes
**Why interviewers ask**: Cloud infra code is interface-heavy (`io.Reader`, gRPC handlers); this tests design judgment, not syntax recall.
**Answer**: Interfaces at the boundaries, embedding only for shared implementation. The consumer defines a narrow interface (e.g. `type Store interface { Get(ctx context.Context, k string) ([]byte, error) }`) rather than inheriting a base struct — embedding gives no substitutability and couples plugins to a concrete type. Receivers: backends are stateful (connection pools, `sync.Mutex`), so use pointer receivers throughout and store `*Backend` values; mixing receivers breaks interface satisfaction for the value form. Nil handling: constructors return the interface type and return literal `nil` on failure; never return a nil `*Backend` as a `Store`. For tests, hand-roll fakes or embed the interface in the fake struct (`struct { Store }`) so unimplemented methods compile but panic if called — embedding an interface is the one case where embedding *does* give substitutability.
**Follow-ups**: Why keep the interface small and consumer-side? What breaks if you embed a value `Backend` (not pointer) in a wrapper struct?

### Q: (Asked at a high-frequency trading / low-latency shop) "Profile shows allocations inside our hot path where we pass small structs as `any` into a logger. Why does that allocate, and what would you change?"
**Frequency**: rare-senior
**Why interviewers ask**: Latency-sensitive teams need engineers who understand interface boxing, not just "avoid interfaces."
**Answer**: An interface value is two words: (type, data pointer). A concrete value larger than one word (or any non-pointer value whose address can't be taken safely) is copied to the heap so `data` can point at it — that's the allocation, visible via `go build -gcflags=-m` ("escapes to heap"). Fixes, in order of preference: (1) switch to `slog`, whose `LogAttrs`/`slog.Int` etc. pass values as `slog.Value` — a union representation that avoids boxing for common kinds; (2) pass pointers (`*T` fits in one word, no boxing copy) if sharing is acceptable; (3) replace the `any` parameter with a generic function or a typed method so no interface conversion happens at all; (4) keep structs small enough that the cost is acceptable and pool buffers with `sync.Pool` for the surrounding machinery. Verify each step with benchmarks and the allocation counts from `testing.B.ReportAllocs`.
**Follow-ups**: Why doesn't passing a pointer box? How does `slog.Value` avoid the allocation for ints and strings?
