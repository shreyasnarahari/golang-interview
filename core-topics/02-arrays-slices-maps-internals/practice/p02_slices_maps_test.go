package practice

import (
	"testing"
)

func TestSafeAppend3Index(t *testing.T) {
	orig := []int{10, 20, 30, 40, 50}
	appended := SafeAppend3Index(orig, 1, 3, 99)

	if len(appended) != 3 || appended[2] != 99 {
		t.Fatalf("Expected len 3 with last element 99, got %v", appended)
	}

	// Verify original slice index 3 was NOT mutated
	if orig[3] != 40 {
		t.Fatalf("Original slice was mutated! Expected orig[3]==40, got %d", orig[3])
	}
}

func TestInvertMap(t *testing.T) {
	m := map[string]int{"one": 1, "two": 2}
	inv := InvertMap(m)
	if inv[1] != "one" || inv[2] != "two" {
		t.Fatalf("Map inversion failed: %v", inv)
	}
}
