# Topic 11: Go Runtime Scheduler (GMP Model)

## 📌 Overview
The Go scheduler is a user-space **M:N scheduler** that multiplexes $M$ user goroutines onto $N$ OS threads using logical processor contexts (**P**).

---

## 🎯 Key Concepts

### 1. The GMP Architecture
- **G (Goroutine)**: Represents a user goroutine. Contains stack pointer, instruction pointer, and status (e.g., `_Grunning`, `_Grunnable`, `_Gwaiting`). Starts with a minimal **2 KB** stack that grows/shrinks dynamically on contiguous memory.
- **M (Machine)**: Represents an OS thread managed by the OS kernel. Must be bound to a **P** to execute Go code.
- **P (Processor)**: Represents a logical execution context/resource. The number of **P**s equals `GOMAXPROCS` (defaults to the available logical CPU cores). Each **P** holds a local runnable goroutine queue (`runq` of size **256**).

```
+-------------------------------------------------------+
|                    GMP Architecture                   |
|                                                       |
|   Global Run Queue (GRQ)                              |
|   +----+  +----+  +----+                              |
|   | G  |  | G  |  | G  |                              |
|   +----+  +----+  +----+                              |
|                                                       |
|   +------------+              +------------+          |
|   |  P (Core0) |              |  P (Core1) |          |
|   | Local RunQ |              | Local RunQ |          |
|   | [G][G][G]  |              | [G][G]     |          |
|   +------------+              +------------+          |
|         |                           |                 |
|   +------------+              +------------+          |
|   |  M (OS Thread)            |  M (OS Thread)        |
|   +------------+              +------------+          |
+-------------------------------------------------------+
```

### 2. Goroutine Placement: `runnext` and Local Queues
- A goroutine created by `go f()` lands in the current **P**'s **`runnext`** slot, which is executed next; whatever was in `runnext` is displaced to the tail of the local `runq`. This gives freshly spawned goroutines priority but can starve the displaced one if abused in a loop.
- If the local `runq` (capacity 256) is full, **half of it is moved to the Global Run Queue (GRQ)** together with the new goroutine.

### 3. Scheduling & Work-Stealing Order
When a **P** runs out of runnable goroutines, `schedule()` searches for work in this order:
1. Its own local `runq` (and the `runnext` slot).
2. The **Global Run Queue** — checked at least every **61 scheduler ticks** so GRQ goroutines cannot starve.
3. The **network poller** for goroutines made runnable by ready I/O.
4. **Work stealing**: steal **half of the runnable goroutines** from another randomly chosen **P**'s local queue.

### 4. Network Poller & Syscall Handoff
- **Async I/O**: Goroutines performing network I/O register with the OS netpoller (`epoll`/`kqueue`/IOCP) and park in `_Gwaiting`, freeing **M** to execute other work.
- **Blocking Syscalls**: When a goroutine makes a blocking OS system call, its thread **M** detaches from **P**. A background thread (`sysmon`) notices **P** is idle while **M** is blocked, and moves **P** to another free OS thread (or spawns a new **M**) so the other goroutines on **P** continue executing.

### 5. Preemption Evolution
- **Pre-Go 1.14 (Cooperative Preemption)**: Goroutines yielded the CPU only at compiler-inserted stack-check prologues (`morestack`) on function entry. Tight CPU loops with no function calls (`for {}`) could stall the GC or block scheduler threads indefinitely.
- **Go 1.14+ (Asynchronous Preemption)**: The runtime sends `SIGURG` OS signals to threads running non-preemptible goroutines, allowing asynchronous stack unwinding and yielding at (almost) any instruction.

---

## ⚠️ Common Gotchas
1. **Preemption Stalls (Pre-1.14)**: Non-preemptible tight loops in legacy Go code could block scheduler cycles; modern Go uses `SIGURG` signal-based asynchronous preemption.
2. **`runnext` Jump-the-Queue**: Every new goroutine takes the `runnext` slot and pushes the previous occupant to the tail of the local `runq`, so a hot loop of `go f()` calls makes each fresh goroutine run ahead of everything already queued on that **P**.
3. **`GOMAXPROCS` Above `NumCPU`**: Setting `GOMAXPROCS` higher than the number of logical cores rarely speeds up CPU-bound work — it only adds scheduling overhead — but it can help workloads dominated by blocking syscalls.
