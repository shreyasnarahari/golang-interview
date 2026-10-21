package solutions

import (
	"sync"
)

// SafeCounter wraps an integer counter with an RWMutex.
type SafeCounter struct {
	mu  sync.RWMutex
	val int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val++
}

func (c *SafeCounter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

// WorkerPool processes jobs concurrently using N workers.
func WorkerPool(numWorkers int, jobs []int, workerFn func(int) int) []int {
	jobCh := make(chan int, len(jobs))
	resCh := make(chan int, len(jobs))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				resCh <- workerFn(j)
			}
		}()
	}

	for _, j := range jobs {
		jobCh <- j
	}
	close(jobCh)

	wg.Wait()
	close(resCh)

	var results []int
	for res := range resCh {
		results = append(results, res)
	}
	return results
}
