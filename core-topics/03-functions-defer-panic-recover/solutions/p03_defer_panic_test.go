package solutions

import (
	"testing"
)

func TestSafeExecute(t *testing.T) {
	panickingFunc := func() {
		panic("boom")
	}

	err := SafeExecute(panickingFunc)
	if err == nil || err.Error() != "recovered panic: boom" {
		t.Fatalf("Expected recovered panic error, got %v", err)
	}
}

func TestRecoverReturnsNilWithoutPanic(t *testing.T) {
	// recover() called when there is no panic returns nil.
	if r := recover(); r != nil {
		t.Fatalf("Expected recover() to return nil when there is no panic, got %v", r)
	}
}

func TestIncrementWithDefer(t *testing.T) {
	// val = 5 -> val * 2 = 10 -> defer adds 10 -> 20
	if res := IncrementWithDefer(5); res != 20 {
		t.Fatalf("Expected 20, got %d", res)
	}
}
