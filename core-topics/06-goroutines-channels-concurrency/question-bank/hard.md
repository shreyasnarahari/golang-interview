# Topic 06 Hard Questions

### Q: Explain the internal `hchan` structure in the Go runtime.
**Frequency**: rare-senior
**Why interviewers ask**: It separates people who have read the runtime from people who have only read the spec.
**Answer**: `hchan` (src/runtime/chan.go) contains:
1. `qcount`: current number of items in the buffer.
2. `dataqsiz`: circular buffer capacity (0 for unbuffered).
3. `buf`: pointer to the circular array of elements.
4. `elemsize`/`elemtype`: element size and type, needed for `memmove` into slots.
5. `sendx` & `recvx`: send/receive index offsets into the circular buffer.
6. `recvq` & `sendq`: wait queues (`waitq` of `sudog` structs) of blocked receiver/sender goroutines.
7. `closed`: flag set by `close`.
8. `lock`: a single mutex protecting all of the above — channel ops are not lock-free.

For an unbuffered channel, send/receive is a direct handoff: the sender copies its value straight into the waiting receiver's `sudog` slot and wakes it, skipping `buf` entirely.
**Follow-ups**: How does `close(ch)` interact with `recvq`/`sendq`? Why does the runtime keep blocked goroutines in `sudog` wrappers instead of on a list of goroutines?

### Q: How do goroutine stacks start at 2 KB and grow, and why did Go abandon segmented stacks?
**Frequency**: rare-senior
**Why interviewers ask**: It tests whether you understand the memory model that makes millions of goroutines feasible — and its trade-offs.
**Answer**: Every goroutine starts with a 2 KB stack (`stackMin` in the runtime). Each function prologue checks whether the frame would overflow the current stack (`stackguard0`); if so, `runtime.morestack` allocates a stack roughly double the size, copies the old stack, and adjusts all interior pointers (possible because Go can find every pointer precisely via GC stack maps). Stacks also shrink during GC. The ceiling is 1 GB on 64-bit (250 MB on 32-bit) — exceeding it is a *fatal* "stack overflow" error, not a recoverable panic. Go 1.3 replaced segmented (linked) stacks with this contiguous copying design because segments suffered the "hot split" problem: a tiny function repeatedly called across a segment boundary forced constant grow/shrink cycles, wrecking performance in tight loops.
**Follow-ups**: Why can't stack copying work in a language with uncontrolled interior pointers? How would you write code that trips the 1 GB limit, and what does the failure look like?

### Q: Walk through what the runtime does for an unbuffered send `ch <- v` when a receiver is already waiting.
**Frequency**: rare-senior
**Why interviewers ask**: It exposes whether you understand the direct-copy handoff versus the buffered path.
**Answer**: In `runtime.chansend`: acquire the channel's `hchan` lock; if the channel is closed, unlock and panic (`send on closed channel`). Then check `recvq`: since a receiver is waiting, dequeue its `sudog` and copy `v` *directly into the receiver's variable* via typed memory move (`sendDirect`) — the value never touches `buf`. The receiver's `sudog` is completed and the goroutine is made runnable (`goready`); the sender continues immediately without blocking. Contrast with no receiver waiting: the sender parks itself by allocating a `sudog`, enqueuing on `sendq`, and calling `gopark` — and this is why an unbuffered receive *observing* the value establishes a happens-before edge with the send, per the memory model.
**Follow-ups**: How does the buffered case differ when `buf` has room? Where does `select` fit — what is `selectgo` doing with multiple `sudog`s?

### Q: How do goroutine leaks actually manifest in a long-running server, and how do you systematically hunt one?
**Frequency**: sometimes
**Why interviewers ask**: Leaks are a production incident pattern (memory growth, latency tail), and seniors are expected to have a playbook.
**Answer**: Each leaked goroutine holds at least its stack (from 2 KB up to whatever it grew to) plus everything reachable from its frames — a leaked handler goroutine can pin request buffers, so 100k leaked goroutines can mean gigabytes. Symptoms: monotonically rising `runtime.NumGoroutine()`, RSS growth, and eventually GC pressure even at steady request rates. Hunt procedure: (1) hit `/debug/pprof/goroutine?debug=1` (or `pprof.Lookup("goroutine").WriteTo(w, 2)`) twice, minutes apart, and diff — leaked goroutines cluster at identical stack sites with growing ages (`goroutine ... [chan receive, 9 minutes]`). (2) Trace the stack to the missing close/cancel. Common culprits: sends/receives with no `ctx.Done()` escape, `time.Tick` (unstoppable ticker), unclosed response bodies spawning read loops, and workers ranging over a channel nobody closes. In tests, `go.uber.org/goleak` snapshots goroutines before/after and fails on survivors.
**Follow-ups**: Why do leaked goroutines show a duration in the profile? What does the `goroutine profile: total N` header tell you versus active count?
