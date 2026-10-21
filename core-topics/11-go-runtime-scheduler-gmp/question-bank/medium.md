# Topic 11 Medium Questions

### Q: How does Go preempt a goroutine stuck in a tight CPU loop?
**Frequency**: sometimes
**Why interviewers ask**: It separates candidates who have debugged "one hot loop starves my service" from those who only know channels.
**Answer**: Two mechanisms. (1) **Cooperative preemption**: at function-entry stack-growth checks the runtime can suspend a goroutine — loops with no function calls used to be uninterruptible. (2) **Asynchronous preemption (Go 1.14+)**: `sysmon` flags any goroutine running longer than 10 ms, and the runtime sends `SIGURG` to its M; the signal handler suspends the goroutine at a safe point so another G can run. So since Go 1.14, even `for {}` loops are preemptible.
**Follow-ups**: Why was `SIGURG` chosen over other signals? What problems existed before Go 1.14 (GC stop-the-world stalls, goroutine starvation)?

### Q: What happens when a goroutine makes a blocking syscall?
**Frequency**: often
**Why interviewers ask**: The syscall handoff is the core of how Go keeps parallelism high despite blocking C libraries and file I/O.
**Answer**: On entry to a potentially blocking syscall, the goroutine's M is dissociated from its P (`entersyscall`). If the syscall blocks, `sysmon` notices (after ~10 µs) and *retakes* the P, handing it to another M (spinning or newly spawned) so the P's remaining goroutines keep running. When the syscall returns, the original M tries to reacquire a P; if none is free, the goroutine is parked on the global run queue and the M goes to sleep. Note: network I/O usually doesn't hit this path at all — it uses the non-blocking netpoller instead.
**Follow-ups**: Why does heavy cgo/disk I/O inflate the OS thread count? What is the netpoller and which file descriptors does it cover?

### Q: In what order does a P look for work when it becomes idle?
**Frequency**: sometimes
**Why interviewers ask**: Reciting the order proves the candidate has actually read or reasoned about `schedule()`, not just the acronym.
**Answer**: In `runtime.schedule()` / `findRunnable()`: (1) every 61st scheduling tick, check the **global run queue** first to avoid GRQ starvation; (2) the P's `runnext` slot; (3) the P's local run queue; (4) the global run queue; (5) the **netpoller** for ready network goroutines; (6) **work stealing** — pick a random victim P and steal half its local queue. If still nothing, the M blocks (or spins briefly first).
**Follow-ups**: Why check the global queue only every 61 ticks instead of every time? Why steal half the victim's queue rather than one goroutine?

### Q: What is `sysmon` and what does it do?
**Frequency**: sometimes
**Why interviewers ask**: It reveals whether the candidate knows the runtime has a watchdog that runs without a P.
**Answer**: `sysmon` is a dedicated M that runs **without a P**, waking periodically (starting ~20 µs, backing off to ~10 ms when idle). It: retakes Ps from Ms stuck in blocking syscalls; forces preemption of goroutines running longer than 10 ms; runs the netpoller if it's been starved; forces GC if none has run for ~2 minutes; and scavenges unused heap memory back to the OS.
**Follow-ups**: Why must sysmon run without a P? What breaks if sysmon's preemption didn't exist in a server with one CPU-bound goroutine?
