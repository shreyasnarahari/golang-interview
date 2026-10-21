package solutions

import (
	"context"
	"errors"
	"time"
)

// FetchDataWithTimeout simulates a network fetch that respects context cancellation/timeout.
func FetchDataWithTimeout(ctx context.Context, latency time.Duration) (string, error) {
	select {
	case <-time.After(latency):
		return "success data", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserContext injects a user ID into a derived context.
func WithUserContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserFromContext retrieves a user ID from context.
func GetUserFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(userIDKey)
	if str, ok := val.(string); ok {
		return str, nil
	}
	return "", errors.New("user ID not found in context")
}
