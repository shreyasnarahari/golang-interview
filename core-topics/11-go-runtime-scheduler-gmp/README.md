# Topic 11: Go Runtime Scheduler (GMP Model)

## 📌 Overview
The Go scheduler is a user-space **M:N scheduler** that multiplexes $M$ user goroutines onto $N$ OS threads using logical processor contexts (**P**).

---

## 🎯 Key Concepts

### 1. The GMP Architecture
- **G (Goroutine)**: Represents execution stack and instructions for a goroutine (starts at **2 KB**).
- **M (Machine)**: Represents an OS thread created and managed by the operating system.
- **P (Processor)**: Represents a logical context required to execute Go code (`GOMAXPROCS`).

```
         +-------------------+
         | Global Run Queue  |
         +-------------------+
                   |
     +-------------+-------------+
     |                           |
+----+----+                 +----+----+
|    P    |                 |    P    |
| LocalQ  |                 | LocalQ  |
+----+----+                 +----+----+
     |                           |
+----+----+                 +----+----+
|    M    |                 |    M    |
+---------+                 +---------+
```

### 2. Scheduler Algorithms
1. **Work Stealing**: When a Processor **P** finishes its local queue, it steals half the goroutines from a random target **P**'s queue.
2. **Network Poller Handoff**: Goroutines performing I/O register with OS netpoller (`epoll`/`kqueue`), freeing **M** to execute other work.
3. **Syscall Handoff**: Blocking syscalls cause **M** to detach from **P**. `sysmon` monitors idle **P**s and binds them to new OS threads.

---

## ⚠️ Common Gotchas
1. **Preemption Stalls (Pre-1.14)**: Non-preemptible tight loops in legacy Go code could block scheduler cycles; modern Go uses `SIGURG` signal-based asynchronous preemption.
