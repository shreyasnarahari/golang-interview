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
1. **Goroutine Leaks**: Spawning a goroutine that blocks on a channel write/read with no corresponding reader/writer causes memory and resource leaks. Always provide an exit path (e.g., a `done` channel or `context.Context`).
2. **Closing Channels**: Only the **sender** should close a channel. Closing from the receiver side, or closing an already-closed channel, panics. Receiving from a closed channel yields the zero value with `ok=false` — always use the two-value form `v, ok := <-ch` when the channel may be closed.
3. **`nil` Channel Deadlock**: Sending to or receiving from a `nil` channel blocks forever. This is sometimes exploited deliberately (disabling a `select` case by setting it to `nil`), but a forgotten `make(chan T)` is a classic deadlock source.
