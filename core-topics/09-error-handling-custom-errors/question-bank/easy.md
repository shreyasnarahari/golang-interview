# Topic 09 Easy Questions

### Q: What formatting verb is used to wrap errors in `fmt.Errorf`?
**Frequency**: often
**Why interviewers ask**: It is the fastest way to check whether a candidate knows modern (Go 1.13+) error handling versus pre-1.13 string-mangling habits.
**Answer**: `%w`. `fmt.Errorf("open config: %w", err)` returns an error whose message includes the wrapped error's text and whose `Unwrap() error` method returns the original error, so `errors.Is`/`errors.As` can still match it. Since Go 1.20, multiple `%w` verbs in one call are allowed, producing an error with an `Unwrap() []error` method. Using `%v` instead formats the message but severs the chain.
**Follow-ups**: What is the difference between `%w` and `%v` in `fmt.Errorf`? What happens if you pass a non-error value to `%w`?

### Q: How is the `error` interface defined in Go?
**Frequency**: often
**Why interviewers ask**: It confirms the candidate understands that errors are just values implementing a tiny interface, not a special language construct.
**Answer**: It is a built-in interface with a single method:
```go
type error interface {
    Error() string
}
```
Any type implementing `Error() string` can be used as an error — this is why custom error types are structs with extra fields, and why functions conventionally return `error` as the last return value. The zero value of the interface is `nil`, which signals "no error".
**Follow-ups**: Why do functions return `error` as the last value by convention? What goes wrong if you return a typed nil pointer as an `error`?

### Q: What is a sentinel error and how do you create one?
**Frequency**: often
**Why interviewers ask**: Sentinel errors are the simplest error-taxonomy tool, and interviewers want to see whether the candidate knows when they are appropriate.
**Answer**: A sentinel error is a package-level, usually exported, error value that callers compare against, e.g. `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`. You create one with `errors.New` (or `fmt.Errorf` without `%w`):
```go
var ErrUserNotFound = errors.New("user not found")
```
Callers check it with `errors.Is(err, ErrUserNotFound)` — never by comparing `err.Error()` strings. Use sentinels when callers need to react to the condition but no extra context (fields) is required.
**Follow-ups**: Why should sentinels be declared once as package-level variables instead of created inline? When would you choose a custom error type over a sentinel?

### Q: How do you check whether a wrapped error matches a sentinel?
**Frequency**: often
**Why interviewers ask**: `errors.Is` is the everyday idiom that replaced `==` comparison once wrapping became standard.
**Answer**: Use `errors.Is(err, target)`. It walks the unwrap chain: at each level it checks `err == target`, then calls `err.Unwrap()` (supporting both `Unwrap() error` and `Unwrap() []error` from `errors.Join`) and repeats. Direct `==` only works when the error was never wrapped, so `errors.Is` is the safe default:
```go
if errors.Is(err, os.ErrNotExist) { /* handle missing file */ }
```
**Follow-ups**: Why is `==` insufficient once errors are wrapped? What is the `errors.As` counterpart used for?
