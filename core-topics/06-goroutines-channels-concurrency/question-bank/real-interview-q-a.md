# Topic 06 Real Interview Q&A

### Q: (Fintech trading platform) Your service merges price feeds with `for v := range ch`. The feed handler on the other end reconnects and stops sending. What happens when reading from a closed channel, and how should the merge loop detect that a feed is gone?
**Frequency**: often
**Why interviewers ask**: Closed-channel semantics are where silent data corruption (zero values read as real prices) begins.
**Answer**: Reading from a closed channel first drains any remaining buffered values, then returns immediately — forever — with the zero value of the element type and `ok = false` (`v, ok := <-ch`). Crucially, it never blocks and never panics, so a plain `range` loop just terminates silently, while a `v := <-ch` loop starts consuming zero values as if they were real data. The robust merge pattern is the comma-ok check on every receive, and nil-ing the channel after closure so `select` stops selecting it:
```go
case v, ok := <-feedA:
    if !ok {
        feedA = nil // disables this case in the select
        break
    }
    handle(v)
```
For reconnection, replace the channel wholesale (new channel per connection) rather than trying to "reopen" — closed is terminal.
**Follow-ups**: Why does nil-ing the channel work in a `select`? How would you distinguish "feed legitimately ended" from "feed died" for alerting?

### Q: (Ride-sharing dispatch system) A dispatcher goroutine shuts down and closes the job channel, but a worker pool keeps calling `send`. What happens when sending to a closed channel, and how do you restructure so shutdown never panics?
**Frequency**: often
**Why interviewers ask**: "send on closed channel" panics in production are a rite of passage; they want to see ownership discipline, not recover hacks.
**Answer**: Sending on a closed channel panics immediately (`panic: send on closed channel`) — even if the buffer still has room. There is no way to check "is it closed" safely beforehand (any check races with the close), and sprinkling `recover` around sends masks the design bug. The fix is ownership: exactly one party is responsible for closing, and no one sends after that point. Concretely: producers signal completion via `sync.WaitGroup`, and a single dedicated goroutine does `wg.Wait(); close(jobs)` — producers never call `close` themselves. Alternatively, invert control: don't close the job channel at all on shutdown; cancel a `context.Context` and have workers `select` on `ctx.Done()` alongside the receive, letting the channel be garbage-collected once unreachable.
**Follow-ups**: Why is `close` + `WaitGroup` preferred over a sentinel "poison pill" value for N workers? When is the context-cancel pattern strictly better than closing?

### Q: (Cloud infra company) pprof shows your API server accumulating 40k goroutines/hour, all parked in `runtime.gopark` under `net/http.(*conn).serve` calling your metrics flusher. Diagnose the leak and fix it.
**Frequency**: sometimes
**Why interviewers ask**: They have paged on exactly this; they want the leak-hunting reflex and the "every goroutine needs an exit path" principle.
**Answer**: The profile signature means each HTTP handler spawned a helper goroutine that never exits — typically a flusher doing `for { time.Sleep(d); flush(m) }` or receiving from a channel that closes only on process exit, so one goroutine per connection accumulates. Diagnosis path: diff two `goroutine?debug=2` dumps — the leaked group shares one stack and its parked durations ("chan receive, 43 minutes") grow linearly; `runtime.NumGoroutine()` confirms the slope. Fix: bind the goroutine's lifetime to the request/server lifecycle, e.g.:
```go
ctx, cancel := context.WithCancel(req.Context())
defer cancel() // guarantee the exit path
go flushUntil(ctx.Done(), m)
```
or share one flusher across requests (a singleton with its own `stop chan struct{}`). Guard against regressions with `go.uber.org/goleak` in integration tests and an alert on goroutine count slope in production. Note `go test -race` cannot catch this — leaks are not data races.
**Follow-ups**: Why is `time.Tick` in long-lived services a milder version of the same bug? How would you set the goleak threshold given pre-existing runtime goroutines?
