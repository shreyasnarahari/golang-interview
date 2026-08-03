# Amazon / AWS — Go Real Interview Questions & Deep Dives

> **Target Roles**: Software Development Engineer (SDE II / Senior), AWS Cloud Platform Developer, Systems Infrastructure Engineer.

---

## 🏢 Why Amazon Asks Go
AWS builds container services (ECS/EKS), Firecracker orchestration tools, AWS Lambda Go runtime, and cloud microservices in Go. Amazon interviews test **Exponential Backoff with Full Jitter**, **Concurrent S3 Chunk Uploaders**, **`golang.org/x/sync/errgroup` error propagation**, and **Context Deadlines**.

---

## 📌 Question 1: Exponential Backoff with Full Jitter Retry Algorithm

### Question
Write an production-grade HTTP/AWS API client retry function in Go implementing **Exponential Backoff with Full Jitter** to mitigate thundering herd problems when AWS endpoints return `429 Too Many Requests` or `503 Service Unavailable`.

### Detailed Solution & Complete Code

```go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type RetryableFunc func(ctx context.Context) error

func ExecuteWithRetry(ctx context.Context, maxRetries int, baseDelay, maxDelay time.Duration, fn RetryableFunc) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}

		if attempt == maxRetries {
			return fmt.Errorf("exceeded max retries (%d): %w", maxRetries, err)
		}

		// Calculate Exponential Backoff with Full Jitter:
		// temp = min(maxDelay, baseDelay * 2^attempt)
		// sleep = random_between(0, temp)
		temp := float64(baseDelay * (1 << uint(attempt)))
		if temp > float64(maxDelay) {
			temp = float64(maxDelay)
		}

		jitteredSleep := time.Duration(r.Float64() * temp)

		fmt.Printf("Attempt %d failed (%v). Sleeping for %v...\n", attempt+1, err, jitteredSleep)

		select {
		case <-time.After(jitteredSleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	attempts := 0
	flakyTask := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("AWS 503 Service Unavailable")
		}
		return nil
	}

	err := ExecuteWithRetry(ctx, 4, 50*time.Millisecond, 500*time.Millisecond, flakyTask)
	fmt.Println("Final status:", err)
}
```
