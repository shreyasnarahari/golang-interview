# Kubernetes / CNCF Ecosystem — Go Real Interview Questions & Deep Dives

> **Target Roles**: Cloud-Native Engineer, Kubernetes Controller Developer, Infrastructure Platform Engineer, Site Reliability Engineer (SRE).

---

## 🏢 Why Kubernetes & CNCF Companies Ask Go
Kubernetes, Docker, Helm, Prometheus, Etcd, and Vitess are all native Go projects. Kubernetes interviews focus heavily on **reconciliation loop mechanics**, **client-go Informers & Workqueue patterns**, **custom controller error handling & rate-limited retries**, and **cgroup memory limits (`GOMEMLIMIT`)**.

---

## 📌 Question 1: Kubernetes Reconciliation Loop & WorkQueue Pattern in Go

### Question
Write a idiomatic Kubernetes Controller reconciliation loop pattern in Go using channels, work queues, and exponential backoff retry semantics.

### Detailed Solution & Complete Code

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ReconcileRequest struct {
	Namespace string
	Name      string
	Retries   int
}

type Controller struct {
	queue   chan ReconcileRequest
	maxRet  int
	workers int
	wg      sync.WaitGroup
}

func NewController(workers, maxRetries int) *Controller {
	return &Controller{
		queue:   make(chan ReconcileRequest, 100),
		maxRet:  maxRetries,
		workers: workers,
	}
}

func (c *Controller) Enqueue(req ReconcileRequest) {
	c.queue <- req
}

func (c *Controller) Start(ctx context.Context) {
	for i := 1; i <= c.workers; i++ {
		c.wg.Add(1)
		go func(workerID int) {
			defer c.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case req, ok := <-c.queue:
					if !ok {
						return
					}
					c.processItem(ctx, workerID, req)
				}
			}
		}(i)
	}
}

func (c *Controller) processItem(ctx context.Context, workerID int, req ReconcileRequest) {
	err := c.Reconcile(ctx, req)
	if err != nil {
		if req.Retries < c.maxRet {
			req.Retries++
			backoff := time.Duration(req.Retries*50) * time.Millisecond
			fmt.Printf("Worker %d: Reconcile failed for %s/%s. Retrying (%d/%d) in %v...\n",
				workerID, req.Namespace, req.Name, req.Retries, c.maxRet, backoff)

			go func(r ReconcileRequest, d time.Duration) {
				time.Sleep(d)
				c.Enqueue(r)
			}(req, backoff)
		} else {
			fmt.Printf("Worker %d: Reconcile PERMANENT FAILURE for %s/%s after max retries!\n",
				workerID, req.Namespace, req.Name)
		}
		return
	}
	fmt.Printf("Worker %d: Successfully reconciled %s/%s\n", workerID, req.Namespace, req.Name)
}

func (c *Controller) Reconcile(ctx context.Context, req ReconcileRequest) error {
	if req.Name == "bad-pod" && req.Retries < 2 {
		return fmt.Errorf("transient API server error")
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ctrl := NewController(2, 3)
	ctrl.Start(ctx)

	ctrl.Enqueue(ReconcileRequest{Namespace: "default", Name: "nginx-pod"})
	ctrl.Enqueue(ReconcileRequest{Namespace: "default", Name: "bad-pod"})

	time.Sleep(300 * time.Millisecond)
}
```

---

## 📌 Question 2: Memory Tuning for Go Microservices in Kubernetes Pods (`GOMEMLIMIT` & `GOGC`)

### Question
Why did Go microservices running inside Kubernetes pods experience random OOM (Out-Of-Memory) kills prior to Go 1.19, and how does `GOMEMLIMIT` solve this problem?

### Detailed Solution
1. **The Problem (Pre-Go 1.19)**:
   - `GOGC` defaults to `100`, meaning GC triggers when live heap doubles.
   - If a pod has a cgroup limit of 1 GB RAM, and the live memory set reaches 600 MB, `GOGC=100` targets the next GC cycle at $600\text{ MB} \times 2 = 1.2\text{ GB}$.
   - Before the heap reaches 1.2 GB to trigger GC, the Linux kernel OOM Killer terminates the pod (`OOMKilled exit code 137`)!
2. **The Solution (`GOMEMLIMIT` in Go 1.19+)**:
   - `GOMEMLIMIT` introduces a soft memory cap (e.g. `GOMEMLIMIT=900MiB`).
   - As memory approaches 900 MiB, the Go runtime automatically increases GC frequency and pacing dynamically, keeping memory under the cgroup limit without requiring arbitrary memory ballasts.
