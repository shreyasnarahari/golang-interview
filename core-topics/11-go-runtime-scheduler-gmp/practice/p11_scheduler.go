package practice

import (
	"runtime"
)

// GetSchedulerState returns current GOMAXPROCS and active goroutines count.
func GetSchedulerState() (int, int) {
	procs := runtime.GOMAXPROCS(0)
	goroutines := runtime.NumGoroutine()
	return procs, goroutines
}
