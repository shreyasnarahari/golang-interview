package practice

import (
	"testing"
)

func TestEscapeAnalysisPatterns(t *testing.T) {
	u1 := StackAllocated(1, "Alice")
	if u1.ID != 1 || u1.Name != "Alice" {
		t.Fatalf("Unexpected stack user: %v", u1)
	}

	u2 := HeapAllocated(2, "Bob")
	if u2.ID != 2 || u2.Name != "Bob" {
		t.Fatalf("Unexpected heap user: %v", u2)
	}
}
