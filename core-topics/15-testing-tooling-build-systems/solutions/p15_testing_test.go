package solutions

import (
	"testing"
)

func TestIsPalindromeTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple palindrome", "racecar", true},
		{"unicode palindrome", "level", true},
		{"non-palindrome", "golang", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := IsPalindrome(tt.input); res != tt.expected {
				t.Errorf("IsPalindrome(%q) = %v; want %v", tt.input, res, tt.expected)
			}
		})
	}
}
