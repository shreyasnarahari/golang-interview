package practice

import (
	"strings"
	"unicode/utf8"
)

// CountRunes returns the number of Unicode code points (runes) in a string.
func CountRunes(s string) int {
	return utf8.RuneCountInString(s)
}

// BuildStringConcatenation efficiently concatenates a slice of strings using strings.Builder.
func BuildStringConcatenation(parts []string) string {
	var builder strings.Builder
	for _, p := range parts {
		builder.WriteString(p)
	}
	return builder.String()
}
