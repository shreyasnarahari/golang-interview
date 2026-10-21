# Topic 08 Easy Questions

### Q: What is the `context.Context` interface and what are its four methods?
**Frequency**: often
**Why interviewers ask**: It is the baseline contract every Go API involving cancellation or deadlines is built on.
**Answer**: `context.Context` carries deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines. Its four methods:
```go
type Context interface {
    Deadline() (deadline time.Time, ok bool) // when work should be abandoned
    Done() <-chan struct{}                   // closed when the context is canceled
    Err() error                              // why Done was closed (Canceled or DeadlineExceeded)
    Value(key any) any                       // request-scoped value lookup
}
```
`Done` returns nil for contexts that can never be canceled (e.g. `context.Background()`), and `Err` returns nil until `Done` is closed.
**Follow-ups**: What is the difference between `context.Background()` and `context.TODO()`? Why is `Done` a receive-only channel?

### Q: Where should `context.Context` be placed in a function signature?
**Frequency**: often
**Why interviewers ask**: It checks whether the candidate knows the standard convention that keeps Go codebases consistent.
**Answer**: By Go convention, `ctx context.Context` is always the **first parameter**, named `ctx`, and is never stored in a struct. Passing it explicitly makes data flow and cancellation propagation visible; struct fields blur the line between object lifetime and request lifetime.
```go
func FetchUser(ctx context.Context, id int64) (User, error)
```
The main exception is request-scoped types like `http.Request`, which carry their own context accessible via `r.Context()`.
**Follow-ups**: Why is storing a context in a struct discouraged? Should a `main` function create the root context with `Background()` or `TODO()`?

### Q: What is the difference between `WithCancel`, `WithTimeout`, and `WithDeadline`?
**Frequency**: often
**Why interviewers ask**: Choosing the right constructor is the everyday idiom; confusion here signals little production experience.
**Answer**: All three derive a cancelable child context from a parent:
- `WithCancel(parent)` — canceled only when the returned `CancelFunc` is called (or the parent dies).
- `WithDeadline(parent, t)` — canceled at an absolute time `t`, or when canceled explicitly, whichever is first.
- `WithTimeout(parent, d)` — shorthand for `WithDeadline(parent, time.Now().Add(d))`; use for "give this operation at most d".
All return a `CancelFunc` that **must** be called (typically `defer cancel()`) to release the child from the parent and stop its timer early.
**Follow-ups**: If the parent's deadline is sooner than the one you pass to `WithDeadline`, which wins? What error does `ctx.Err()` report on a timeout vs explicit cancel?

### Q: What is the difference between `context.Background()` and `context.TODO()`?
**Frequency**: sometimes
**Why interviewers ask**: It separates candidates who copy-paste from those who understand intent and tooling.
**Answer**: Both return the same empty, never-canceled context — they differ only in documented intent. `Background()` is the root of a context tree, used in `main`, initialization, and tests. `TODO()` is a placeholder for when it is unclear which context to use or the surrounding code has not been updated to accept one yet; static analysis tools can flag `TODO()` usages as pending work.
**Follow-ups**: What does `Done()` return on a background context? Is it safe to call `Value` on it?
