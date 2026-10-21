package ratelimiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Acceptance tests for the rate-limiter capstone. They are RED by design
// until the stubs are implemented. Total sleep budget across all tests is
// kept under 200ms.

func TestTokenBucketBurstThenDeny(t *testing.T) {
	tb := NewTokenBucket(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !tb.Allow("user-1") {
			t.Fatalf("request %d should be allowed within burst of 5", i+1)
		}
	}
	if tb.Allow("user-1") {
		t.Fatal("request beyond the burst should be denied")
	}
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	tb := NewTokenBucket(5, 40*time.Millisecond)
	for i := 0; i < 5; i++ {
		tb.Allow("k")
	}
	if tb.Allow("k") {
		t.Fatal("bucket should be empty after the burst")
	}
	time.Sleep(50 * time.Millisecond) // more than one full refill period
	if !tb.Allow("k") {
		t.Fatal("bucket should have refilled after 50ms")
	}
}

func TestTokenBucketPerKeyIndependence(t *testing.T) {
	tb := NewTokenBucket(2, time.Minute)
	if !tb.Allow("a") || !tb.Allow("a") {
		t.Fatal("first two requests for a should be allowed")
	}
	if tb.Allow("a") {
		t.Fatal("a's bucket should be exhausted")
	}
	if !tb.Allow("b") {
		t.Fatal("b has its own bucket and should be allowed")
	}
}

func TestTokenBucketConcurrentHammering(t *testing.T) {
	const rate = 50
	// Refill is negligible over the test's lifetime, so exactly rate
	// requests must be allowed no matter how goroutines interleave.
	tb := NewTokenBucket(rate, time.Minute)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if tb.Allow("hot-key") {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if got := allowed.Load(); got != rate {
		t.Fatalf("expected exactly %d allowed requests, got %d", rate, got)
	}
}

func TestTokenBucketEvictIdle(t *testing.T) {
	tb := NewTokenBucket(5, time.Minute)
	tb.Allow("a")
	tb.Allow("b")
	if got := tb.ActiveKeys(); got != 2 {
		t.Fatalf("expected 2 active keys, got %d", got)
	}
	time.Sleep(10 * time.Millisecond)
	if n := tb.EvictIdle(5 * time.Millisecond); n != 2 {
		t.Fatalf("expected 2 evictions, got %d", n)
	}
	if got := tb.ActiveKeys(); got != 0 {
		t.Fatalf("expected 0 active keys after eviction, got %d", got)
	}
	tb.Allow("c")
	if n := tb.EvictIdle(time.Minute); n != 0 {
		t.Fatalf("recently used key must not be evicted, evicted %d", n)
	}
}

func TestSlidingWindowBurstThenDeny(t *testing.T) {
	sw := NewSlidingWindow(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !sw.Allow("user-1") {
			t.Fatalf("request %d should be allowed within limit of 5", i+1)
		}
	}
	if sw.Allow("user-1") {
		t.Fatal("request beyond the limit should be denied")
	}
}

func TestSlidingWindowRefillsOverTime(t *testing.T) {
	sw := NewSlidingWindow(2, 80*time.Millisecond)
	if !sw.Allow("k") || !sw.Allow("k") {
		t.Fatal("first two requests should be allowed")
	}
	if sw.Allow("k") {
		t.Fatal("limit should be reached")
	}
	// After ~1 window, the previous count is weighted by the remaining
	// overlap, so at least one more request fits.
	time.Sleep(90 * time.Millisecond)
	if !sw.Allow("k") {
		t.Fatal("window should have slid enough to allow another request")
	}
}

func TestSlidingWindowPerKeyIndependence(t *testing.T) {
	sw := NewSlidingWindow(1, time.Minute)
	if !sw.Allow("a") {
		t.Fatal("first request for a should be allowed")
	}
	if sw.Allow("a") {
		t.Fatal("a's window should be full")
	}
	if !sw.Allow("b") {
		t.Fatal("b has its own window and should be allowed")
	}
}

func TestSlidingWindowConcurrentHammering(t *testing.T) {
	const limit = 100
	// The window never rolls during the test, so exactly limit requests
	// must be allowed no matter how goroutines interleave.
	sw := NewSlidingWindow(limit, time.Minute)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				if sw.Allow("hot-key") {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if got := allowed.Load(); got != limit {
		t.Fatalf("expected exactly %d allowed requests, got %d", limit, got)
	}
}

func TestSlidingWindowEvictIdle(t *testing.T) {
	sw := NewSlidingWindow(5, time.Minute)
	sw.Allow("a")
	sw.Allow("b")
	if got := sw.ActiveKeys(); got != 2 {
		t.Fatalf("expected 2 active keys, got %d", got)
	}
	time.Sleep(10 * time.Millisecond)
	if n := sw.EvictIdle(5 * time.Millisecond); n != 2 {
		t.Fatalf("expected 2 evictions, got %d", n)
	}
	if got := sw.ActiveKeys(); got != 0 {
		t.Fatalf("expected 0 active keys after eviction, got %d", got)
	}
}

func TestSweeperEvictsIdleKeys(t *testing.T) {
	sw := NewSlidingWindow(5, 50*time.Millisecond)
	sw.Allow("x")
	if got := sw.ActiveKeys(); got != 1 {
		t.Fatalf("expected 1 active key, got %d", got)
	}
	s := NewSweeper(sw, 5*time.Millisecond, 8*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	s.Stop()
	if got := sw.ActiveKeys(); got != 0 {
		t.Fatalf("sweeper should have evicted the idle key, %d remain", got)
	}
	s.Stop() // a second Stop must not panic or deadlock
}

func TestLimitersImplementInterfaces(t *testing.T) {
	var _ Limiter = NewTokenBucket(1, time.Second)
	var _ Limiter = NewSlidingWindow(1, time.Second)
	var _ IdleEvictor = NewTokenBucket(1, time.Second)
	var _ IdleEvictor = NewSlidingWindow(1, time.Second)
}
