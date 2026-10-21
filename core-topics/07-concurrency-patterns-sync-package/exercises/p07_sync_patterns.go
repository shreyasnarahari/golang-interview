package exercises

import (
	"sync"
)

// SafeCounter wraps an integer counter with an RWMutex.
type SafeCounter struct {
	mu  sync.RWMutex
	val int
}

func (c *SafeCounter) Inc() {
	panic("TODO: implement Inc")
}

func (c *SafeCounter) Value() int {
	panic("TODO: implement Value")
}

// WorkerPool processes jobs concurrently using N workers.
func WorkerPool(numWorkers int, jobs []int, workerFn func(int) int) []int {
	panic("TODO: implement WorkerPool")
}
