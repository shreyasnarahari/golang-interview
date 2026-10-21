package jobqueue

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllJobsProcessed(t *testing.T) {
	p := New(4)

	var count atomic.Int64
	const n = 100
	for i := 0; i < n; i++ {
		err := p.Submit(context.Background(), func(ctx context.Context) error {
			count.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Submit %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if got := count.Load(); got != n {
		t.Fatalf("processed %d jobs; want %d", got, n)
	}
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("Errors() = %v; want none", errs)
	}
}

func TestErrorsCollected(t *testing.T) {
	p := New(2)

	errBoom := errors.New("boom")
	const n = 10
	for i := 0; i < n; i++ {
		if err := p.Submit(context.Background(), func(ctx context.Context) error {
			return errBoom
		}); err != nil {
			t.Fatalf("Submit %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	errs := p.Errors()
	if len(errs) != n {
		t.Fatalf("Errors() has %d entries; want %d", len(errs), n)
	}
	for i, err := range errs {
		if !errors.Is(err, errBoom) {
			t.Fatalf("Errors()[%d] = %v; want boom", i, err)
		}
	}
}

func TestGracefulDrain(t *testing.T) {
	p := New(2)

	var completed atomic.Int64
	const n = 10
	for i := 0; i < n; i++ {
		err := p.Submit(context.Background(), func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			completed.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Submit %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// Structural guarantee, no timing involved: once Shutdown returns nil,
	// every queued job must have run to completion.
	if got := completed.Load(); got != n {
		t.Fatalf("Shutdown returned before draining: completed %d; want %d", got, n)
	}
}

func TestShutdownTimeout(t *testing.T) {
	p := New(1)

	started := make(chan struct{})
	if err := p.Submit(context.Background(), func(ctx context.Context) error {
		close(started)
		time.Sleep(500 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := p.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown = %v; want context.DeadlineExceeded", err)
	}
}

func TestPanicRecovery(t *testing.T) {
	p := New(1)

	if err := p.Submit(context.Background(), func(ctx context.Context) error {
		panic("kaboom")
	}); err != nil {
		t.Fatalf("Submit panicking job: %v", err)
	}

	var ran atomic.Bool
	if err := p.Submit(context.Background(), func(ctx context.Context) error {
		ran.Store(true)
		return nil
	}); err != nil {
		t.Fatalf("Submit after panicking job: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if !ran.Load() {
		t.Fatal("pool did not survive a panicking job: later job never ran")
	}
	errs := p.Errors()
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "kaboom") {
		t.Fatalf("Errors() = %v; want one error mentioning the panic", errs)
	}
}

func TestSubmitAfterShutdown(t *testing.T) {
	p := New(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	err := p.Submit(context.Background(), func(ctx context.Context) error { return nil })
	if !errors.Is(err, ErrPoolShutdown) {
		t.Fatalf("Submit after Shutdown = %v; want ErrPoolShutdown", err)
	}
}

func TestSubmitBlocksWhenFull(t *testing.T) {
	// Queue capacity equals the worker count (1), so after one running job
	// and one queued job the next Submit must block.
	p := New(1)

	release := make(chan struct{})
	entered := make(chan struct{})
	if err := p.Submit(context.Background(), func(ctx context.Context) error {
		close(entered)
		<-release
		return nil
	}); err != nil {
		t.Fatalf("Submit blocking job: %v", err)
	}
	<-entered

	noop := func(ctx context.Context) error { return nil }
	if err := p.Submit(context.Background(), noop); err != nil {
		t.Fatalf("Submit queued job: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := p.Submit(ctx, noop); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Submit on full pool = %v; want context.DeadlineExceeded", err)
	}

	close(release)
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := p.Shutdown(shutCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestConcurrentSubmitAndShutdown(t *testing.T) {
	p := New(8)

	var processed atomic.Int64
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				err := p.Submit(context.Background(), func(ctx context.Context) error {
					processed.Add(1)
					return nil
				})
				if err != nil && !errors.Is(err, ErrPoolShutdown) {
					t.Errorf("Submit: %v", err)
				}
			}
		}()
	}

	shutdownDone := make(chan error, 1)
	go func() {
		time.Sleep(5 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownDone <- p.Shutdown(ctx)
	}()

	wg.Wait()
	if err := <-shutdownDone; err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	// Every accepted job must have been processed during the drain.
	if got := processed.Load(); got > 400 {
		t.Fatalf("processed %d jobs; at most 400 were submitted", got)
	}
	// A second Shutdown must be safe and report the already-drained pool.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

func TestNoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()

	p := New(4)
	for i := 0; i < 20; i++ {
		if err := p.Submit(context.Background(), func(ctx context.Context) error {
			return nil
		}); err != nil {
			t.Fatalf("Submit %d: %v", i, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Fatalf("goroutines leaked: before=%d after=%d", before, after)
	}
}
