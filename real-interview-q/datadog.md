# Datadog — Company-Flavored Go Interview Questions & Deep Dives

> **Target Roles**: Telemetry Agent Engineer, Distributed Metrics Engineer, APM & Logging Infrastructure Engineer.
>
> *These are company-flavored questions inspired by Datadog's public Go stack (Datadog Agent) and candidate reports — not verbatim interview transcripts.*

---

## 🏢 Why Datadog Asks Go
Datadog's system monitoring agent (Datadog Agent 7+) and telemetry pipelines are written in Go. This file covers **high-frequency metric collection with atomic counters**, **low-contention registries (`sync.RWMutex` fast path)**, and **memory allocation monitoring (`runtime.MemStats`)**.

---

## 📌 Question 1: Concurrent Metric Aggregator (Atomic Counters + Registry)

### Question
Write a lock-free or low-contention concurrent Metric Aggregator in Go that aggregates high-frequency telemetry counters `(MetricName, Count)` from thousands of goroutines without creating map allocation lock contention.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type CounterMetric struct {
	Name  string
	value int64
}

func (c *CounterMetric) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

func (c *CounterMetric) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

type MetricsRegistry struct {
	mu      sync.RWMutex
	metrics map[string]*CounterMetric
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]*CounterMetric),
	}
}

func (r *MetricsRegistry) Counter(name string) *CounterMetric {
	// Fast Path: Read lock
	r.mu.RLock()
	m, ok := r.metrics[name]
	r.mu.RUnlock()
	if ok {
		return m
	}

	// Slow Path: Write lock
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok = r.metrics[name]
	if !ok {
		m = &CounterMetric{Name: name}
		r.metrics[name] = m
	}
	return m
}

func main() {
	reg := NewMetricsRegistry()
	var wg sync.WaitGroup

	// Simulate 100 concurrent agent worker goroutines incrementing metrics
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := reg.Counter("http.requests.total")
			for j := 0; j < 1000; j++ {
				c.Add(1)
			}
		}()
	}

	wg.Wait()
	fmt.Println("Total http.requests.total:", reg.Counter("http.requests.total").Value()) // 100,000
}
```

---

## 📌 Question 2: Memory Profiling & Allocation Tracing with `runtime.MemStats`

### Question
How do you programmatically monitor Go heap memory usage inside a long-running telemetry agent to alert on memory growth before triggering GC?

### Detailed Solution & Code

```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func MonitorMemory(interval time.Duration) {
	var m runtime.MemStats
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		runtime.ReadMemStats(&m)
		fmt.Printf("HeapAlloc: %d MB | HeapObjects: %d | NumGC: %d\n",
			m.HeapAlloc/1024/1024, m.HeapObjects, m.NumGC)
	}
}

func main() {
	go MonitorMemory(100 * time.Millisecond)

	// Simulate allocation
	var data [][]byte
	for i := 0; i < 5; i++ {
		data = append(data, make([]byte, 10*1024*1024)) // Allocate 10MB
		time.Sleep(150 * time.Millisecond)
	}
}
```
