package exercises

import (
	"slices"
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

func TestGenericStackString(t *testing.T) {
	stringStack := &Stack[string]{}
	stringStack.Push("hello")
	stringStack.Push("world")

	val, ok := stringStack.Pop()
	if !ok || val != "world" {
		t.Fatalf("Expected world, got val=%q, ok=%v", val, ok)
	}

	val, ok = stringStack.Pop()
	if !ok || val != "hello" {
		t.Fatalf("Expected hello, got val=%q, ok=%v", val, ok)
	}

	val, ok = stringStack.Pop()
	if ok || val != "" {
		t.Fatalf("Expected empty pop, got val=%q, ok=%v", val, ok)
	}
}

func TestDeduplicate(t *testing.T) {
	input := []string{"apple", "banana", "apple", "cherry"}
	expected := []string{"apple", "banana", "cherry"}
	dedup := Deduplicate(input)

	if !slices.Equal(dedup, expected) {
		t.Fatalf("Expected %v, got %v", expected, dedup)
	}
}
