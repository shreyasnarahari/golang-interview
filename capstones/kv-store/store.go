// Package kvstore is the exercise stub for the in-memory key-value store
// capstone. Implement every method on Store; see README.md for the spec.
package kvstore

import (
	"io"
	"time"
)

// Store is a thread-safe in-memory key-value store with optional per-key
// expiration.
//
// The zero value is not ready for use; construct a Store with NewStore.
type Store struct {
	// TODO: add the fields your implementation needs (map, mutex, ...).
}

// NewStore returns an empty Store ready for use.
func NewStore() *Store {
	panic("TODO: implement NewStore")
}

// Set stores value under key with no expiration, overwriting any existing
// entry (including its TTL, if it had one).
func (s *Store) Set(key, value string) {
	panic("TODO: implement Set")
}

// Get returns the value stored under key. The second return value is false
// if the key does not exist or has expired.
func (s *Store) Get(key string) (string, bool) {
	panic("TODO: implement Get")
}

// Delete removes key from the store. Deleting a missing key is a no-op.
func (s *Store) Delete(key string) {
	panic("TODO: implement Delete")
}

// SetWithTTL stores value under key and expires it after ttl. Calling Set
// or SetWithTTL on the same key replaces both the value and the deadline.
func (s *Store) SetWithTTL(key, value string, ttl time.Duration) {
	panic("TODO: implement SetWithTTL")
}

// Len returns the number of keys in the store. Expired keys are never
// counted, even if they have not been purged from memory yet.
func (s *Store) Len() int {
	panic("TODO: implement Len")
}

// Snapshot writes a serialized copy of all live keys to w. The format must
// preserve each key's remaining TTL so that a restored store expires keys
// on the original schedule.
func (s *Store) Snapshot(w io.Writer) error {
	panic("TODO: implement Snapshot")
}

// Restore loads a snapshot previously written by Snapshot, overwriting any
// existing entries with the same keys. A failed Restore must leave the
// store unchanged.
func (s *Store) Restore(r io.Reader) error {
	panic("TODO: implement Restore")
}

// StartSweeper launches a background goroutine that purges expired keys
// every interval. It is optional: the store must behave correctly without
// it (lazy expiration); the sweeper only bounds memory growth.
func (s *Store) StartSweeper(interval time.Duration) {
	panic("TODO: implement StartSweeper")
}

// Stop shuts down the background sweeper started by StartSweeper and blocks
// until it has exited. It is safe to call Stop without StartSweeper.
func (s *Store) Stop() {
	panic("TODO: implement Stop")
}
