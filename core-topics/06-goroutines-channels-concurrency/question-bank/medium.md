# Topic 06 Medium Questions

### Q: Why does closing a channel twice cause a panic?
**Frequency**: often
**Why interviewers ask**: It probes whether you understand channel invariants and who owns closing.
**Answer**: In the runtime's `hchan`, `close` sets `c.closed = 1`; calling `close` again (or closing a nil channel) explicitly throws `panic: close of closed channel`. The design rationale: closure is a broadcast signal to receivers, not resource cleanup — a second close is meaningless, so silently ignoring it would hide bugs. The idiom is that only the *sender* closes; receivers never close. With multiple senders, coordinate via `sync.WaitGroup` (close from a dedicated goroutine after `wg.Wait()`) or a done channel rather than closing directly.
**Follow-ups**: Can you test whether a channel is closed before closing it? Why is that check racy anyway?

### Q: Describe the behavior of send, receive, and close on nil, closed, and open channels.
**Frequency**: often
**Why interviewers ask**: This "channel state table" is the single most quizzed channel fact and the root of many deadlock bugs.
**Answer**:
| Operation | nil channel | open channel | closed channel |
|---|---|---|---|
| send | blocks forever | blocks until receiver/buffer room | **panics** |
| receive | blocks forever | blocks until value/close | drains buffer, then zero value + `ok=false` |
| close | **panics** | succeeds, wakes receivers | **panics** |

Key consequences: a nil channel in a `select` case is never ready — a feature used to disable a case dynamically (set the channel var to nil). Blocking forever on nil is a goroutine leak, not a panic, so it passes tests silently.
**Follow-ups**: Why is the "never-ready nil case" useful in `select`? What does `close(nil)` panic message say?

### Q: If multiple `select` cases are ready, which one runs?
**Frequency**: often
**Why interviewers ask**: It catches candidates who assume priority order or source order — a real correctness bug in fairness-sensitive code.
**Answer**: The choice is pseudo-random with uniform probability — the runtime (`selectgo`) shuffles the poll order before scanning. There is no priority among ready cases; source order does not matter. Only if *no* case is ready does `default` run (or the select blocks without one). If you need priority, structure it explicitly — e.g. do a non-blocking check of the high-priority channel first, then fall into a full `select`. If you need strict fairness, one common trick is to merge streams or drain in a loop rather than relying on select's random pick.
**Follow-ups**: How would you implement a priority channel pair? What does adding `default` change about blocking behavior?

### Q: A goroutine sends on an unbuffered channel that nobody ever receives from. What happens, and how do you find it?
**Frequency**: sometimes
**Why interviewers ask**: Goroutine leaks are the most common production concurrency bug, and interviewers want a detection story, not just a definition.
**Answer**: The send blocks forever; the goroutine and everything it references stay alive — a goroutine leak. There is no panic or timeout. Detection: monitor `runtime.NumGoroutine()` (a steadily rising count is the smell), capture `pprof.Lookup("goroutine")` / the `/debug/pprof/goroutine?debug=1` endpoint to see stacks of parked goroutines — leaked ones cluster at the same `chansend`/`chanreceive` site, often with their parked duration shown. In tests, `go.uber.org/goleak` fails the test if goroutines outlive it. Prevention: always give sends an escape (`select` on `ctx.Done()`), close channels to unblock receivers, and bound the number of producers.
**Follow-ups**: Why doesn't `go test -race` catch leaks? How does context cancellation fix the blocked send?
