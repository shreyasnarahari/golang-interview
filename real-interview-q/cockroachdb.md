# CockroachDB / TiDB — Go Real Interview Questions & Deep Dives

> **Target Roles**: Database Engine Engineer, Storage Systems Developer, High-Performance Distributed Systems Engineer.

---

## 🏢 Why CockroachDB & Database Companies Ask Go
CockroachDB and PingCAP (TiDB) build enterprise distributed SQL engines in Go. Database Go interviews test **low-level binary serialization (`encoding/binary`)**, **custom memory allocators & memory pools**, **MVCC (Multi-Version Concurrency Control) storage engine locks**, and **Raft log compaction**.

---

## 📌 Question 1: Zero-Allocation Binary Key Encoder using `encoding/binary`

### Question
Write a high-performance binary encoder in Go that serializes composite primary keys `(TenantID uint64, UserID uint64, Timestamp int64)` into lexicographically sortable byte slices without heap allocations.

### Detailed Solution & Complete Code

```go
package main

import (
	"encoding/binary"
	"fmt"
)

type Key struct {
	TenantID uint64
	UserID   uint64
	Timestamp int64
}

// EncodeKey serializes Key into a 24-byte big-endian byte array for lexicographical sorting
func EncodeKey(buf []byte, k Key) []byte {
	if cap(buf) < 24 {
		buf = make([]byte, 24)
	} else {
		buf = buf[:24]
	}

	binary.BigEndian.PutUint64(buf[0:8], k.TenantID)
	binary.BigEndian.PutUint64(buf[8:16], k.UserID)
	binary.BigEndian.PutUint64(buf[16:24], uint64(k.Timestamp))

	return buf
}

func DecodeKey(buf []byte) Key {
	return Key{
		TenantID:  binary.BigEndian.Uint64(buf[0:8]),
		UserID:    binary.BigEndian.Uint64(buf[8:16]),
		Timestamp: int64(binary.BigEndian.Uint64(buf[16:24])),
	}
}

func main() {
	k := Key{TenantID: 100, UserID: 5500, Timestamp: 1690000000}
	var stackBuf [24]byte // Stack-allocated buffer prevents heap escaping!

	encoded := EncodeKey(stackBuf[:0], k)
	decoded := DecodeKey(encoded)

	fmt.Printf("Encoded bytes: %x\n", encoded)
	fmt.Printf("Decoded Key: %+v\n", decoded)
}
```

---

## 📌 Question 2: Concurrent Multi-Version Concurrency Control (MVCC) Reader/Writer

### Question
Implement a basic MVCC (Multi-Version Concurrency Control) storage layer in Go where reads at timestamp $T$ retrieve the latest version committed at or before $T$ without blocking concurrent writes.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

type Version struct {
	Timestamp int64
	Value     string
}

type MVCCStore struct {
	mu    sync.RWMutex
	data  map[string][]Version
}

func NewMVCCStore() *MVCCStore {
	return &MVCCStore{
		data: make(map[string][]Version),
	}
}

func (s *MVCCStore) Put(key string, val string, ts int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions := s.data[key]
	versions = append(versions, Version{Timestamp: ts, Value: val})
	// Keep versions sorted by timestamp
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Timestamp < versions[j].Timestamp
	})
	s.data[key] = versions
}

func (s *MVCCStore) Get(key string, readTs int64) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, ok := s.data[key]
	if !ok || len(versions) == 0 {
		return "", false
	}

	// Binary search for greatest version <= readTs
	idx := sort.Search(len(versions), func(i int) bool {
		return versions[i].Timestamp > readTs
	}) - 1

	if idx >= 0 {
		return versions[idx].Value, true
	}

	return "", false
}

func main() {
	mvcc := NewMVCCStore()
	mvcc.Put("balance", "100", 10)
	mvcc.Put("balance", "250", 20)
	mvcc.Put("balance", "500", 30)

	val15, _ := mvcc.Get("balance", 15)
	val25, _ := mvcc.Get("balance", 25)
	val35, _ := mvcc.Get("balance", 35)

	fmt.Println("Read at ts=15:", val15) // 100
	fmt.Println("Read at ts=25:", val25) // 250
	fmt.Println("Read at ts=35:", val35) // 500
}
```
