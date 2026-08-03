# Stripe — Go Real Interview Questions & Deep Dives

> **Target Roles**: Payments Infrastructure Engineer, API Platform Engineer, Distributed Ledger Developer.

---

## 🏢 Why Stripe Asks Go
Stripe uses Go for high-availability payment processing engines, API proxies, and financial ledger services. Stripe interviews focus on **Idempotent Request Handling**, **Distributed Lock Leases**, **Transaction Rollbacks & Error Chains (`errors.Is`/`errors.As`)**, and **Context Deadline Propagation**.

---

## 📌 Question 1: Idempotency Key Middleware in Go

### Question
Write an Idempotent API Middleware in Go that prevents double-charge payment operations when network retries occur with the same `Idempotency-Key` header.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Response struct {
	StatusCode int
	Body       string
}

type IdempotencyStore struct {
	mu    sync.RWMutex
	cache map[string]*Response
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{cache: make(map[string]*Response)}
}

func (s *IdempotencyStore) ProcessPayment(idempotencyKey string, chargeAmount int) (Response, bool) {
	s.mu.Lock()
	// Check if already processed
	if resp, exists := s.cache[idempotencyKey]; exists {
		s.mu.Unlock()
		return *resp, true // Cached idempotency replay!
	}
	s.mu.Unlock()

	// Simulate payment processing execution
	resp := Response{
		StatusCode: 200,
		Body:       fmt.Sprintf("Charged $%d successfully", chargeAmount),
	}

	s.mu.Lock()
	s.cache[idempotencyKey] = &resp
	s.mu.Unlock()

	return resp, false
}

func main() {
	store := NewIdempotencyStore()
	key := "ik_test_883921"

	// Initial request
	res1, replayed1 := store.ProcessPayment(key, 100)
	fmt.Printf("Req 1: res=%+v, replayed=%v\n", res1, replayed1)

	// Retried request with identical key
	res2, replayed2 := store.ProcessPayment(key, 100)
	fmt.Printf("Req 2: res=%+v, replayed=%v\n", res2, replayed2)
}
```

---

## 📌 Question 2: Custom Financial Error Types with Wrapping & Inspection

### Question
Implement a custom payment error system in Go supporting nested error wrapping (`fmt.Errorf("%w")`), custom status codes, and `errors.As` extraction for payment decline codes.

### Detailed Solution & Code

```go
package main

import (
	"errors"
	"fmt"
)

type PaymentDeclineError struct {
	Code    string
	Message string
}

func (e *PaymentDeclineError) Error() string {
	return fmt.Sprintf("decline [%s]: %s", e.Code, e.Message)
}

func ProcessCharge(cardNum string) error {
	if cardNum == "4000_0000" {
		return fmt.Errorf("gateway error: %w", &PaymentDeclineError{
			Code:    "insufficient_funds",
			Message: "Card has insufficient funds",
		})
	}
	return nil
}

func main() {
	err := ProcessCharge("4000_0000")

	var declineErr *PaymentDeclineError
	if errors.As(err, &declineErr) {
		fmt.Printf("Handled Decline Error Code: %s\n", declineErr.Code)
	}
}
```
