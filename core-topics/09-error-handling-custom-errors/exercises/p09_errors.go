package exercises

import (
	"errors"
)

var ErrNotFound = errors.New("record not found")

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	panic("TODO: implement Error")
}

// PerformQuery simulates a DB query returning wrapped errors.
func PerformQuery(id int) error {
	panic("TODO: implement PerformQuery")
}
