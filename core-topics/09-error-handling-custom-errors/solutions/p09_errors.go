package solutions

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("record not found")

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s': %s", e.Field, e.Msg)
}

// PerformQuery simulates a DB query returning wrapped errors.
func PerformQuery(id int) error {
	if id < 0 {
		return fmt.Errorf("query error: %w", &ValidationError{Field: "id", Msg: "must be non-negative"})
	}
	if id == 0 {
		return fmt.Errorf("db failure: %w", ErrNotFound)
	}
	return nil
}
