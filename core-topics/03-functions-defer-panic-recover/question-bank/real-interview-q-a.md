# Topic 03 Real Interview Q&A

### Q: (Asked at a cloud infra company) "Can a `recover()` in a parent goroutine catch a panic from a spawned child goroutine?"
**Frequency**: often
**Why interviewers ask**: Companies running long-lived services need engineers who internalize that one unrecovered goroutine panic takes down the whole process.
**Answer**: No. Panics unwind the stack of the specific goroutine where the panic occurred; there is no cross-goroutine recovery. If any goroutine panics without its own deferred `recover()`, the runtime prints the stack traces of **all** goroutines and the entire process exits — regardless of how many other goroutines are healthy. The robust pattern is to give every spawned goroutine its own recovery boundary:
```go
go func() {
	defer func() {
		if r := recover(); r != nil { report(r) }
	}()
	work()
}()
```
Frameworks like net/http do exactly this per connection/request.
**Follow-ups**: How would you propagate a child's failure to the parent without crashing — channels, `errgroup`, or context?

### Q: (Asked at a fintech company) "We return `error` everywhere and never panic — but a dependency panicked and crashed us in prod. How do you convert panics into errors at API boundaries?"
**Frequency**: sometimes
**Why interviewers ask**: It tests the named-return + defer + recover idiom that every defensive boundary layer (handlers, plugins, RPC stubs) needs.
**Answer**: Combine a named error result with a deferred recover that writes into it:
```go
func handle(req *Request) (resp *Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic handling request: %v", r)
		}
	}()
	return process(req)
}
```
Key points interviewers probe: the result parameter **must be named** (otherwise the deferred closure cannot affect what the caller receives); since Go 1.21 `panic(nil)` yields a non-nil `*runtime.PanicNilError`, so the `r != nil` check is safe; and you should log the stack (`debug.Stack()`) before swallowing the panic, or you lose debuggability. This belongs at trust boundaries only — do not blanket-recover inside ordinary code.
**Follow-ups**: Why does this break if you change the signature to unnamed returns `(resp *Response, _ error)`? Where would you *not* want this pattern?

### Q: (Asked at a dev-tools company) "When are a deferred function's arguments evaluated — and show a bug that comes from getting it wrong."
**Frequency**: often
**Why interviewers ask**: A live-coding variant of the classic; interviewers want to watch the candidate reason about evaluation timing, not recite it.
**Answer**: Arguments are evaluated immediately when the `defer` statement executes; only the call itself is delayed. Classic bug — logging the wrong error:
```go
err := open()
defer log.Printf("cleanup, err=%v", err) // snapshots err == nil NOW
err = work() // fails, but the log prints err=<nil>
```
Fix with a closure, which evaluates `err` at call time: `defer func() { log.Printf("cleanup, err=%v", err) }()`. The same trap applies to deferred **method receivers**: `defer buf.Flush()` snapshots `buf`; if you later reassign `buf = newBuf`, the old buffer is flushed. Closures and pointer receivers are the standard escapes.
**Follow-ups**: In `defer func(i int) {...}(i)` inside a loop, why is passing `i` as an argument preferable to closing over it, even after Go 1.22's per-iteration loop variables?
