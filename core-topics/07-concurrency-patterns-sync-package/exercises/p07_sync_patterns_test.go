package exercises

import (
	"sync"
	"testing"
)

func TestSafeCounter(t *testing.T) {
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}

	wg.Wait()
	if val := counter.Value(); val != 100 {
		t.Fatalf("Expected 100, got %d", val)
	}
}

func TestWorkerPool(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5}
	results := WorkerPool(3, jobs, func(x int) int {
		return x * 2
	})

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	// Result order is nondeterministic, so compare as a multiset.
	want := map[int]int{2: 1, 4: 1, 6: 1, 8: 1, 10: 1}
	for _, r := range results {
		if want[r] == 0 {
			t.Fatalf("Unexpected result %d (results: %v)", r, results)
		}
		want[r]--
	}
}
