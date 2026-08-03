# Cloudflare — Go Real Interview Questions & Deep Dives

> **Target Roles**: Network Systems Engineer, Edge Platform Engineer, Security Infrastructure Developer, Distributed Systems Engineer.

---

## 🏢 Why Cloudflare Asks Go
Cloudflare processes millions of requests per second across global edge data centers. Cloudflare Go interviews drill into **zero-copy I/O (`unsafe` conversions)**, **garbage collection latency minimization**, **custom byte buffer pooling (`sync.Pool`)**, and **lockless data structures**.

---

## 📌 Question 1: Zero-Copy String to Byte Slice Conversion without Memory Allocations

### Question
In high-throughput proxy servers, converting between `string` and `[]byte` via standard type conversion `[]byte(str)` allocates new heap memory and copies all underlying bytes. Write zero-allocation conversion functions for Go 1.20+ using `unsafe` and explain safety constraints.

### Detailed Solution & Code

```go
package main

import (
	"fmt"
	"unsafe"
)

// StringToBytes converts a string to a byte slice without memory allocation.
func StringToBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without memory allocation.
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func main() {
	msg := "Cloudflare Edge Proxy"
	b := StringToBytes(msg)
	fmt.Printf("Bytes: %v, string: %s\n", b, BytesToString(b))
}
```

#### Safety Warnings & Gotchas
- **Immutability Violation**: Standard strings in Go reside in read-only memory or are contractually immutable. If you mutate the bytes in `b = StringToBytes(s)` via `b[0] = 'X'`, it triggers a Segmentation Fault or Undefined Behavior!
- **Lifetime Safety**: `unsafe.Slice(unsafe.StringData(s), len(s))` retains a direct pointer to the backing array of `s`. If `s` is garbage collected while `b` is stored in a `uintptr` (instead of slice), memory will be invalidated.

---

## 📌 Question 2: Zero-Allocation Byte Buffer Recycler with `sync.Pool`

### Question
Network proxies allocate large byte buffers (`[]byte`) for every incoming TCP socket connection. Explain how to implement a zero-allocation byte buffer recycler using `sync.Pool` and handle slice capacity resets safely.

### Detailed Solution & Code

```go
package main

import (
	"bytes"
	"fmt"
	"sync"
)

type BufferPool struct {
	pool sync.Pool
}

func NewBufferPool(bufferSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

func (p *BufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset() // Resets len=0 while retaining underlying allocated capacity!
	return buf
}

func (p *BufferPool) Put(buf *bytes.Buffer) {
	// Guard against memory inflation: do not return overgrown buffers (>64KB) to pool
	if buf.Cap() > 65536 {
		return
	}
	p.pool.Put(buf)
}

func main() {
	pool := NewBufferPool(4096)

	buf := pool.Get()
	buf.WriteString("GET /index.html HTTP/1.1\r\nHost: cloudflare.com\r\n\r\n")

	fmt.Println("Processed Request Length:", buf.Len())
	pool.Put(buf)
}
```

---

## 📌 Question 3: Non-Blocking Ring Buffer Channel Multiplexer

### Question
Write a lock-free or low-lock Fan-In Event Aggregator in Go that merges multiple event channels into a single consolidated output channel without blocking fast producers.

### Detailed Solution & Code

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func FanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	multiplex := func(c <-chan T) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-c:
				if !ok {
					return
				}
				select {
				case out <- val:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch1 := make(chan string, 2)
	ch2 := make(chan string, 2)

	ch1 <- "edge-1"
	ch2 <- "edge-2"
	close(ch1)
	close(ch2)

	merged := FanIn(ctx, ch1, ch2)
	for msg := range merged {
		fmt.Println("Merged event:", msg)
	}
}
```
