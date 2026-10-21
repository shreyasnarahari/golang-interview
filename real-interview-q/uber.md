# Uber — Company-Flavored Go Interview Questions & Deep Dives

> **Target Roles**: Backend Engineer, Distributed Systems Engineer, High-Throughput Infrastructure Engineer, Platform Architect.
>
> *These are company-flavored questions inspired by Uber's public Go engineering (4,000+ microservices, Jaeger) and candidate reports — not verbatim interview transcripts.*

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

## 📌 Question 2: Lockless Metrics — `sync/atomic` vs `sync.Mutex` Contention

> This question is covered in the Datadog question bank — see [datadog.md, Question 1: Concurrent Metric Aggregator](datadog.md), which keeps a working implementation. The Uber-flavored variant emphasizes the **contention cost comparison** rather than the registry design: a goroutine that loses a `sync.Mutex` is parked by the runtime (`gopark`) and needs a wakeup (`goready`) on unlock — on the order of ~1 µs for the park/unpark round trip — versus ~5–10 ns for an uncontended `atomic.AddInt64`. Rule of thumb: single counters and pointer swaps → `sync/atomic` (`atomic.Int64`, `atomic.Pointer[T]`); multi-field invariants that must stay consistent → `sync.Mutex`/`sync.RWMutex`.

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
