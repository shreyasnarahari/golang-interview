package exercises

import (
	"context"
	"time"
)

// FetchDataWithTimeout simulates a network fetch that respects context cancellation/timeout.
func FetchDataWithTimeout(ctx context.Context, latency time.Duration) (string, error) {
	panic("TODO: implement FetchDataWithTimeout")
}

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserContext injects a user ID into a derived context.
func WithUserContext(ctx context.Context, userID string) context.Context {
	panic("TODO: implement WithUserContext")
}

// GetUserFromContext retrieves a user ID from context.
func GetUserFromContext(ctx context.Context) (string, error) {
	panic("TODO: implement GetUserFromContext")
}
