# DoorDash — Go Real Interview Questions & Deep Dives

> **Target Roles**: Logistics Engine Developer, Real-Time Dispatch Engineer, Microservices Platform Engineer.

---

## 🏢 Why DoorDash Asks Go
DoorDash uses Go for real-time order delivery assignment, Dasher location telemetry, and dynamic ETA calculation pipelines. DoorDash interviews test **Circuit Breaker State Machines**, **Spatial Geofence Hashing**, and **Concurrent Order Lifecycle Coordination**.

---

## 📌 Question 1: Circuit Breaker State Machine Pattern in Go

### Question
Implement a Circuit Breaker state machine in Go (`Closed` $\rightarrow$ `Open` $\rightarrow$ `Half-Open`) that protects downstream delivery microservices from cascading failures when error rates exceed threshold.

### Detailed Solution & Complete Code

```go
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	return [...]string{"Closed", "Open", "Half-Open"}[s]
}

type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failureCount int
	threshold    int
	timeout      time.Duration
	lastFailure  time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) Execute(req func() error) error {
	cb.mu.Lock()

	now := time.Now()
	if cb.state == StateOpen {
		if now.Sub(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen
			fmt.Println("Circuit Breaker state transitioned to HALF-OPEN")
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is OPEN: request rejected")
		}
	}

	cb.mu.Unlock()

	err := req()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()
		if cb.failureCount >= cb.threshold {
			cb.state = StateOpen
			fmt.Println("Circuit Breaker state transitioned to OPEN!")
		}
		return err
	}

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failureCount = 0
		fmt.Println("Circuit Breaker reset to CLOSED")
	}

	return nil
}

func main() {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	failingCall := func() error { return errors.New("downstream 500 error") }

	cb.Execute(failingCall)
	cb.Execute(failingCall) // Trips circuit breaker to OPEN

	err := cb.Execute(failingCall) // Rejected!
	fmt.Println("Call result:", err)

	time.Sleep(150 * time.Millisecond) // Wait for timeout
	okCall := func() error { return nil }
	cb.Execute(okCall) // Transitions to HALF-OPEN then CLOSED
}
```
