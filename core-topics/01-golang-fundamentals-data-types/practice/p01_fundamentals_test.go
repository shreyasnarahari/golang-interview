package practice

import (
	"testing"
)

func TestCountRunes(t *testing.T) {
	str := "Go 🐹"
	expectedRuneCount := 4
	if count := CountRunes(str); count != expectedRuneCount {
		t.Errorf("Expected %d runes, got %d", expectedRuneCount, count)
	}
}

func TestBuildStringConcatenation(t *testing.T) {
	parts := []string{"Hello", ", ", "Go", "!"}
	expected := "Hello, Go!"
	if res := BuildStringConcatenation(parts); res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}
