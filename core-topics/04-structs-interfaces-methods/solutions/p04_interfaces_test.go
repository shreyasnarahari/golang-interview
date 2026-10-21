package solutions

import (
	"testing"
)

func TestPointerReceiverInterface(t *testing.T) {
	// Compile-time check
	var _ Speaker = (*Dog)(nil)

	dog := &Dog{Name: "Rex"}
	if msg := dog.Speak(); msg != "Woof! I am Rex" {
		t.Fatalf("Unexpected output: %s", msg)
	}
}

func TestNilInterfaceTrap(t *testing.T) {
	var d *Dog = nil
	var s Speaker = d // Assign nil *Dog pointer to Speaker interface

	// Interface is NOT nil because it holds type *Dog
	if s == nil {
		t.Fatalf("Expected interface to be non-nil due to typed pointer assignment")
	}
}

func TestTypeAssertionCommaOk(t *testing.T) {
	var s Speaker = &Dog{Name: "Rex"}

	// Successful assertion returns the concrete value and ok=true
	d, ok := s.(*Dog)
	if !ok || d.Name != "Rex" {
		t.Fatalf("Expected comma-ok assertion to *Dog to succeed, ok=%v", ok)
	}

	// Failed assertion with comma-ok returns zero value instead of panicking
	type Barker interface{ Bark() string }
	if b, ok := s.(Barker); ok || b != nil {
		t.Fatalf("Expected comma-ok assertion to Barker to fail with zero value")
	}
}
