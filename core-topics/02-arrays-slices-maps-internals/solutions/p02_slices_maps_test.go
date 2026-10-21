package solutions

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

func TestSafeAppend3Index_DoesNotMutateParent(t *testing.T) {
	orig := []int{10, 20, 30, 40, 50}
	snapshot := make([]int, len(orig))
	copy(snapshot, orig)

	appended := SafeAppend3Index(orig, 1, 3, 99)

	// The capacity-capped sub-slice must reallocate on append,
	// leaving the parent backing array completely untouched.
	for i, v := range snapshot {
		if orig[i] != v {
			t.Fatalf("Parent backing array mutated at index %d: expected %d, got %d (appended=%v)", i, v, orig[i], appended)
		}
	}

	// Mutating the appended slice must not leak into the parent either.
	appended[2] = -1
	if orig[3] != 40 {
		t.Fatalf("Append wrote into parent backing array: expected orig[3]==40, got %d", orig[3])
	}
}

func TestInvertMap(t *testing.T) {
	m := map[string]int{"one": 1, "two": 2}
	inv := InvertMap(m)
	if inv[1] != "one" || inv[2] != "two" {
		t.Fatalf("Map inversion failed: %v", inv)
	}
}
