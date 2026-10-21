# Topic 07: Concurrency Patterns & The `sync` Package

## 📌 Overview
Go provides standard low-level synchronization primitives inside the `sync` and `sync/atomic` packages alongside higher-level channel concurrency patterns.

---

## 🎯 Key Concepts

### 1. `sync` Primitives
- **`sync.WaitGroup`**: Wait for a collection of goroutines to finish. Always pass `*sync.WaitGroup` by pointer!
- **`sync.Mutex` vs `sync.RWMutex`**:
  - `Mutex`: Exclusive mutual exclusion lock.
  - `RWMutex`: Multiple readers allowed concurrently (`RLock()`), single writer (`Lock()`).
- **`sync.Once`**: Guarantees a function executes **exactly once**, regardless of how many goroutines invoke `once.Do(fn)`. Perfect for lazy singletons.
- **`sync.Map`**: Thread-safe map optimized for two specific scenarios:
  1. Keys are written once and read many times.
  2. Multiple goroutines read, write, and overwrite keys for disjoint sets of keys.

### 2. Common Concurrency Patterns
1. **Worker Pool**: N worker goroutines processing jobs from a shared jobs channel and outputting to a results channel.
2. **Fan-Out / Fan-In**:
   - **Fan-Out**: Multiple goroutines reading from the same input channel to perform parallel work.
   - **Fan-In**: Multiplexing multiple input channels into a single consolidated output channel.
3. **Pipeline**: A sequence of processing stages connected by channels.

---

## ⚠️ Common Gotchas
1. **Copying Mutex / WaitGroup**: Passing `sync.Mutex` or `sync.WaitGroup` by value creates a copy of the lock state, causing catastrophic race conditions or deadlocks! Always pass by pointer. (`go vet` flags most copies.)
2. **`WaitGroup.Add` Placement**: Call `Add(n)` **before** launching the goroutines, never inside them — otherwise `Wait()` can return before all goroutines have registered.
3. **`sync.Map` Is Not a Drop-In Replacement**: It only wins for read-heavy workloads with stable or disjoint keys. For write-heavy or churning key sets, a plain `map` + `RWMutex` is usually faster and gives you type safety (see `question-bank/hard.md`).
4. **`sync.Cond` Misuse**: `Cond.Wait()` must always be called in a loop guarding the condition predicate with the lock held — spurious and missed wakeups are possible otherwise. Prefer channels for most signaling.
