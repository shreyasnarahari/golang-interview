# Topic 06: Goroutines, Channels, and Concurrency Basics

## 📌 Overview
Concurrency is Go's flagship feature based on Hoare's **CSP (Communicating Sequential Processes)** model: *"Do not communicate by sharing memory; instead, share memory by communicating."*

---

## 🎯 Key Concepts

### 1. Goroutines vs OS Threads
- **OS Threads**: Managed by OS kernel, fixed 1 MB to 8 MB stack size, expensive context switches (~1–2 µs).
- **Goroutines**: Managed by Go runtime in user-space, dynamic stack starting at **2 KB**, lightweight context switches (~10–100 ns). Millions can run concurrently.

### 2. Channels: Unbuffered vs Buffered
- **Unbuffered Channel (`make(chan T)`)**: Synchronous handoff. Sender blocks until receiver receives, receiver blocks until sender sends.
- **Buffered Channel (`make(chan T, N)`)**: Asynchronous queue. Sender blocks only when buffer capacity $N$ is full; receiver blocks only when buffer is empty.

```
Unbuffered Channel (Sync Handoff)      Buffered Channel (Capacity Queue)
+-----------------------+              +-----------------------------+
| Sender ---> [ ] ---> Recv |          | Sender -> [ 1 | 2 | 3 ] -> Recv|
+-----------------------+              +-----------------------------+
```

### 3. Channel Operation Rules Matrix

| State | Read (`<-ch`) | Write (`ch <- v`) | Close (`close(ch)`) |
|---|---|---|---|
| **`nil`** | Blocks forever | Blocks forever | Panics (`close of nil channel`) |
| **Open** | Blocks if empty; else yields value | Blocks if full; else sends value | Closes channel |
| **Closed** | Yields buffered items, then zero value + `ok=false` | Panics (`send on closed channel`) | Panics (`close of closed channel`) |

### 4. `select` Statement
- Evaluates non-blocking communication cases.
- If multiple cases are ready simultaneously, Go picks one at **pseudo-random**.
- If a `default` case is present, `select` becomes non-blocking.

---

## ⚠️ Common Gotchas
1. **Goroutine Leaks**: Spawning a goroutine that blocks on a channel write/read with no corresponding reader/writer causes memory and resource leaks.
