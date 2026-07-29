package practice

import (
	"testing"
)

func TestSumWithGoroutines(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expected := 55
	if res := SumWithGoroutines(nums); res != expected {
		t.Fatalf("Expected sum %d, got %d", expected, res)
	}
}

func TestNonBlockingReceive(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	val, ok := NonBlockingReceive(ch)
	if !ok || val != 42 {
		t.Fatalf("Expected 42, got val=%d, ok=%v", val, ok)
	}

	// Next read should timeout non-blockingly
	_, ok2 := NonBlockingReceive(ch)
	if ok2 {
		t.Fatalf("Expected timeout false, got true")
	}
}
