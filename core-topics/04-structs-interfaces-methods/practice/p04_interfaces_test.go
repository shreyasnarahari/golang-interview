package practice

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
