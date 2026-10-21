package exercises

import (
	"runtime"
	"testing"
)

func TestGCPacing(t *testing.T) {
	prev := SetGCPacing(200)
	defer SetGCPacing(prev)

	alloc, numGC := GetHeapStats()
	if alloc == 0 {
		t.Fatalf("Expected non-zero heap alloc, got %d", alloc)
	}

	runtime.GC()

	_, numGC2 := GetHeapStats()
	if numGC2 <= numGC {
		t.Fatalf("Expected NumGC count to increment after runtime.GC(), got %d -> %d", numGC, numGC2)
	}
}
