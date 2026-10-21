# Topic 11 Easy Questions

### Q: What are G, M, and P in the Go scheduler?
**Frequency**: often
**Why interviewers ask**: The GMP model is the mental framework for every Go concurrency question that follows.
**Answer**: **G** is a goroutine — a unit of work holding a stack, instruction pointer, and register state; **M** is a machine, an OS thread that actually executes code; **P** is a logical processor — a scheduling context that holds a local run queue of Gs. An M must hold a P to run Go code, so at most `GOMAXPROCS` goroutines execute simultaneously. The model decouples "how many things want to run" (Gs, unbounded) from "how many can run in parallel" (Ps).
**Follow-ups**: What is the default value of `GOMAXPROCS`? What happens to a P when its goroutine blocks in a syscall?

### Q: What is the difference between a goroutine and an OS thread?
**Frequency**: often
**Why interviewers ask**: It tests whether the candidate understands why Go can run millions of goroutines but only thousands of threads.
**Answer**: Goroutines are user-space, multiplexed onto OS threads by the Go runtime. A goroutine starts with a small growable stack (~8 KB today) versus a fixed multi-megabyte thread stack; context-switching between goroutines is done by the scheduler in user space (a few hundred nanoseconds, no kernel trap) versus a kernel-mode thread switch; creation/teardown is cheap. Blocking I/O on goroutines goes through the netpoller instead of blocking a thread. You can have hundreds of thousands of goroutines on a handful of threads.
**Follow-ups**: What happens when a goroutine makes a truly blocking syscall like a file read? What is the default maximum number of OS threads (`SetMaxThreads`, 10,000)?

### Q: What is GOMAXPROCS and what does it default to?
**Frequency**: often
**Why interviewers ask**: A wrong mental model here leads to real production bugs, especially in containers.
**Answer**: `GOMAXPROCS` sets the number of Ps — the maximum number of goroutines executing simultaneously. It defaults to `runtime.NumCPU()` (logical CPUs). Since Go 1.25 on Linux, the default also respects the cgroup CPU bandwidth limit and is periodically re-evaluated if the quota changes, so container CPU limits now mostly "just work"; `uber-go/automaxprocs` is largely obsolete. You can override it with the `GOMAXPROCS` env var or `runtime.GOMAXPROCS(n)`, which returns the previous value.
**Follow-ups**: Why would you ever set GOMAXPROCS lower than the core count? What happened in containers before Go 1.25?
