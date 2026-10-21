// Package ratelimiter provides thread-safe per-key rate limiters.
//
// This is the exercise (stub) package for the rate-limiter capstone.
// Every exported function and method panics with a TODO. Implement the
// bodies until the acceptance tests in limiter_test.go go green, then
// compare your work against the reference implementation in solution/.
//
// You will build two classic algorithms:
//
//   - TokenBucket: each key owns a bucket that refills lazily, on access,
//     at a fixed rate. Bursts up to the bucket capacity are allowed.
//   - SlidingWindow: each key owns a two-generation counter that
//     approximates the request count over a trailing time window.
//
// Both limiters support eviction of keys that have been idle too long,
// either lazily via EvictIdle or in the background via a Sweeper.
package ratelimiter

import "time"

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

// TokenBucket is a per-key token bucket limiter.
//
// Each key gets its own bucket holding up to rate tokens (the burst size).
// Tokens refill lazily — computed from elapsed time on each access, so no
// background goroutine is required. A request consumes one token; it is
// denied when the bucket holds fewer than one token.
//
// All methods are safe for concurrent use.
type TokenBucket struct {
	// TODO: add unexported state (mutex, refill rate, burst size, per-key
	// buckets tracking token count and last-access time).
}

// NewTokenBucket returns a TokenBucket where each key may burst up to rate
// requests and then refills rate tokens per interval. It panics if rate or
// per is not positive.
func NewTokenBucket(rate int, per time.Duration) *TokenBucket {
	panic("TODO: implement NewTokenBucket")
}

// Allow reports whether one request for key is allowed right now, consuming
// a token from that key's bucket if so.
func (tb *TokenBucket) Allow(key string) bool {
	panic("TODO: implement (*TokenBucket).Allow")
}

// EvictIdle removes state for every key that has not been seen for longer
// than maxIdle and returns the number of keys evicted.
func (tb *TokenBucket) EvictIdle(maxIdle time.Duration) int {
	panic("TODO: implement (*TokenBucket).EvictIdle")
}

// ActiveKeys reports how many keys currently hold state. It is primarily
// useful for observability and for verifying eviction.
func (tb *TokenBucket) ActiveKeys() int {
	panic("TODO: implement (*TokenBucket).ActiveKeys")
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
	// TODO: add unexported state (mutex, limit, window size, per-key
	// counters tracking window start, previous/current counts, last access).
}

// NewSlidingWindow returns a SlidingWindow that allows at most limit
// requests per key per trailing window. It panics if limit or window is not
// positive.
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	panic("TODO: implement NewSlidingWindow")
}

// Allow reports whether one request for key is allowed right now, counting
// it against that key's window if so.
func (sw *SlidingWindow) Allow(key string) bool {
	panic("TODO: implement (*SlidingWindow).Allow")
}

// EvictIdle removes state for every key that has not been seen for longer
// than maxIdle and returns the number of keys evicted.
func (sw *SlidingWindow) EvictIdle(maxIdle time.Duration) int {
	panic("TODO: implement (*SlidingWindow).EvictIdle")
}

// ActiveKeys reports how many keys currently hold state. It is primarily
// useful for observability and for verifying eviction.
func (sw *SlidingWindow) ActiveKeys() int {
	panic("TODO: implement (*SlidingWindow).ActiveKeys")
}

// Sweeper periodically evicts idle keys from a limiter in a background
// goroutine. Call Stop to release the goroutine.
type Sweeper struct {
	// TODO: add unexported state (done channel, sync.Once, sync.WaitGroup).
}

// NewSweeper starts a goroutine that calls l.EvictIdle(maxIdle) every
// interval until Stop is called.
func NewSweeper(l IdleEvictor, interval, maxIdle time.Duration) *Sweeper {
	panic("TODO: implement NewSweeper")
}

// Stop shuts the sweeper down and blocks until its goroutine has exited.
// It is safe to call Stop multiple times.
func (s *Sweeper) Stop() {
	panic("TODO: implement (*Sweeper).Stop")
}
