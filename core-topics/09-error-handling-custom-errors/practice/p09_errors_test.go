package practice

import (
	"errors"
	"testing"
)

func TestErrorsIs(t *testing.T) {
	err := PerformQuery(0)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected errors.Is to match ErrNotFound, got %v", err)
	}
}

func TestErrorsAs(t *testing.T) {
	err := PerformQuery(-1)
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected errors.As to extract ValidationError, got %v", err)
	}

	if valErr.Field != "id" {
		t.Fatalf("Expected field 'id', got %s", valErr.Field)
	}
}
