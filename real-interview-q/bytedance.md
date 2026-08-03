# ByteDance / TikTok — Go Real Interview Questions & Deep Dives

> **Target Roles**: High-Concurrency Backend Engineer, RPC & Framework Developer, Video Pipeline Infrastructure Engineer.

---

## 🏢 Why ByteDance Asks Go
ByteDance operates one of the largest Go microservice architectures in the world (powering TikTok/Douyin). ByteDance open-sourced **Kitex** (high-performance RPC framework) and **Netpoll** (epoll-based Go network library). ByteDance interviews focus on **epoll-based network I/O**, **high-concurrency goroutine worker pools**, **allocation elimination**, and **map bucket mechanics**.

---

## 📌 Question 1: High-Performance Fixed Memory Goroutine Pool

### Question
Standard `go func()` spawns dynamic goroutines with 2 KB stacks. Under extreme traffic spikes (e.g. 500,000 req/sec), spawning uncapped goroutines degrades GC performance. Write a fixed-capacity, reusable Goroutine Pool in Go.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Task func()

type Pool struct {
	capacity int
	taskCh   chan Task
	wg       sync.WaitGroup
}

func NewPool(capacity int) *Pool {
	p := &Pool{
		capacity: capacity,
		taskCh:   make(chan Task, capacity*2),
	}
	p.init()
	return p
}

func (p *Pool) init() {
	for i := 0; i < p.capacity; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for task := range p.taskCh {
				task()
			}
		}(i)
	}
}

func (p *Pool) Submit(t Task) {
	p.taskCh <- t
}

func (p *Pool) Release() {
	close(p.taskCh)
	p.wg.Wait()
}

func main() {
	pool := NewPool(4)

	for i := 1; i <= 10; i++ {
		jobID := i
		pool.Submit(func() {
			fmt.Printf("Task %d executed\n", jobID)
		})
	}

	time.Sleep(50 * time.Millisecond)
	pool.Release()
	fmt.Println("Goroutine pool released cleanly")
}
```

---

## 📌 Question 2: Custom Map vs `hmap` Bucket Evacuation Internals

### Question
Explain why ByteDance created custom map implementations in high-performance RPC frameworks like Kitex instead of default `map[string]any`. How does Go's runtime `hmap` handle key-value buckets during resize evacuation?

### Detailed Solution
1. **Default `hmap` Resize**:
   - Go map (`hmap`) maintains array of buckets `bmap`.
   - When load factor exceeds 6.5, Go doubles bucket count ($2 \times B$).
   - Re-allocation is incremental (evacuating 1–2 buckets per write operation to avoid STW latency spikes).
2. **Interface Boxing Cost**:
   - `map[string]any` stores values in empty interface `any` (`eface`), which forces heap allocations for scalar types.
3. **ByteDance Optimization**:
   - Uses typed generic maps or static memory buffers to avoid `eface` boxing allocations entirely.
