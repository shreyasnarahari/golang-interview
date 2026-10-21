# Topic 11 Real Interview Q&A

### Q: (Asked at a cloud infrastructure company) Your Go service runs in a Kubernetes pod limited to 2 CPUs on a 64-core node, and tail latency is terrible. What do you check first?
**Frequency**: often
**Why interviewers ask**: Container CPU limits vs GOMAXPROCS is one of the most common real-world scheduler bugs SREs hit.
**Answer**: Check the effective `GOMAXPROCS`. Before Go 1.25, the runtime defaulted to the **node's** 64 logical CPUs, ignoring the cgroup CPU quota — so 64 Ps fought over 2 cores of bandwidth, causing throttling stalls of up to the 100 ms CFS quota period and brutal p99 spikes. Fixes: upgrade to Go 1.25+ (the default is now cgroup-aware and re-evaluates if the limit changes), set `GOMAXPROCS` env explicitly, or use `uber-go/automaxprocs` on older toolchains. Also verify the app isn't spawning far more OS threads than expected via cgo/file I/O, and confirm with `runtime/metrics` (`/sched/gomaxprocs:threads`).
**Follow-ups**: Why not just set the pod's CPU *requests* instead of limits? How does CPU throttling interact with the scheduler's 10 ms preemption timeslice?

### Q: (Asked at a high-frequency trading firm) A goroutine busy-polls a lock-free queue with a tight loop and other goroutines on the same P stall. Explain why, and what you'd do.
**Frequency**: sometimes
**Why interviewers ask**: Low-latency teams live at the boundary where the scheduler's general-purpose fairness hurts deterministic latency.
**Answer**: On a single P the run queue is FIFO, so the spinning goroutine repeatedly preempts... in fact modern Go handles this better than the question implies: since Go 1.14 the loop is asynchronously preempted via `SIGURG` after ~10 ms, so stall is bounded — but 10 ms is an eternity for HFT. Options: (1) `runtime.Gosched()` in the loop to yield cooperatively (cheap, but still scheduler-mediated); (2) isolate the hot goroutine with `runtime.LockOSThread()` plus a dedicated P arrangement; (3) reduce `GOMAXPROCS` so the poller gets an effectively dedicated core; (4) restructure to blocking channel/cond waits and accept wake latency; (5) on Linux, pin threads with taskset/cpuset cgroups outside Go. The real answer is: don't fight the GMP scheduler on the hot path — isolate it.
**Follow-ups**: Why does `LockOSThread` not fully solve this by itself? What does `Gosched` do to the goroutine's position in the queue (it's requeued on the global run queue)?

### Q: (Asked at a database company) Walk us through work stealing in Go's scheduler and why it's designed the way it is.
**Frequency**: sometimes
**Why interviewers ask**: Database engineers must reason about load distribution and cache locality at a deep level.
**Answer**: When a P exhausts its local run queue (and `runnext`), `findRunnable()` checks the global queue, then the netpoller, then steals: it picks a **random victim P** and takes **half** the victim's runnable goroutines (rounding up, plus `runnext` as a last resort). Why random victim: deterministic order would create hot spots and convoy effects under skewed load — randomization gives probabilistically balanced distribution in O(1) attempts, retried a few times (up to 4 tries before giving up). Why half: stealing one goroutine would make the thief steal again immediately; stealing all would leave the victim starving. Local-first execution plus stealing keeps hot data cache-affine while bounded stealing bounds imbalance. Contrast with a single shared queue: no `sched.lock` contention on the fast path.
**Follow-ups**: What happens to locality when half a queue migrates to another core? Why does the scheduler check the global run queue every 61st tick even when local work exists?
