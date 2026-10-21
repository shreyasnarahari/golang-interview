// Command gomaxprocs measures the same CPU-bound workload with GOMAXPROCS
// pinned to 1 versus set to NumCPU, showing how parallel execution scales.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const iterations = 200_000_000

func busy(n int64) uint64 {
	var acc uint64
	for i := int64(1); i <= n; i++ {
		acc += uint64(i*i) % 1_000_003
	}
	return acc
}

func measure(procs int, workers int) time.Duration {
	runtime.GOMAXPROCS(procs)
	var wg sync.WaitGroup
	start := time.Now()
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = busy(iterations)
		}()
	}
	wg.Wait()
	return time.Since(start)
}

func main() {
	cores := runtime.NumCPU()
	fmt.Printf("NumCPU=%d, %d workers x %d iterations each\n", cores, cores, iterations)

	single := measure(1, cores)
	multi := measure(cores, cores)

	fmt.Printf("GOMAXPROCS=1:   %v\n", single)
	fmt.Printf("GOMAXPROCS=%d: %v\n", cores, multi)
	fmt.Printf("speedup: %.2fx on %d cores\n", float64(single)/float64(multi), cores)
}
