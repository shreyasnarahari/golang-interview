# Topic 03 Easy Questions

### Q: In what order do multiple `defer` statements execute?
**Frequency**: often
**Why interviewers ask**: It is the single most-quoted defer fact and instantly reveals whether the candidate has actually used defer.
**Answer**: LIFO (Last-In, First-Out). Each `defer` pushes the call onto a per-goroutine defer stack; when the surrounding function returns, deferred calls run from the most recently registered to the oldest. So in
```go
defer fmt.Println("a")
defer fmt.Println("b")
```
`b` prints before `a`.
**Follow-ups**: What happens to a `defer` registered inside a loop — when does it run?

### Q: When does a deferred function run, and what are the classic use cases?
**Frequency**: often
**Why interviewers ask**: Interviewers want candidates to reach for defer instinctively for cleanup rather than scattering close/unlock calls.
**Answer**: A deferred function runs after the surrounding function's return statement executes but before control returns to the caller — and it runs on *every* exit path, including panics. Classic uses: releasing resources (`f.Close()`, `mu.Unlock()`), timing/tracing, and modifying named return values or recovering from panics. Deferred functions do **not** run if the process exits via `os.Exit()` or `log.Fatal()`.
**Follow-ups**: Why is `defer mu.Unlock()` preferable to calling `Unlock()` manually at each return?

### Q: What is a named return value and when is it useful?
**Frequency**: often
**Why interviewers ask**: Named returns interact with defer in ways that separate users of Go from people who merely know the syntax.
**Answer**: Result parameters can be named, e.g. `func f() (n int, err error)`. They are initialized to zero values, so a bare `return` returns their current values. They are most useful for documentation and for letting a deferred function inspect or modify the results:
```go
func parse() (err error) {
	defer func() { err = errors.Join(err, cleanup()) }()
	return nil
}
```
Avoid them purely to save typing in long functions — they hurt readability there.
**Follow-ups**: What does a deferred function see if you shadow the named return with `:=`?

### Q: What does `recover()` do and where must it be called?
**Frequency**: often
**Why interviewers ask**: Recover is the only mechanism that can turn a crash into a controlled error, so its constraints are core Go knowledge.
**Answer**: `recover()` stops the unwinding of a panicking goroutine and returns the value passed to `panic()`. It only works when called **directly** by a deferred function; called anywhere else (including in a helper function called by a deferred function) it returns `nil` and the panic continues. It is the standard way servers wrap per-request handlers so one bad request cannot crash the process.
**Follow-ups**: What does `recover()` return when there is no panic at all?
