# Topic 03 Medium Questions

### Q: What does `recover()` return if there is no active panic?
**Frequency**: often
**Why interviewers ask**: It tests whether the candidate understands recover's contract, not just its happy path.
**Answer**: It returns `nil` — both when called outside any deferred function and when called in a deferred function while no panic is in progress. Because of this, a common pattern guards on the result:
```go
defer func() {
	if r := recover(); r != nil {
		log.Printf("recovered: %v", r)
	}
}()
```
Since Go 1.21, `panic(nil)` no longer makes `recover()` return `nil` (it returns a `*runtime.PanicNilError`), so this nil-check remains reliable.
**Follow-ups**: How did Go 1.21 change the semantics of `panic(nil)`?

### Q: When are a deferred call's arguments evaluated?
**Frequency**: often
**Why interviewers ask**: The output-prediction variant of this question is a staple screen for mid-level candidates.
**Answer**: Arguments (and the receiver, for deferred method calls) are evaluated **immediately when the `defer` statement executes**, not when the deferred call runs:
```go
i := 0
defer fmt.Println(i) // prints 0
i = 99
```
To see the *final* value, defer a closure instead — closures capture variables, so `defer func() { fmt.Println(i) }()` prints 99. Only the function value and its argument list are snapshotted early.
**Follow-ups**: If you `defer file.Close()`, when is the `file` receiver evaluated, and why does that matter if you reassign `file` later?

### Q: Predict the output: a function `return`s 0 but a deferred closure sets the named return to 42. What does the caller see?
**Frequency**: often
**Why interviewers ask**: Named returns plus defer closures are where most candidates' mental model of "return" breaks down.
**Answer**: The caller sees 42:
```go
func f() (r int) {
	defer func() { r = 42 }()
	return 0
}
```
Execution order: `return 0` assigns 0 to the named result `r`, then deferred functions run (the closure assigns 42), then control returns to the caller with the final value of `r`. Without the named result, `return 0` would copy the value out and the deferred closure could not affect it.
**Follow-ups**: What happens if instead of a closure you write `defer fmt.Println(r)` — which value prints?

### Q: Why doesn't calling `recover()` inside a helper invoked by the deferred function stop the panic?
**Frequency**: sometimes
**Why interviewers ask**: It exposes whether the candidate knows recover's "directly called by a deferred function" constraint, which bites people writing reusable recover helpers.
**Answer**: `recover()` only works when it is called **directly** by a function on the deferred-call stack. If the deferred function calls `helper()` and `helper()` calls `recover()`, the recover returns `nil` and the panic keeps unwinding — verified behavior: the process still crashes. The fix is to inline the recover call in the deferred function itself, or pass the recovered value onward:
```go
defer func() {
	if r := recover(); r != nil { handlePanic(r) }
}()
```
**Follow-ups**: Given this constraint, how do libraries like HTTP frameworks structure their per-request panic recovery?
