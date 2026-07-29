package practice

import (
	"runtime"
	"runtime/debug"
)

// SetGCPacing configures GOGC percentage and returns previous setting.
func SetGCPacing(percent int) int {
	return debug.SetGCPercent(percent)
}

// GetHeapStats returns total allocated heap bytes and GC cycle count.
func GetHeapStats() (uint64, uint32) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc, m.NumGC
}
