# Topic 09 Hard Questions

### Q: How does `errors.As` inspect custom error types through pointer reflection?
**Frequency**: rare-senior
**Why interviewers ask**: It probes whether the candidate understands the reflection machinery behind the API, not just its call signature.
**Answer**: `errors.As(err, target)` requires `target` to be a non-nil pointer to a type that implements `error` or to an interface type; anything else panics. It walks the unwrap chain, and at each node uses reflection to test whether the node's concrete type is *assignable to* the type `target` points to. On the first match it uses `reflect.Value.Set` to write the node into the memory behind `target` and returns `true`. Consequences:
- Assignability, not equality: a target of `*error` (interface) matches the first node in the chain.
- Custom types can intercept matching by defining `As(target any) bool`.
- Because it writes through a pointer, `errors.As` allocates/reflects more than `errors.Is` — in hot paths a direct type assertion on an unwrapped error is cheaper, though rarely worth it.
**Follow-ups**: Why does `errors.As` panic on a bad target instead of returning false? How does the `As(target any) bool` override change traversal?

### Q: Why is a function that returns a typed nil pointer as `error` a bug, and how do you prevent it?
**Frequency**: sometimes
**Why interviewers ask**: The typed-nil trap is the classic Go gotcha that reveals genuine understanding of interface representation (type + value pair).
**Answer**: An `error` interface value is nil only when both its type and value words are nil. Returning a typed nil pointer stores a non-nil type word, so `err != nil` is true even though the pointer is nil:
```go
type AppError struct{ Msg string }
func (e *AppError) Error() string { return e.Msg }

func lookup() error {
    var e *AppError // nil
    if notFound {
        return e // BUG: non-nil interface wrapping a nil pointer
    }
    return nil
}
```
`err != nil` passes, and calling `err.Error()` panics (nil pointer dereference) unless the method guards against nil receivers. Prevention: declare the return path as the interface (`return nil` literally), run `go vet`/staticcheck which flag some cases, and if the API must return the concrete type, make `Error()` nil-receiver-safe and document it.
**Follow-ups**: Why doesn't `go vet` catch every instance of this? How does this interact with `errors.Is(err, nil)`?

### Q: How do the `Is(error) bool` and `As(any) bool` method overrides on custom error types work?
**Frequency**: rare-senior
**Why interviewers ask**: These hooks are how sophisticated packages (e.g. `os`, gRPC status) customize error matching, and senior candidates should know they exist and why.
**Answer**: At each node of the unwrap chain, `errors.Is` checks, in order: (1) `err == target` if target is comparable, then (2) if the node implements `Is(error) bool`, calls it. `errors.As` similarly consults an `As(any) bool` method before falling back to reflection-based assignability. Use cases:
```go
type QuotaError struct{ Resource string }

func (e *QuotaError) Is(target error) bool {
    t, ok := target.(*QuotaError)
    return ok && t.Resource == "" // empty Resource matches any QuotaError
}
```
This lets callers match a *class* of errors while ignoring fields (`errors.Is(err, &QuotaError{})` matches any quota failure), mirroring how `os.ErrNotExist` matches many distinct underlying errors. Caution: `Is` must be reflexive and cheap — it runs once per chain node per query; also these methods make equality non-symmetric logic that is easy to get subtly wrong.
**Follow-ups**: Why must `Is` be reflexive (`e.Is(e)` should be true)? How would you design a match-by-code custom type like gRPC's status errors?

### Q: How do `errors.Is`/`errors.As` traverse a multi-error tree from `errors.Join`, and what are the two `Unwrap` shapes?
**Frequency**: rare-senior
**Why interviewers ask**: Post-1.20, errors form trees, not chains, and traversal semantics matter for correctness of matching in fan-out systems.
**Answer**: There are two unwrap forms: `Unwrap() error` (chain node, from `fmt.Errorf` with a single `%w`) and `Unwrap() []error` (multi node, from `errors.Join` or multiple `%w`). `errors.Is`/`errors.As` perform a preorder depth-first walk: they test the current node, then recurse into the single child or, for `[]error`, into each child in order, first match wins. Implications:
- A `nil` inside a joined `[]error` is harmless: `Join` drops nil inputs (and returns `nil` if every input is nil).
- A single type *cannot* implement both unwrap forms — Go forbids two methods named `Unwrap` differing only by return type, and embedding two types that each provide one makes the promoted method ambiguous, so neither interface is satisfied. Every node is either a chain node or a multi node.
- Matching can succeed against any leaf of the tree, so `errors.Is(joined, os.ErrNotExist)` is true if *any* constituent wraps it — design join sites so that "any match" is the semantics callers want.
- The joined error's `Error()` concatenates child messages with newlines; there is no built-in "give me all matches" API — you must walk `Unwrap() []error` yourself to collect them.
**Follow-ups**: In what order are children of `errors.Join(a, errors.Join(b, c))` visited? How would you collect *all* errors in a tree matching a target, given that `errors.Is` stops at the first?
