package practice

import (
	"testing"
)

func TestGenericStack(t *testing.T) {
	intStack := &Stack[int]{}
	intStack.Push(10)
	intStack.Push(20)

	val, ok := intStack.Pop()
	if !ok || val != 20 {
		t.Fatalf("Expected 20, got val=%d, ok=%v", val, ok)
	}
}

func TestDeduplicate(t *testing.T) {
	input := []string{"apple", "banana", "apple", "cherry"}
	dedup := Deduplicate(input)

	if len(dedup) != 3 {
		t.Fatalf("Expected len 3, got %d", len(dedup))
	}
}
