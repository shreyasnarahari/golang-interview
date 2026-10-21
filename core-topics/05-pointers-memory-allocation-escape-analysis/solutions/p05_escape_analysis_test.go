package solutions

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

func TestStackAllocatedStaysOnStack(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		u := StackAllocated(1, "Alice")
		if u.ID != 1 || u.Name != "Alice" {
			t.Fatalf("Unexpected stack user: %v", u)
		}
	})
	if allocs != 0 {
		t.Fatalf("StackAllocated caused %v allocations, want 0 (stack-allocated)", allocs)
	}
}

func TestHeapAllocatedEscapesToHeap(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		u := HeapAllocated(2, "Bob")
		if u.ID != 2 || u.Name != "Bob" {
			t.Fatalf("Unexpected heap user: %v", u)
		}
	})
	if allocs < 1 {
		t.Fatalf("HeapAllocated caused %v allocations, want >= 1 (escapes to heap)", allocs)
	}
}
