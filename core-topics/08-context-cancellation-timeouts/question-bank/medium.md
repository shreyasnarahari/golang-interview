# Topic 08 Medium Questions

### Q: Why should context keys be custom unexported types instead of strings?
**Frequency**: often
**Why interviewers ask**: Key collisions are a real bug class when multiple packages share one request context.
**Answer**: Context keys are compared by both type and value. Using plain strings (e.g. `"userID"`) lets two unrelated packages that happen to pick the same string overwrite each other's values. A package-local unexported type makes collisions impossible:
```go
type ctxKey int

const userIDKey ctxKey = iota

ctx = context.WithValue(ctx, userIDKey, id)
```
Since the type is unexported, no external package can construct a colliding key. (An exported helper like `UserFromContext(ctx)` then provides typed access.)
**Follow-ups**: What should (and should not) be stored in context values? How does `Value` lookup walk the context chain?

### Q: When is it appropriate to use `context.WithValue`?
**Frequency**: often
**Why interviewers ask**: Overusing context values as a grab-bag for parameters is a common design smell.
**Answer**: Only for **request-scoped data that transits process and API boundaries**, not for optional function parameters: trace/span IDs, auth tokens, request IDs, and similar metadata. If a value is required for the function to do its job, it should be an explicit parameter — hiding it in a context removes type safety (everything is `any`), makes dependencies invisible, and forces nil-checks. A good litmus test: would the program be just as correct if the value were absent? If not, make it a parameter.
**Follow-ups**: Why did the designers type keys and values as `any` instead of making context generic? What is the lookup cost of `ctx.Value` in a deep chain?

### Q: What does this code do, and what is the subtle bug?
```go
ctx := context.Background()
select {
case <-ctx.Done():
    fmt.Println("canceled")
case v := <-work:
    fmt.Println(v)
}
```
**Frequency**: sometimes
**Why interviewers ask**: Behavior prediction with a nil channel catches candidates who treat `Done()` as always-blocking-safe boilerplate.
**Answer**: It compiles and never prints "canceled": `Background().Done()` returns a **nil channel**, and receiving from a nil channel blocks forever, so that case is simply never selectable (a nil channel in a `select` case is effectively disabled — no deadlock, the other cases still run). The subtle risk is the reverse: a `select` that *only* waits on `ctx.Done()` of a background context deadlocks silently. This is also why `context.WithoutCancel` contexts (whose `Done()` is nil) must never drive a cancellation loop.
**Follow-ups**: How would you guard code that must work with both cancelable and non-cancelable contexts? What does `Err()` return for a background context?

### Q: What is `context.WithoutCancel` (Go 1.21) and when would you use it?
**Frequency**: sometimes
**Why interviewers ask**: It tests knowledge of the modern API that replaced the old "detach the context" hacks.
**Answer**: `WithoutCancel(parent)` returns a context that keeps the parent's **values** but is detached from its cancellation: `Done()` returns nil, `Err()` and `Deadline()` report nothing, and `context.Cause` returns nil even after the parent is canceled. The classic use case is finishing work that must outlive a canceled request — e.g. flushing metrics, audit logging, or a fire-and-forget notification after an HTTP handler's context is canceled:
```go
detached := context.WithoutCancel(ctx) // keeps trace ID, drops cancellation
ctx2, cancel := context.WithTimeout(detached, 5*time.Second)
```
Before Go 1.21 people re-created a bare `Background()` context and lost all request-scoped values.
**Follow-ups**: Why is it still a good idea to wrap the detached context in `WithTimeout`? What happens to values stored by ancestors when the parent is canceled?
