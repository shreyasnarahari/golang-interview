// Package solution is the reference implementation for the rate-limiter
// capstone: thread-safe per-key rate limiters (token bucket and sliding
// window counter) with lazy and background eviction of idle keys.
package solution

import (
	"sync"
	"time"
)

// Limiter is a per-key rate limiter. Allow reports whether one request for
// key is permitted right now.
type Limiter interface {
	Allow(key string) bool
}

// IdleEvictor is implemented by limiters that can drop state for keys that
// have been idle longer than maxIdle. EvictIdle reports how many keys were
// evicted.
type IdleEvictor interface {
	EvictIdle(maxIdle time.Duration) int
}

var (
	_ Limiter     = (*TokenBucket)(nil)
	_ Limiter     = (*SlidingWindow)(nil)
	_ IdleEvictor = (*TokenBucket)(nil)
	_ IdleEvictor = (*SlidingWindow)(nil)
)

// TokenBucket is a per-key token bucket limiter.
//
// Each key gets its own bucket holding up to rate tokens (the burst size).
// Tokens refill lazily — computed from elapsed time on each access, so no
// background goroutine is required. A request consumes one token; it is
// denied when the bucket holds fewer than one token.
//
// All methods are safe for concurrent use.
type TokenBucket struct {
	mu      sync.Mutex
	rate    float64 // tokens per nanosecond
	burst   float64
	buckets map[string]*bucket
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// NewTokenBucket returns a TokenBucket where each key may burst up to rate
// requests and then refills rate tokens per interval. It panics if rate or
// per is not positive.
func NewTokenBucket(rate int, per time.Duration) *TokenBucket {
	if rate <= 0 || per <= 0 {
		panic("solution: rate and per must be positive")
	}
	return &TokenBucket{
		rate:    float64(rate) / float64(per),
		burst:   float64(rate),
		buckets: make(map[string]*bucket),
	}
}

// Allow reports whether one request for key is allowed right now, consuming
// a token from that key's bucket if so.
func (tb *TokenBucket) Allow(key string) bool {
	now := time.Now()
	tb.mu.Lock()
	defer tb.mu.Unlock()
	b, ok := tb.buckets[key]
	if !ok {
		// A key's first request sees a full bucket.
		b = &bucket{tokens: tb.burst, lastSeen: now}
		tb.buckets[key] = b
	}
	b.tokens += float64(now.Sub(b.lastSeen)) * tb.rate
	if b.tokens > tb.burst {
		b.tokens = tb.burst
	}
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// EvictIdle removes state for every key that has not been seen for longer
// than maxIdle and returns the number of keys evicted.
func (tb *TokenBucket) EvictIdle(maxIdle time.Duration) int {
	cutoff := time.Now().Add(-maxIdle)
	tb.mu.Lock()
	defer tb.mu.Unlock()
	evicted := 0
	for key, b := range tb.buckets {
		if b.lastSeen.Before(cutoff) {
			delete(tb.buckets, key)
			evicted++
		}
	}
	return evicted
}

// ActiveKeys reports how many keys currently hold state. It is primarily
// useful for observability and for verifying eviction.
func (tb *TokenBucket) ActiveKeys() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return len(tb.buckets)
}

// SlidingWindow is a per-key sliding-window counter limiter.
//
// Time is divided into fixed windows per key. The estimated request count
// over the trailing window is:
//
//	previousWindowCount * (1 - elapsed/window) + currentWindowCount
//
// This weights the previous window by how much of it still overlaps the
// sliding window — a cheap O(1)-per-key approximation of an exact log.
//
// All methods are safe for concurrent use.
type SlidingWindow struct {
	mu       sync.Mutex
	limit    float64
	window   time.Duration
	counters map[string]*windowCounter
}

type windowCounter struct {
	windowStart time.Time
	prev        float64
	curr        float64
	lastSeen    time.Time
}

// NewSlidingWindow returns a SlidingWindow that allows at most limit
// requests per key per trailing window. It panics if limit or window is not
// positive.
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	if limit <= 0 || window <= 0 {
		panic("solution: limit and window must be positive")
	}
	return &SlidingWindow{
		limit:    float64(limit),
		window:   window,
		counters: make(map[string]*windowCounter),
	}
}

// Allow reports whether one request for key is allowed right now, counting
// it against that key's window if so.
func (sw *SlidingWindow) Allow(key string) bool {
	now := time.Now()
	sw.mu.Lock()
	defer sw.mu.Unlock()
	c, ok := sw.counters[key]
	if !ok {
		c = &windowCounter{windowStart: now}
		sw.counters[key] = c
	}
	c.lastSeen = now
	elapsed := now.Sub(c.windowStart)
	switch {
	case elapsed >= 2*sw.window:
		// The previous window has slid entirely out of view.
		c.prev, c.curr = 0, 0
		c.windowStart = now
		elapsed = 0
	case elapsed >= sw.window:
		// Rotate: current becomes previous, keep window alignment.
		c.prev, c.curr = c.curr, 0
		c.windowStart = c.windowStart.Add(sw.window)
		elapsed = now.Sub(c.windowStart)
	}
	weighted := c.prev*(1-float64(elapsed)/float64(sw.window)) + c.curr
	if weighted >= sw.limit {
		return false
	}
	c.curr++
	return true
}

// EvictIdle removes state for every key that has not been seen for longer
// than maxIdle and returns the number of keys evicted.
func (sw *SlidingWindow) EvictIdle(maxIdle time.Duration) int {
	cutoff := time.Now().Add(-maxIdle)
	sw.mu.Lock()
	defer sw.mu.Unlock()
	evicted := 0
	for key, c := range sw.counters {
		if c.lastSeen.Before(cutoff) {
			delete(sw.counters, key)
			evicted++
		}
	}
	return evicted
}

// ActiveKeys reports how many keys currently hold state. It is primarily
// useful for observability and for verifying eviction.
func (sw *SlidingWindow) ActiveKeys() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return len(sw.counters)
}

// Sweeper periodically evicts idle keys from a limiter in a background
// goroutine. Call Stop to release the goroutine.
type Sweeper struct {
	done chan struct{}
	once sync.Once
	wg   sync.WaitGroup
}

// NewSweeper starts a goroutine that calls l.EvictIdle(maxIdle) every
// interval until Stop is called.
func NewSweeper(l IdleEvictor, interval, maxIdle time.Duration) *Sweeper {
	s := &Sweeper{done: make(chan struct{})}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.EvictIdle(maxIdle)
			case <-s.done:
				return
			}
		}
	}()
	return s
}

// Stop shuts the sweeper down and blocks until its goroutine has exited.
// It is safe to call Stop multiple times.
func (s *Sweeper) Stop() {
	s.once.Do(func() { close(s.done) })
	s.wg.Wait()
}
