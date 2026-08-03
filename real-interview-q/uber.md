# Uber — Go Real Interview Questions & Deep Dives

> **Target Roles**: Backend Engineer, Distributed Systems Engineer, High-Throughput Infrastructure Engineer, Platform Architect.

---

## 🏢 Why Uber Asks Go
Uber is one of the world's largest production Go deployments (running over 4,000+ microservices in Go). Uber focuses heavily on **high-throughput concurrency**, **lock contention mitigation (`sync.RWMutex` vs `sync/atomic`)**, **distributed rate limiting algorithms**, and **zero-allocation memory optimization**.

---

## 📌 Question 1: Distributed Rate Limiter (Token Bucket Algorithm)

### Question
Implement a thread-safe, high-performance Token Bucket Rate Limiter in Go. Each key (e.g. User ID or Client IP) can consume up to $B$ tokens per second, refilling at rate $R$ tokens/sec.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	refillRate float64 // Tokens per second
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*TokenBucket
	capacity   float64
	refillRate float64
}

func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) Allow(clientKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[clientKey]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rl.capacity,
			refillRate: rl.refillRate,
			tokens:     rl.capacity - 1, // Consume 1 token immediately
			lastRefill: now,
		}
		rl.buckets[clientKey] = bucket
		return true
	}

	// Calculate tokens added since last refill
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * bucket.refillRate
	if bucket.tokens > bucket.capacity {
		bucket.tokens = bucket.capacity
	}
	bucket.lastRefill = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}

func main() {
	limiter := NewRateLimiter(3, 1) // Cap=3 tokens, Refill=1 token/sec

	client := "user_123"
	for i := 1; i <= 5; i++ {
		allowed := limiter.Allow(client)
		fmt.Printf("Req %d: allowed=%v\n", i, allowed)
	}
}
```

---

## 📌 Question 2: Lockless Lock Contention vs `sync/atomic` vs `sync.RWMutex`

### Question
Uber microservices handle hundreds of thousands of requests per second. Compare the performance and memory behavior of `sync.Mutex`, `sync.RWMutex`, and `sync/atomic` for an active metric counter. Write benchmarkable code demonstrating `atomic.AddInt64` vs `sync.Mutex`.

### Detailed Solution & Benchmark Code

```go
package main

import (
	"sync"
	"sync/atomic"
)

type MutexCounter struct {
	mu  sync.Mutex
	val int64
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.val++
	c.mu.Unlock()
}

func (c *MutexCounter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.val
}

type AtomicCounter struct {
	val int64
}

func (c *AtomicCounter) Inc() {
	atomic.AddInt64(&c.val, 1)
}

func (c *AtomicCounter) Value() int64 {
	return atomic.LoadInt64(&c.val)
}
```

#### Key Differences & Uber Engineering Guidelines
- **`sync.Mutex`**: Requires kernel/runtime thread parking when contention occurs (~1–2 µs context switch overhead per lock collision).
- **`sync/atomic`**: Uses hardware CAS (Compare-And-Swap) or bus lock instructions directly in CPU (~5–10 ns per operation).
- **Rule**: For single counter metrics or atomic pointer swaps (`atomic.Pointer[T]`), ALWAYS use `sync/atomic`. For multi-field struct updates that must remain consistent, use `sync.Mutex` or `sync.RWMutex`.

---

## 📌 Question 3: Worker Pool Pattern with Graceful Shutdown

### Question
Design a Worker Pool in Go that processes incoming ride dispatch jobs. The pool must support:
1. Fixed concurrency of $W$ worker goroutines.
2. Graceful shutdown via Context cancellation.
3. Flushing remaining queued jobs before exiting.

### Detailed Solution & Complete Code

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID int
}

type DispatchPool struct {
	numWorkers int
	jobCh      chan Job
	wg         sync.WaitGroup
}

func NewDispatchPool(numWorkers, queueCap int) *DispatchPool {
	return &DispatchPool{
		numWorkers: numWorkers,
		jobCh:      make(chan Job, queueCap),
	}
}

func (p *DispatchPool) Start(ctx context.Context) {
	for i := 1; i <= p.numWorkers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Drain remaining jobs before exiting
					for job := range p.jobCh {
						p.process(workerID, job)
					}
					return
				case job, ok := <-p.jobCh:
					if !ok {
						return
					}
					p.process(workerID, job)
				}
			}
		}(i)
	}
}

func (p *DispatchPool) Submit(job Job) bool {
	select {
	case p.jobCh <- job:
		return true
	default:
		return false // Queue full!
	}
}

func (p *DispatchPool) Stop() {
	close(p.jobCh)
	p.wg.Wait()
}

func (p *DispatchPool) process(workerID int, job Job) {
	fmt.Printf("Worker %d processing Job %d\n", workerID, job.ID)
	time.Sleep(10 * time.Millisecond)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewDispatchPool(3, 10)
	pool.Start(ctx)

	for i := 1; i <= 5; i++ {
		pool.Submit(Job{ID: i})
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	pool.Stop()
	fmt.Println("Dispatch pool cleanly stopped")
}
```
