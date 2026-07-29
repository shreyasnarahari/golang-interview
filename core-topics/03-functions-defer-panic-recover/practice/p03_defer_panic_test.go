package practice

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

func TestIncrementWithDefer(t *testing.T) {
	// val = 5 -> val * 2 = 10 -> defer adds 10 -> 20
	if res := IncrementWithDefer(5); res != 20 {
		t.Fatalf("Expected 20, got %d", res)
	}
}
