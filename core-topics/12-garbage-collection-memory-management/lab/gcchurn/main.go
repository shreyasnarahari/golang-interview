// Command gcchurn generates allocation churn (many short-lived heap objects)
// so the learner can observe GC pacing under different GOGC / GOMEMLIMIT
// settings with GODEBUG=gctrace=1.
package main

import (
	"fmt"
	"runtime"
)

func main() {
	const rounds = 5
	for r := 1; r <= rounds; r++ {
		churn()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("round %d: NumGC=%d HeapAlloc=%d MiB HeapSys=%d MiB\n",
			r, m.NumGC, m.HeapAlloc>>20, m.HeapSys>>20)
	}
}

// churn allocates 100,000 short-lived 1 KiB buffers (~100 MiB of garbage).
// Only the most recent buffer stays reachable; the rest die immediately.
func churn() {
	const allocations = 100_000
	var keep []byte
	for i := 0; i < allocations; i++ {
		buf := make([]byte, 1024)
		buf[0] = byte(i)
		keep = buf
	}
	runtime.KeepAlive(keep)
}
