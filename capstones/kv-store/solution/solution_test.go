package solution

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestBasicCRUD(t *testing.T) {
	s := NewStore()

	if _, ok := s.Get("missing"); ok {
		t.Fatal("Get on missing key should return ok=false")
	}

	s.Set("name", "gopher")
	if got, ok := s.Get("name"); !ok || got != "gopher" {
		t.Fatalf("Get(name) = %q, %v; want %q, true", got, ok, "gopher")
	}

	s.Set("name", "ferret")
	if got, ok := s.Get("name"); !ok || got != "ferret" {
		t.Fatalf("Get after overwrite = %q, %v; want %q, true", got, ok, "ferret")
	}

	s.Delete("name")
	if _, ok := s.Get("name"); ok {
		t.Fatal("Get after Delete should return ok=false")
	}

	s.Delete("name") // deleting a missing key must be a no-op
}

func TestLen(t *testing.T) {
	s := NewStore()
	if n := s.Len(); n != 0 {
		t.Fatalf("Len on empty store = %d; want 0", n)
	}

	s.Set("a", "1")
	s.SetWithTTL("b", "2", time.Minute)
	if n := s.Len(); n != 2 {
		t.Fatalf("Len = %d; want 2", n)
	}

	s.Delete("a")
	if n := s.Len(); n != 1 {
		t.Fatalf("Len after Delete = %d; want 1", n)
	}
}

func TestTTLExpiry(t *testing.T) {
	s := NewStore()
	s.SetWithTTL("session", "abc123", 30*time.Millisecond)

	if _, ok := s.Get("session"); !ok {
		t.Fatal("key should be visible before its TTL elapses")
	}

	time.Sleep(60 * time.Millisecond)

	if _, ok := s.Get("session"); ok {
		t.Fatal("Get on expired key should return ok=false")
	}
	if n := s.Len(); n != 0 {
		t.Fatalf("Len must never count expired keys, got %d", n)
	}
}

func TestSetClearsTTL(t *testing.T) {
	s := NewStore()
	s.SetWithTTL("k", "temp", 30*time.Millisecond)
	s.Set("k", "perm")

	time.Sleep(60 * time.Millisecond)

	if got, ok := s.Get("k"); !ok || got != "perm" {
		t.Fatalf("plain Set should clear any previous TTL: Get = %q, %v", got, ok)
	}
}

func TestSnapshotRestorePreservesRemainingTTL(t *testing.T) {
	src := NewStore()
	src.Set("perm", "forever")
	src.SetWithTTL("temp", "soon-gone", 120*time.Millisecond)

	var buf bytes.Buffer
	if err := src.Snapshot(&buf); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Let half the TTL elapse before restoring.
	time.Sleep(60 * time.Millisecond)

	dst := NewStore()
	if err := dst.Restore(&buf); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if got, ok := dst.Get("perm"); !ok || got != "forever" {
		t.Fatalf("restored Get(perm) = %q, %v; want %q, true", got, ok, "forever")
	}
	if got, ok := dst.Get("temp"); !ok || got != "soon-gone" {
		t.Fatalf("restored Get(temp) = %q, %v; want %q, true", got, ok, "soon-gone")
	}

	// Sleep past the ORIGINAL deadline (120ms from SetWithTTL) but less than
	// a full TTL since Restore. If Restore preserved the remaining TTL
	// (~60ms), the key is gone; if it reset the TTL or dropped it, the key
	// is still alive.
	time.Sleep(100 * time.Millisecond)
	if _, ok := dst.Get("temp"); ok {
		t.Fatal("remaining TTL not preserved: temp should expire 120ms after the original SetWithTTL")
	}
	if _, ok := dst.Get("perm"); !ok {
		t.Fatal("key without TTL must survive")
	}
}

func TestRestoreOnCorruptDataLeavesStoreIntact(t *testing.T) {
	s := NewStore()
	s.Set("k", "v")

	if err := s.Restore(bytes.NewReader([]byte("not a snapshot"))); err == nil {
		t.Fatal("Restore of corrupt data should return an error")
	}
	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Fatalf("failed Restore must not touch the store: Get = %q, %v", got, ok)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore()
	s.StartSweeper(5 * time.Millisecond)
	defer s.Stop()

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", g%4)
			for i := 0; i < 50; i++ {
				switch i % 5 {
				case 0:
					s.Set(key, "plain")
				case 1:
					s.SetWithTTL(key, "expiring", 20*time.Millisecond)
				case 2:
					s.Get(key)
				case 3:
					s.Delete(key)
				case 4:
					s.Len()
				}
			}
		}(g)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			if err := s.Snapshot(io.Discard); err != nil {
				t.Errorf("Snapshot: %v", err)
			}
		}
	}()

	wg.Wait()
}

func TestSweeperPurgesExpiredKeys(t *testing.T) {
	s := NewStore()
	s.StartSweeper(5 * time.Millisecond)
	defer s.Stop()

	s.SetWithTTL("short", "gone-soon", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	if _, ok := s.Get("short"); ok {
		t.Fatal("sweeper should have purged the expired key")
	}
	if n := s.Len(); n != 0 {
		t.Fatalf("Len = %d; want 0", n)
	}
}

func TestStopWithoutStart(t *testing.T) {
	s := NewStore()
	s.Stop() // must not panic or block
	s.Stop() // and must be safe to call twice
}
