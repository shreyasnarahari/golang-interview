// Package solution is the reference implementation of the in-memory
// key-value store capstone.
package solution

import (
	"container/heap"
	"encoding/gob"
	"io"
	"sync"
	"time"
)

// entry is one key's worth of state. A zero expiresAt means the key never
// expires. Entries scheduled on the expiry heap become stale when the key
// is overwritten or deleted; the heap skips stale entries when they reach
// the top.
type entry struct {
	key       string
	value     string
	expiresAt time.Time
}

// expiryHeap is a min-heap of entries ordered by expiresAt.
type expiryHeap []*entry

func (h expiryHeap) Len() int           { return len(h) }
func (h expiryHeap) Less(i, j int) bool { return h[i].expiresAt.Before(h[j].expiresAt) }
func (h expiryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *expiryHeap) Push(x any)        { *h = append(*h, x.(*entry)) }
func (h *expiryHeap) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return e
}

// Store is a thread-safe in-memory key-value store with optional per-key
// expiration. It guards a plain map with a single sync.RWMutex and tracks
// deadlines in a min-heap so that purging expired keys costs O(log n) per
// expired key instead of a full scan.
type Store struct {
	mu       sync.RWMutex
	items    map[string]*entry
	expiries expiryHeap

	sweeping bool
	stopped  bool
	stop     chan struct{}
	done     chan struct{}
}

// NewStore returns an empty Store ready for use.
func NewStore() *Store {
	return &Store{
		items: make(map[string]*entry),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

// Set stores value under key with no expiration, overwriting any existing
// entry (including its TTL, if it had one).
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// The previous heap entry, if any, becomes stale.
	s.items[key] = &entry{key: key, value: value}
}

// Get returns the value stored under key. The second return value is false
// if the key does not exist or has expired. Expired keys are removed
// lazily.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	e, ok := s.items[key]
	if !ok {
		s.mu.RUnlock()
		return "", false
	}
	if e.expiresAt.IsZero() || time.Now().Before(e.expiresAt) {
		s.mu.RUnlock()
		return e.value, true
	}
	s.mu.RUnlock()

	// Expired: upgrade to a write lock and delete, but only if the entry we
	// saw is still the one stored (it may have been overwritten meanwhile).
	s.mu.Lock()
	if cur, ok := s.items[key]; ok && cur == e {
		delete(s.items, key)
	}
	s.mu.Unlock()
	return "", false
}

// Delete removes key from the store. Deleting a missing key is a no-op.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// SetWithTTL stores value under key and expires it after ttl. Calling Set
// or SetWithTTL on the same key replaces both the value and the deadline.
func (s *Store) SetWithTTL(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := &entry{key: key, value: value, expiresAt: time.Now().Add(ttl)}
	s.items[key] = e
	heap.Push(&s.expiries, e)
}

// Len returns the number of keys in the store. Expired keys are never
// counted, even if they have not been purged from memory yet.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(time.Now())
	return len(s.items)
}

// record is the wire format for one snapshot entry. DeadlineUnixNano is the
// key's absolute expiration deadline as a Unix timestamp in nanoseconds;
// zero means the key never expires. Storing the absolute deadline (rather
// than the remaining TTL) keeps a restored store on the original expiration
// schedule even when time passes between Snapshot and Restore.
type record struct {
	Key              string
	Value            string
	DeadlineUnixNano int64
}

// Snapshot writes a gob-encoded copy of all live keys to w. Each key's
// absolute deadline is recorded so that Restore expires keys on the original
// schedule.
func (s *Store) Snapshot(w io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.purgeExpiredLocked(time.Now())

	records := make([]record, 0, len(s.items))
	for _, e := range s.items {
		r := record{Key: e.key, Value: e.value}
		if !e.expiresAt.IsZero() {
			r.DeadlineUnixNano = e.expiresAt.UnixNano()
		}
		records = append(records, r)
	}
	return gob.NewEncoder(w).Encode(records)
}

// Restore loads a snapshot previously written by Snapshot, overwriting any
// existing entries with the same keys. A failed Restore leaves the store
// unchanged because decoding completes before the store is touched. Keys
// whose deadline passed while the snapshot was in transit are restored with
// their (past) deadline and expire on first access or sweep.
func (s *Store) Restore(r io.Reader) error {
	var records []record
	if err := gob.NewDecoder(r).Decode(&records); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range records {
		e := &entry{key: rec.Key, value: rec.Value}
		if rec.DeadlineUnixNano > 0 {
			e.expiresAt = time.Unix(0, rec.DeadlineUnixNano)
			heap.Push(&s.expiries, e)
		}
		s.items[rec.Key] = e
	}
	return nil
}

// StartSweeper launches a background goroutine that purges expired keys
// every interval. Calling it twice is a no-op.
func (s *Store) StartSweeper(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sweeping || s.stopped {
		return
	}
	s.sweeping = true
	go s.sweep(interval)
}

// Stop shuts down the background sweeper and blocks until it has exited.
// It is safe to call Stop without StartSweeper, and safe to call twice.
func (s *Store) Stop() {
	s.mu.Lock()
	started := s.sweeping
	if !s.stopped {
		s.stopped = true
		if started {
			close(s.stop)
		}
	}
	s.mu.Unlock()
	if started {
		<-s.done
	}
}

func (s *Store) sweep(interval time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.mu.Lock()
			s.purgeExpiredLocked(now)
			s.mu.Unlock()
		}
	}
}

// purgeExpiredLocked removes every entry whose deadline has passed. The
// caller must hold s.mu for writing.
func (s *Store) purgeExpiredLocked(now time.Time) {
	for len(s.expiries) > 0 {
		top := s.expiries[0]
		if top.expiresAt.After(now) {
			return
		}
		heap.Pop(&s.expiries)
		// Skip stale heap entries: the key was overwritten or deleted
		// after this entry was pushed.
		if cur, ok := s.items[top.key]; ok && cur == top {
			delete(s.items, top.key)
		}
	}
}
