# Topic 07 Easy Questions

### Q: What is the purpose of `sync.Once`?
**Frequency**: often
**Why interviewers ask**: It is the canonical idiom for lazy initialization and reveals whether you know the basic synchronization toolbox.
**Answer**: `sync.Once` guarantees a function runs at most once, no matter how many goroutines call `Do`, and all callers block until the first execution completes — so every goroutine that returns from `Do` observes the initialized state. Typical uses: lazy singletons, one-time config loading.
```go
var once sync.Once
var db *sql.DB

func getDB() *sql.DB {
    once.Do(func() { db = openDB() })
    return db
}
```
Go 1.21 added generic wrappers `sync.OnceFunc`, `sync.OnceValue`, and `sync.OnceValues` for functions that return values.
**Follow-ups**: What happens if the function passed to `Do` panics? Why is `once.Do` usually better than a package-level `init()` for expensive setup?

### Q: What is the difference between `sync.Mutex` and `sync.RWMutex`, and when would you choose each?
**Frequency**: often
**Why interviewers ask**: Choosing the right lock is a basic performance judgment every Go developer makes.
**Answer**: A `sync.Mutex` allows exactly one holder at a time. A `sync.RWMutex` allows many concurrent readers (`RLock`/`RUnlock`) but a writer (`Lock`/`Unlock`) excludes everyone. Choose `RWMutex` when reads vastly outnumber writes and critical sections are long enough to matter; for short critical sections or write-heavy workloads a plain `Mutex` is usually faster because `RWMutex` has extra bookkeeping (an atomic reader count plus writer coordination) and readers still contend on that shared counter.
**Follow-ups**: Can a goroutine upgrade an `RLock` to a `Lock` in Go? What is the zero value of both types?

### Q: How do you use `sync.WaitGroup` to wait for a set of goroutines to finish?
**Frequency**: often
**Why interviewers ask**: It is the standard join mechanism; misuse causes the most common goroutine bugs.
**Answer**: Call `wg.Add(n)` (or `wg.Add(1)`) *before* launching each goroutine, each goroutine calls `defer wg.Done()`, and the parent calls `wg.Wait()` to block until the counter reaches zero.
```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    go func(j Job) {
        defer wg.Done()
        process(j)
    }(job)
}
wg.Wait()
```
Since Go 1.25, `wg.Go(func() { ... })` wraps the Add/go/Done pattern for you. The counter must never go negative, and a `WaitGroup` must not be copied after first use.
**Follow-ups**: What happens if `Done` is called more times than `Add`? Why must `Add` happen in the parent goroutine rather than inside the child?

### Q: What are the fan-out and fan-in concurrency patterns?
**Frequency**: often
**Why interviewers ask**: These are the vocabulary of Go pipeline design; interviewers use them to start a design conversation.
**Answer**: **Fan-out** means multiple goroutines read from the same channel to distribute work — one channel, many consumers (the core of a worker pool). **Fan-in** means multiple producer goroutines write into a single channel, or a multiplexer goroutine merges several channels into one, so a consumer reads from one place. Together with a pipeline (stages connected by channels, each stage's output being the next stage's input), they are the standard way to structure concurrent data processing in Go.
**Follow-ups**: Who should be responsible for closing the channel in a fan-in? How do you limit the number of workers in a fan-out?
