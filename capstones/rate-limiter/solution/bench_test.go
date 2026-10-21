package solution

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Benchmarks for the stretch goals. Run with:
//
//	go test -bench=. -benchmem ./capstones/rate-limiter/solution/
//
// The hot path on an existing key should be 0 allocs/op: the bucket/counter
// struct is allocated once per key and the map holds a pointer to it.

func BenchmarkTokenBucketAllow(b *testing.B) {
	tb := NewTokenBucket(1_000_000_000, time.Nanosecond) // refill never denies
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tb.Allow("hot-key")
	}
}

func BenchmarkTokenBucketAllowParallel(b *testing.B) {
	tb := NewTokenBucket(1_000_000_000, time.Nanosecond)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow("hot-key")
		}
	})
}

func BenchmarkSlidingWindowAllow(b *testing.B) {
	sw := NewSlidingWindow(1_000_000_000, time.Minute)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sw.Allow("hot-key")
	}
}

func BenchmarkSlidingWindowAllowParallel(b *testing.B) {
	sw := NewSlidingWindow(1_000_000_000, time.Minute)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sw.Allow("hot-key")
		}
	})
}

// The two limiters below are deliberately minimal single-key fixed-window
// counters: same semantics, one guarded by sync.Mutex, one built from
// atomics only. They exist to compare the hot-path cost of the two
// synchronization strategies under BenchmarkFixedWindow*.

type mutexWindow struct {
	mu    sync.Mutex
	start time.Time
	count int64
	limit int64
	dur   time.Duration
}

func (w *mutexWindow) Allow() bool {
	now := time.Now()
	w.mu.Lock()
	defer w.mu.Unlock()
	if now.Sub(w.start) >= w.dur {
		w.start, w.count = now, 0
	}
	if w.count >= w.limit {
		return false
	}
	w.count++
	return true
}

type atomicWindow struct {
	start atomic.Int64 // unix nanoseconds of the current window start
	count atomic.Int64
	limit int64
	dur   int64 // window in nanoseconds
}

// Allow is lock-free but approximate under contention: a window roll can
// clobber a concurrent in-flight increment, so a busy boundary may allow a
// few requests less than the limit. That is the price of dropping the mutex;
// a rate limiter usually trades exactness at roll-over for throughput.
func (w *atomicWindow) Allow() bool {
	now := time.Now().UnixNano()
	for {
		start := w.start.Load()
		if now-start >= w.dur {
			if w.start.CompareAndSwap(start, now) {
				w.count.Store(0)
			}
			continue
		}
		count := w.count.Load()
		if count >= w.limit {
			return false
		}
		if w.count.CompareAndSwap(count, count+1) {
			return true
		}
	}
}

func BenchmarkFixedWindowMutex(b *testing.B) {
	w := &mutexWindow{start: time.Now(), limit: 1_000_000_000, dur: time.Minute}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Allow()
		}
	})
}

func BenchmarkFixedWindowAtomic(b *testing.B) {
	w := &atomicWindow{limit: 1_000_000_000, dur: int64(time.Minute)}
	w.start.Store(time.Now().UnixNano())
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Allow()
		}
	})
}
