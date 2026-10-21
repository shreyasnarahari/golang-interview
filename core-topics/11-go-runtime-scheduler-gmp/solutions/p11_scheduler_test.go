package solutions

import (
	"testing"
)

func TestGetSchedulerState(t *testing.T) {
	procs, goroutines := GetSchedulerState()
	if procs <= 0 {
		t.Fatalf("Expected GOMAXPROCS > 0, got %d", procs)
	}
	if goroutines <= 0 {
		t.Fatalf("Expected NumGoroutine > 0, got %d", goroutines)
	}
}
