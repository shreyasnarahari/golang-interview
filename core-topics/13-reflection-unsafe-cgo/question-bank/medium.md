# Topic 13 Medium Questions

### Q: How do you read struct field tags using reflection?
**Frequency**: often
**Why interviewers ask**: Tag parsing is the backbone of JSON/XML/ORM libraries; it tests practical reflection fluency.
**Answer**: Iterate the type's fields and read `StructField.Tag`:

```go
t := reflect.TypeOf(s)
for i := 0; i < t.NumField(); i++ {
    tag := t.Field(i).Tag.Get("json") // "name,omitempty"
}
```

`Tag` is a `reflect.StructTag` (a string); `.Get("json")` parses `` `json:"..."` `` and returns the value, while `.Lookup("json")` additionally reports whether the key was present (distinguishing an empty value from an absent key). Reflection works on the **type**, so it only sees exported field metadata; unexported fields are visible via `Field(i)` but their values cannot be read with `.Interface()` (it panics).
**Follow-ups**: How would you build a minimal JSON-style encoder that honors `json:"-"`? Why does reading an unexported field's value panic, and how do some libraries bypass that with `unsafe`?

### Q: What happens when you call `reflect.ValueOf(x).SetInt(7)`, and how do you fix it?
**Frequency**: often
**Why interviewers ask**: Behavior prediction around addressability separates people who have debugged reflection from those who have only read about it.
**Answer**: It panics: `reflect: reflect.Value.SetInt using unaddressable value`. `ValueOf(x)` receives a **copy** of `x` inside an interface; the resulting `Value` points at that copy and is not settable — mutating it would silently discard the change, so Go panics instead. Fix by passing a pointer and dereferencing:

```go
v := reflect.ValueOf(&x).Elem()
if v.CanSet() {
    v.SetInt(7)
}
```

The same rule applies to struct fields reached through an unaddressable struct, and to unexported fields, which are never settable even through a pointer.
**Follow-ups**: Why is `reflect.ValueOf(&s).Elem().Field(0)` settable but a field obtained after copying the struct is not? How would you write a generic `SetField(obj any, name string, val any)` helper safely?

### Q: How do you convert between `string` and `[]byte` without copying in Go 1.20+?
**Frequency**: sometimes
**Why interviewers ask**: Zero-copy conversion is the most common legitimate use of `unsafe`; the Go 1.20 helpers replaced older header-hacking tricks.
**Answer**: Use `unsafe.StringData`, `unsafe.Slice`, and `unsafe.String`:

```go
s := "hello"
b := unsafe.Slice(unsafe.StringData(s), len(s)) // []byte view of s
s2 := unsafe.String(&b[0], len(b))              // string view of b
```

Both avoid allocation by sharing the backing array. The danger: **strings are immutable**, so mutating `b` corrupts `s` — and constant strings may live in read-only memory, causing a segfault. `unsafe.StringData("")` returns an unspecified, possibly nil, pointer for an empty string, and `unsafe.String(ptr, 0)` yields `""` even if `ptr` is nil. The old pattern of forging `reflect.StringHeader`/`SliceHeader` is deprecated and unsafe across GC changes; the 1.20 APIs are the supported form.
**Follow-ups**: When is a zero-copy conversion actually worth it (hint: large buffers in hot paths, e.g. parsing)? Why is `string(bytes)` normally a copy, and how does the compiler already optimize `string(b)` in map lookups?

### Q: Why is reflection slower than direct code, and how do you mitigate it?
**Frequency**: often
**Why interviewers ask**: Tests whether the candidate can make sound engineering trade-offs rather than reciting "reflection is slow."
**Answer**: Reflection costs come from: (1) **dispatch through interfaces and runtime type metadata** — every field access goes through `reflect.Value` indirection instead of a compiled offset; (2) **allocations** — boxing values into interfaces (`v.Interface()`) and building `Value` wrappers escape to the heap; (3) **lost optimizations** — the compiler cannot inline or constant-fold reflected calls, and `Call` on a `Method` uses a generic trampoline (similar in spirit to FFI). Benchmarks typically show reflected field access several times slower than direct access. Mitigations: cache `reflect.Type`/`StructField` lookups once (as `encoding/json` does per type), precompute method `Value`s, generate code instead (`go generate`, like `easyjson`), or restrict reflection to init/warmup paths rather than per-request hot loops.
**Follow-ups**: Why does `encoding/json` pay reflection cost on every marshal — and how do `json.Marshaler` and code generation change that? How would you measure reflection overhead (`go test -bench`, `-gcflags=-m` for escapes)?
