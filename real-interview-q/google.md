# Google — Go Real Interview Questions & Deep Dives

> **Target Roles**: Systems Software Engineer, Infrastructure Engineer, Cloud Platform Engineer (GCP), gRPC / Distributed Systems Engineer.

---

## 🏢 Why Google Asks Go
Google created Go to build massive-scale distributed infrastructure, network services, gRPC microservices, container runtimes (Borg/Kubernetes), and build systems. Google interviews focus heavily on **runtime mechanics (`iface`, GC, GMP)**, **goroutine leak prevention**, **distributed context propagation**, and **concurrency primitives**.

---

## 📌 Question 1: The Nil Interface Trap (`iface` vs `eface` Runtime Representation)

### Question
What does the following code print? Explain the exact Go runtime data structure reason for this behavior, and show how to properly fix it.

```go
package main

import "fmt"

type CustomErr struct {
	Msg string
}

func (e *CustomErr) Error() string { return e.Msg }

func ValidateUser(id int) error {
	var err *CustomErr = nil
	if id < 0 {
		err = &CustomErr{Msg: "invalid id"}
	}
	return err
}

func main() {
	err := ValidateUser(10)
	if err != nil {
		fmt.Println("Validation failed:", err)
	} else {
		fmt.Println("Validation passed")
	}
}
```

### Detailed Solution & Explanation

#### Output
```
Validation failed: <nil>
```

#### Why This Happens (Runtime Internals)
In Go runtime, an interface variable is represented by two 16-byte C structs:
- **`iface`** (for non-empty interfaces with methods, like `error`): Contains two 8-byte pointers:
  1. `tab *itab`: Pointer to the interface table (type metadata + method dispatch pointers).
  2. `data unsafe.Pointer`: Pointer to the concrete value memory.
- An interface variable evaluates to `nil` **if and only if both `tab` AND `data` are `nil`**.

In `ValidateUser`:
1. `var err *CustomErr = nil` creates a typed pointer `*CustomErr` initialized to `nil`.
2. Returning `err` converts `*CustomErr` to the `error` interface.
3. The resulting `iface` struct contains:
   - `tab` = `*itab` for `*CustomErr` (non-nil!).
   - `data` = `nil`.
4. At `if err != nil`, Go compares `iface.tab != nil`. Since `tab` holds type information for `*CustomErr`, `err != nil` evaluates to `true` even though `data` is `nil`!

#### Fix: Return Untyped Explicit `nil`
```go
func ValidateUser(id int) error {
	if id < 0 {
		return &CustomErr{Msg: "invalid id"}
	}
	return nil // Returns true nil interface (tab = nil, data = nil)
}
```

---

## 📌 Question 2: Distributed Context Cancellation & Timeout Propagation in RPCs

### Question
Write a production-ready HTTP/RPC handler client in Go that fires $N$ concurrent backend queries with a timeout. If the context expires or any query fails, all remaining outstanding goroutines must be canceled immediately to prevent resource leakage.

### Detailed Solution & Code

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Result struct {
	BackendID int
	Data      string
	Err       error
}

func QueryBackend(ctx context.Context, backendID int, delay time.Duration) (string, error) {
	select {
	case <-time.After(delay):
		return fmt.Sprintf("data-from-backend-%d", backendID), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func QueryAllBackends(parentCtx context.Context, backends map[int]time.Duration, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel() // Ensures resource cleanup on all exit paths

	resultsCh := make(chan Result, len(backends))
	var wg sync.WaitGroup

	for id, delay := range backends {
		wg.Add(1)
		go func(bID int, d time.Duration) {
			defer wg.Done()
			data, err := QueryBackend(ctx, bID, d)
			resultsCh <- Result{BackendID: bID, Data: data, Err: err}
		}(id, delay)
	}

	// Wait for goroutines to finish in background and close channel
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var collected []string
	for res := range resultsCh {
		if res.Err != nil {
			cancel() // Signal all other outstanding goroutines to cancel immediately!
			return nil, fmt.Errorf("backend %d failed: %w", res.BackendID, res.Err)
		}
		collected = append(collected, res.Data)
	}

	return collected, nil
}

func main() {
	backends := map[int]time.Duration{
		1: 50 * time.Millisecond,
		2: 100 * time.Millisecond,
		3: 150 * time.Millisecond,
	}

	res, err := QueryAllBackends(context.Background(), backends, 200*time.Millisecond)
	fmt.Println("Success Results:", res, "Err:", err)
}
```

#### Key Engineering Takeaways
- Always call `defer cancel()` when creating derived contexts with `WithTimeout` or `WithCancel`.
- Use buffered channels matching total worker counts (`len(backends)`) so workers never block on channel send if main loop exits early.

---

## 📌 Question 3: Goroutine Leak Prevention & Detection

### Question
How do goroutine leaks happen in Go servers, how do you detect them using `pprof`, and how do you prevent them in worker channels?

### Detailed Solution
1. **Causes**:
   - Goroutine blocked trying to send to an unbuffered channel with no active receiver.
   - Goroutine blocked trying to read from a channel that is never closed or written to.
   - Goroutine stuck in infinite loop or blocked on un-timed mutex/network I/O.
2. **Detection**:
   - Import `_ "net/http/pprof"` and query `/debug/pprof/goroutine?debug=2`.
   - Use `runtime.NumGoroutine()` in metrics dashboards (Prometheus) to alert on unbounded linear growth.
3. **Prevention Pattern**:
   - Use buffered channels for async result gathering.
   - Always pass `context.Context` with timeout to network calls.

---

## 📌 Question 4: Concurrent Thread-Safe LRU Cache with TTL

### Question
Implement a thread-safe LRU (Least Recently Used) cache with key TTL (Time-To-Live) expiration in Go.

### Detailed Solution & Complete Code

```go
package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type cacheItem struct {
	key       string
	val       any
	expiresAt time.Time
}

type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	evictList *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *LRUCache) Set(key string, val any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp := time.Now().Add(ttl)

	// Update existing key
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		item := elem.Value.(*cacheItem)
		item.val = val
		item.expiresAt = exp
		return
	}

	// Evict oldest if capacity reached
	if c.evictList.Len() >= c.capacity {
		c.removeOldest()
	}

	item := &cacheItem{key: key, val: val, expiresAt: exp}
	elem := c.evictList.PushFront(item)
	c.items[key] = elem
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock() // Write lock needed because moving element in list mutates list pointers!
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	item := elem.Value.(*cacheItem)
	if time.Now().After(item.expiresAt) {
		c.removeElement(elem)
		return nil, false
	}

	c.evictList.MoveToFront(elem)
	return item.val, true
}

func (c *LRUCache) removeOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	item := elem.Value.(*cacheItem)
	delete(c.items, item.key)
}

func main() {
	cache := NewLRUCache(2)
	cache.Set("a", 100, 1*time.Second)
	cache.Set("b", 200, 1*time.Second)

	val, ok := cache.Get("a")
	fmt.Println("Get a:", val, ok) // 100 true

	cache.Set("c", 300, 1*time.Second) // Evicts "b" because "a" was recently accessed

	_, okB := cache.Get("b")
	fmt.Println("Get b (evicted):", okB) // false
}
```
