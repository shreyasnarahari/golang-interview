# Microsoft / Azure — Go Real Interview Questions & Deep Dives

> **Target Roles**: Cloud SDE, Azure Kubernetes Systems Developer, Distributed Middleware Developer (Dapr).

---

## 🏢 Why Microsoft Asks Go
Microsoft uses Go for Azure Kubernetes Service (AKS), Dapr middleware runtime, Azure SDK for Go, and container orchestration plugins. Microsoft interviews focus on **Interface Abstractions**, **Table-Driven Testing**, **Context Cancellations**, and **Error Wrapping**.

---

## 📌 Question 1: Table-Driven Unit Testing with Subtests (`t.Run`) and Mock Interfaces

### Question
Write an idiomatic Go unit test using table-driven test cases and interface mocking for a user authentication service.

### Detailed Solution & Complete Code

```go
package main

import (
	"errors"
	"testing"
)

type UserStore interface {
	GetUser(id int) (string, error)
}

type MockUserStore struct {
	users map[int]string
	err   error
}

func (m *MockUserStore) GetUser(id int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if name, ok := m.users[id]; ok {
		return name, nil
	}
	return "", errors.New("user not found")
}

type AuthService struct {
	store UserStore
}

func (a *AuthService) GetWelcomeMessage(id int) (string, error) {
	name, err := a.store.GetUser(id)
	if err != nil {
		return "", err
	}
	return "Welcome, " + name, nil
}

func TestAuthService(t *testing.T) {
	mockStore := &MockUserStore{
		users: map[int]string{1: "Alice", 2: "Bob"},
	}
	service := &AuthService{store: mockStore}

	tests := []struct {
		name        string
		userID      int
		wantMsg     string
		wantErr     bool
	}{
		{"valid user 1", 1, "Welcome, Alice", false},
		{"valid user 2", 2, "Welcome, Bob", false},
		{"non-existent user", 99, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := service.GetWelcomeMessage(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWelcomeMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if msg != tt.wantMsg {
				t.Errorf("GetWelcomeMessage() = %v, want %v", msg, tt.wantMsg)
			}
		})
	}
}
```
