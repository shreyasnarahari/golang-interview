package solutions

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

func TestNilError(t *testing.T) {
	err := PerformQuery(1)
	if err != nil {
		t.Fatalf("Expected nil error for valid id, got %v", err)
	}

	if errors.Is(err, ErrNotFound) {
		t.Fatal("Expected errors.Is(nil, ErrNotFound) to be false")
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		t.Fatal("Expected errors.As(nil, &valErr) to be false")
	}
}
