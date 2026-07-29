# Go Real Interview Questions & Gotcha Code Snippets

> **Organized by Topic & Target Company**
> **Company Legend:** 
> - **G** = Google | **U** = Uber | **M** = Meta | **AM** = Amazon | **MS** = Microsoft 
> - **K8S** = Docker / Kubernetes / CNCF Ecosystem | **HC** = HashiCorp | **CRDB** = CockroachDB | **CF** = Cloudflare

---

## 80/20 Priority & Frequency Matrix

| Priority | Core Topic | Frequency | Target Roles |
|---|---|---|---|
| 🔴 **P1 (Do First)** | Data Types, Slice/Map Internals, Structs & Interfaces, Defer/Panic | 60% of all Go screens | Software Engineer, Backend, Infrastructure |
| 🟡 **P2 (Week 2)** | Goroutines & Channels, `sync` Package, Context, Errors, Generics | 25% of senior screens | Distributed Systems, Cloud-Native, Platform |
| 🟢 **P3 (Senior)** | GMP Scheduler Model, Tri-Color GC, Escape Analysis, `pprof`, `unsafe` | 15% specialized rounds | Staff Engineer, Systems Architect, HFT |

---

## 1. Fundamentals, Slices, Maps & Interfaces

### Beginner Questions
- What is the zero value of slices, maps, channels, interfaces, and pointers in Go? **[G, AM, MS]**
  - *Answer*: All evaluate to `nil`.
- What is the difference between `new` and `make`? **[G, U, K8S]**
  - *Answer*: `new(T)` allocates zeroed memory for type `T` and returns `*T`. `make(T, args)` initializes and returns non-zero internal data structures for slices, maps, and channels only (`T`, not `*T`).

---

### Gotcha Code 1: The Nil Interface Trap
```go
// Q: What does this program print?
package main

import "fmt"

type CustomError struct{}

func (e *CustomError) Error() string { return "custom error" }

func getError() error {
	var err *CustomError = nil
	return err // Returns non-nil error interface containing a nil *CustomError pointer!
}

func main() {
	err := getError()
	if err != nil {
		fmt.Println("Error occurred:", err) // Executed!
	} else {
		fmt.Println("No error")
	}
}

// A: Prints "Error occurred: <nil>".
// Explanation: An interface in Go (`iface`) is represented as a tuple of (type, data).
// `err` contains type `*CustomError` and data `nil`. Since type != nil, `err != nil` evaluates to TRUE!
// Fix: Return explicit `nil` interface when no error occurs: `return nil`.
```

---

### Gotcha Code 2: Goroutine Loop Variable Capture
```go
// Q: What does this print in Go versions prior to Go 1.22 vs Go 1.22+?
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	values := []string{"a", "b", "c"}

	for _, v := range values {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(v)
		}()
	}
	wg.Wait()
}

// A: 
// Pre-Go 1.22: Prints "c", "c", "c" (or similar non-deterministic duplicates).
// Explanation: The variable `v` was shared across loop iterations. By the time goroutines ran, `v` held "c".
// Go 1.22+: Prints "a", "b", "c" (in non-deterministic order).
// Explanation: Go 1.22 changes loop variable scoping so each iteration creates a distinct memory location for `v`.
// Fix for older Go versions: Pass `v` as an argument to the goroutine function: `go func(val string) { ... }(v)`.
```

---

### Gotcha Code 3: Slice Append & Backing Array Mutation
```go
// Q: What does this print?
package main

import "fmt"

func main() {
	a := []int{1, 2, 3}
	b := append(a, 4)
	c := append(a, 5)

	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)
}

// A: 
// a: [1 2 3]
// b: [1 2 3 5]  <-- Surprising!
// c: [1 2 3 5]
// Explanation: `a` has length 3 and capacity 3 (or slice allocation default). When `append(a, 4)` runs,
// if capacity is sufficient or reallocation occurs, `b` shares the backing array. When `append(a, 5)` runs on `a`,
// it overwrites index 3 of the shared backing array with 5!
```

---

### Gotcha Code 4: Map Non-Deterministic Iteration & Concurrent Panic
```go
// Q: Can you safely iterate over a Go map and mutate it?
// A: In a single goroutine, yes — adding or deleting keys during iteration is allowed,
// but key evaluation order is randomized by design in Go to prevent reliance on hash bucket order.
// However, concurrent read/write access from MULTIPLE goroutines triggers an unrecoverable runtime panic:
// "fatal error: concurrent map read and map write".
// Fix: Protect map with `sync.RWMutex` or use `sync.Map`.
```

---

### Gotcha Code 5: `defer` Parameter Evaluation Timing
```go
// Q: What does this code print?
package main

import "fmt"

func printValue(x int) {
	fmt.Println("Deferred:", x)
}

func main() {
	x := 10
	defer printValue(x) // Argument `x` evaluated IMMEDIATELY at defer statement line!
	defer func() {
		fmt.Println("Closure:", x) // Variable `x` evaluated at function EXIT time!
	}()
	x = 20
}

// A: 
// Prints:
// Closure: 20
// Deferred: 10
```

---

## 2. Concurrency, Channels & Goroutines

### Intermediate Questions
- What happens when sending to, receiving from, or closing a `nil` or `closed` channel? **[G, U, K8S, CF]**

| Channel State | Read (Receive) | Write (Send) | Close |
|---|---|---|---|
| **`nil`** | Blocks forever | Blocks forever | Panics (`panic: close of nil channel`) |
| **Open** | Blocks if empty | Blocks if full | Succeeds (transitions to Closed) |
| **Closed** | Returns remaining buffered items, then zero value + `ok=false` | Panics (`panic: send on closed channel`) | Panics (`panic: close of closed channel`) |

---

### Gotcha Code 6: Deadlock on Unbuffered Channel
```go
// Q: Why does this program panic with deadlock?
package main

func main() {
	ch := make(chan int) // Unbuffered channel
	ch <- 42             // Blocks waiting for a receiver in ANOTHER goroutine!
	<-ch
}

// A: Fatal error: all goroutines are asleep - deadlock!
// Unbuffered channels require a simultaneous sender and receiver (synchronous handoff).
// Fix: Spawn a goroutine to send, or use a buffered channel `make(chan int, 1)`.
```

---

## 3. Go Runtime Internals & Performance

### Senior & Staff Level Questions
- Explain the **GMP Model** of the Go runtime scheduler. What are G, M, and P? **[G, U, CRDB, CF]**
  - *Answer*: See `high-performance-go.md` for full breakdown. **G** is Goroutine, **M** is OS Thread, **P** is Processor execution context.
- How does Go's **Tri-Color Concurrent Mark-Sweep Garbage Collector** minimize Stop-The-World (STW) pauses? **[G, HC, K8S]**
  - *Answer*: Uses a Hybrid Write Barrier (Dijkstra + Yuasa) allowing mutator goroutines to run concurrently during marking phase without losing reachable objects.
- How do you tune Go garbage collector behavior in cloud containers? **[U, K8S, CF]**
  - *Answer*: Set `GOMEMLIMIT` (Go 1.19+) to match container cgroup memory limits (e.g. 90% of RAM limit) to prevent OOM kills while maintaining efficient GC pacing (`GOGC`).
