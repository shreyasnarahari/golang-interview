# Capstone: In-Memory Key-Value Store

Build a thread-safe, in-memory key-value store in Go. This is the classic
"first 30 minutes" interview exercise: the surface API is small, but every
design decision — lock granularity, expiration strategy, snapshot format —
is a talking point.

## Requirements

Implement `Store` in `store.go` (package `kvstore`) so that `store_test.go`
passes, including under `go test -race`.

1. **Basic operations**
   - `Set(key, value string)` — store a value with no expiration.
     Overwriting a key replaces both its value and any TTL it had.
   - `Get(key string) (string, bool)` — return the value; the `bool` is
     false if the key is missing **or expired**.
   - `Delete(key string)` — remove a key; deleting a missing key is a no-op.
2. **Expiration**
   - `SetWithTTL(key, value string, ttl time.Duration)` — store a value that
     expires `ttl` from now.
   - `Get` on an expired key returns not-found.
   - `Len() int` returns the number of live keys and must **never** count
     expired ones, even if they have not been purged from memory.
3. **Persistence**
   - `Snapshot(w io.Writer) error` — serialize all live keys.
   - `Restore(r io.Reader) error` — load a snapshot, overwriting same-named
     keys. A failed Restore must leave the store unchanged.
   - Snapshots must preserve expiration: a restored key expires on the
     **original** schedule, so time that passes between Snapshot and Restore
     counts against the TTL.
4. **Concurrency**
   - All methods are safe for concurrent use.

The full contract is in the doc comments of `store.go`.

## Design notes — think about these before you code

### Locking: one RWMutex vs sharded maps

The straightforward design is one `sync.RWMutex` guarding one
`map[string]entry`. Reads (`Get`) take the read lock and run concurrently;
writes take the write lock. This is correct, simple, and fast enough for
almost everything.

A sharded store splits keys across N maps (e.g. 256), each with its own
lock, choosing a shard by hashing the key. Contention drops because
unrelated keys no longer serialize on one lock — but the win only matters
under heavy write contention, and you pay for it: `Len`, `Snapshot`, and
`Restore` must now lock **every** shard (carefully, in a fixed order, to
avoid deadlock), and `Len` no longer reflects a single instant in time.
Say all of this out loud before reaching for shards.

### Expiration: lazy vs active

- **Lazy**: `Get` (and `Len`) check the deadline and treat expired keys as
  absent. Zero background work, but expired keys that are never touched
  linger in memory forever.
- **Active**: a background sweeper periodically purges expired keys
  (`StartSweeper(interval)`, `Stop()`). Bounds memory at the cost of a
  goroutine and periodic lock traffic. `Stop` must be idempotent, safe
  without `StartSweeper`, and must not leak the goroutine.

The reference design is lazy expiration for correctness plus an optional
sweeper for memory hygiene — the store behaves identically with or without
the sweeper. To purge efficiently, keep deadlines in a min-heap
(`container/heap`) so a sweep only visits keys that have actually expired,
not the whole map. When a key is overwritten, its old heap entry becomes
*stale*: skip it when it reaches the top instead of trying to remove it
from the middle of the heap.

### Snapshots and TTLs

Serializing with `encoding/gob` (or `encoding/json`) is the easy part; the
interesting question is what timestamp goes on the wire. If you write the
*remaining* TTL at snapshot time, time spent in transit is not counted and
restored keys live too long. Writing the *absolute* deadline keeps restored
keys on the original schedule. Also decide what Restore does with a key
whose deadline passed in transit (restoring it with a past deadline and
letting lazy expiry collect it is fine), and make Restore atomic: decode
fully, then swap data in, so corrupt input leaves the store untouched.

## Stretch goals

- Implement the sharded variant (256 shards) and benchmark it against the
  single-lock version under a write-heavy workload. Be ready to explain why
  `Len`/`Snapshot` get harder.
- Add `Expire(key, ttl)` / `TTL(key)` helpers.
- Make the sweeper adaptive: sleep until the next deadline (heap top +
  `time.Timer`) instead of ticking on a fixed interval.
- Bound the store with a max-entries LRU eviction policy.

## What the interviewer is listening for

- **Correctness under race**: every map access behind a lock; no lock held
  across I/O in a way that deadlocks; `go test -race` clean.
- **The read-lock upgrade problem**: `Get` finds an expired key under
  `RLock`, wants to delete it under `Lock`. Between releasing one and
  taking the other, another goroutine may have overwritten the key — only
  delete if the entry is still the one you saw.
- **`Len` semantics**: expired-but-unpurged keys must not be counted; know
  whether your answer does a purge, a filtered count, or maintains a live
  counter.
- **Stale heap entries**: why removing arbitrary heap elements is not worth
  it and lazy deletion at the top is the standard trick.
- **Snapshot TTL choice**: remaining-TTL vs absolute-deadline, and what
  each does to keys in transit.
- **Sharding trade-off honesty**: shards reduce lock contention but break
  cheap global operations; not a free win.
- **Graceful shutdown**: how `Stop` signals the sweeper (channel close vs
  `context`), how it waits for exit, and why it must be idempotent.

## Reference solution

`solution/` contains a complete implementation (single `RWMutex`,
`container/heap` deadline tracking, gob snapshots with absolute deadlines,
optional ticker-based sweeper) with the same test suite, green under
`-race`.
