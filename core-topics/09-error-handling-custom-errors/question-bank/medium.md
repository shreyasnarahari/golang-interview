# Topic 09 Medium Questions

### Q: What feature was introduced in Go 1.20 for joining multiple errors?
**Frequency**: sometimes
**Why interviewers ask**: It separates candidates who keep up with the standard library from those still on Go 1.13-era patterns, and it matters for batch/parallel code.
**Answer**: `errors.Join(err1, err2, ...)`. It combines multiple errors into one composite error. Key behaviors:
- Nil inputs are skipped; `Join(nil, nil)` returns `nil`.
- The joined error's `Error()` string is the children's messages separated by newlines.
- Its `Unwrap() []error` method returns the children, forming an error *tree* that `errors.Is` and `errors.As` traverse depth-first.
Typical use: aggregating failures from a fan-out of goroutines or a batch operation instead of keeping only the first error. Go 1.20 also enabled multiple `%w` verbs in `fmt.Errorf`, which produces the same `Unwrap() []error` shape.
**Follow-ups**: How do `errors.Is`/`errors.As` traverse a joined error? When would you prefer `errors.Join` over returning the first error encountered?

### Q: What does this print, and why?
```go
err := fmt.Errorf("open config: %w", os.ErrNotExist)
fmt.Println(err == os.ErrNotExist, errors.Is(err, os.ErrNotExist))
```
**Frequency**: often
**Why interviewers ask**: Behavior prediction on wrapping exposes whether the candidate actually understands the unwrap chain or just memorized API names.
**Answer**: It prints `false true`. `==` compares interface values directly: the outer error is a `*fmt.wrapError` holding the sentinel inside, so it is not equal to `os.ErrNotExist`. `errors.Is` unwraps the chain — outer error → `os.ErrNotExist` — and finds a match at the second level. Rule of thumb: use `==` only when you are certain the error was never wrapped (rare across package boundaries); otherwise use `errors.Is`.
**Follow-ups**: What concrete type does `fmt.Errorf("...: %w", err)` actually return? How would the answer change with `errors.Join(os.ErrNotExist, io.EOF)`?

### Q: When should you define a custom error type instead of using a sentinel error?
**Frequency**: often
**Why interviewers ask**: Choosing between sentinels and typed errors is the core design decision in Go error handling and reveals API-design judgment.
**Answer**: Use a sentinel (`var ErrX = errors.New(...)`) when callers only need to know *that* a condition occurred. Define a custom type when callers need structured context to react programmatically:
```go
type ValidationError struct {
    Field string
    Msg   string
}
func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }
```
Callers extract it with `errors.As`. Guidelines: keep the taxonomy small and part of your documented API (every exported error is a compatibility promise); implement `Error()` on the pointer receiver if the type carries mutable state or you want nil-checkable semantics; do not export error types callers have no reason to inspect.
**Follow-ups**: Why does exporting an error type become an API compatibility commitment? Should `Error()` go on the value or pointer receiver, and why?

### Q: Why must the target passed to `errors.As` be a non-nil pointer, and what panics otherwise?
**Frequency**: sometimes
**Why interviewers ask**: It tests whether the candidate understands `errors.As` mechanically (reflection-based assignment) rather than as a magic cast.
**Answer**: `errors.As(err, target)` needs to *write* the found error into your variable, so `target` must be a non-nil pointer to either a type implementing `error` or to any interface type. Typical usage:
```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println(pathErr.Op, pathErr.Path)
}
```
`errors.As` panics (not returns false) if `target` is nil, is not a pointer, or points to a type that neither implements `error` nor is an interface — these are programmer errors caught immediately. Note the double indirection in `&pathErr`: the outer pointer lets `As` set your `*os.PathError` variable.
**Follow-ups**: Why use a `var x *T; errors.As(err, &x)` pattern instead of a type assertion? What happens when a custom error defines its own `As(any) bool` method?
