package solutions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFetchDataWithTimeout_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := FetchDataWithTimeout(ctx, 10*time.Millisecond)
	if err != nil || data != "success data" {
		t.Fatalf("Expected success, got data=%s, err=%v", data, err)
	}
}

func TestFetchDataWithTimeout_Exceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := FetchDataWithTimeout(ctx, 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestFetchDataWithTimeout_ContextErrAfterTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("Expected ctx.Err() to be DeadlineExceeded, got %v", ctx.Err())
	}
}

func TestContextValue(t *testing.T) {
	ctx := WithUserContext(context.Background(), "user_99")
	user, err := GetUserFromContext(ctx)
	if err != nil || user != "user_99" {
		t.Fatalf("Expected user_99, got user=%s, err=%v", user, err)
	}
}
