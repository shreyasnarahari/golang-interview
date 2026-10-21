# Topic 13 Hard Questions

### Q: Explain the valid conversions allowed for `unsafe.Pointer`.
**Frequency**: sometimes
**Why interviewers ask**: The conversion rules encode Go's contract with the GC; getting them wrong produces memory corruption that only appears under load.
**Answer**: Four conversions are defined:

1. A pointer of any type to `unsafe.Pointer`.
2. `unsafe.Pointer` to a pointer of any type (the target may be a type the memory doesn't actually hold — Go trusts you).
3. `unsafe.Pointer` to `uintptr`.
4. `uintptr` to `unsafe.Pointer`.

Pointer arithmetic **must** be performed as `unsafe.Pointer` → `uintptr` → arithmetic → `unsafe.Pointer` **within a single expression**:

```go
p := unsafe.Pointer(uintptr(unsafe.Pointer(&s.b)) + unsafe.Offsetof(s.b))
```

The compiler cannot prove an intermediate `uintptr` stays valid — the GC may move (future/compacting collectors) or free the object between statements. Additional rules: results of `uintptr` conversions must not be stored, and pointers derived from `unsafe.SliceData`/`StringData` must point into the original allocation. Violations are caught at runtime by `-d=checkptr` instrumentation (enabled under `-race`/`-msan`, or explicitly with `-gcflags=all=-d=checkptr`).
**Follow-ups**: Why is storing a `uintptr` in a struct field a latent bug even if it "works" in tests? What checks does `checkptr` add beyond the race detector?

### Q: What are the cgo pointer-passing rules, and why do they exist?
**Frequency**: rare-senior
**Why interviewers ask**: Breaking these rules is the top cause of production crashes in cgo-heavy codebases; seniors must know them cold.
**Answer**: The rules (enforced at runtime when `GODEBUG=cgocheck=1`, the default for tests and race builds):

1. Go code may pass a Go pointer to C **only if** the memory it points to contains **no Go pointers**. A `*struct{ p *int }` cannot be passed; a `*byte` buffer or a struct of plain scalars can.
2. C code **must not retain** a Go pointer after the call returns — no storing it in C globals or passing it to long-lived C threads.
3. C code receiving a Go pointer must treat the memory as pinned only for the duration of the call.

These exist because the GC moves/frees Go memory and scans Go pointers: a Go pointer hidden in C memory is invisible to the GC, so the object could be collected while C still references it; conversely a Go pointer stored inside Go memory that C can mutate would let C forge invalid Go pointers. To share handles long-term, use an integer indirection table (or `runtime/cgo.Handle`, added in Go 1.17), passing an integer ID to C instead of a pointer.
**Follow-ups**: How does `runtime/cgo.Handle` solve the "C needs a long-lived reference to a Go object" problem? What does `GODEBUG=cgocheck=2` add, and when would you enable it?

### Q: Why is `reflect.ValueOf(&x).Elem()` settable while `reflect.ValueOf(x)` is not — what is the internal mechanism?
**Frequency**: rare-senior
**Why interviewers ask**: This probes real runtime knowledge — the answer lives in `reflect/value.go`, not in tutorials.
**Answer**: Internally a `reflect.Value` is `{typ, ptr, flag}`. The `flag` word packs the `Kind` plus permission bits: `flagRO` (read-only) and `flagAddr` (addressable). `ValueOf(x)` receives `x` boxed into an interface; the copy the `Value` points to is just escape-analyzed stack/heap data that no one else can name, so the runtime sets `flagRO` — allowing mutation would change only an unnameable copy. `(&x).Elem()` walks through the pointer to memory that `x` itself occupies; since that memory is reachable by ordinary assignment, the runtime clears the read-only restriction and marks it addressable/settable. `CanSet()` is exactly "addressable and not read-only". Unexported struct fields are flagged `flagRO` even through a pointer — which is why serializers need `unsafe` to set private fields (and why doing so breaks inlining/aliasing assumptions).
**Follow-ups**: How do libraries like `encoding/json` handle unexported embedded structs? Why does `v.Interface()` on a value with `flagRO` panic instead of returning a copy?

### Q: What is `-d=checkptr`, and which classes of `unsafe` bugs does it catch?
**Frequency**: rare-senior
**Why interviewers ask**: Senior candidates should know how to *validate* unsafe code, not just write it — this is the standard tooling answer.
**Answer**: `checkptr` is compiler instrumentation that inserts runtime checks at every `unsafe.Pointer` → typed-pointer conversion. It is enabled automatically by `-race` and `-msan` builds, at stricter levels via `-gcflags=all=-d=checkptr=1` (cheap checks) or `=2` (adds checks at `unsafe.Slice` and other 1.17+ APIs). It catches:

1. **Arithmetic escaping the object**: converting a pointer that no longer points into the original allocation (e.g. walking past the end of a slice).
2. **Misaligned pointers**: converting to a pointer type whose alignment the address doesn't satisfy (a `*uint64` at an odd address).
3. **Forged header fields**: since Go 1.17, misuse of `unsafe.Slice`/arithmetic that produces out-of-bounds views.

It does **not** catch use-after-free of C memory, races, or type-punning to a *wrong but in-bounds* type — that still requires ASan/valgrind (`-asan`, supported for cgo since Go 1.18) or careful review. Standard CI for unsafe-heavy packages: `go test -race -gcflags=all=-d=checkptr=2 ./...`.
**Follow-ups**: Which `unsafe.Pointer` misuses remain undetectable by checkptr? How does `-asan` complement checkptr for cgo code?
