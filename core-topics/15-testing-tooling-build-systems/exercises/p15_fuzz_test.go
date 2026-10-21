package exercises

import (
	"testing"
)

func FuzzIsPalindrome(f *testing.F) {
	seeds := []string{"racecar", "level", "golang", "", "ab", "こんにちは"}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		rev := string(runes)

		if IsPalindrome(s) != IsPalindrome(rev) {
			t.Errorf("IsPalindrome(%q) != IsPalindrome(reverse %q)", s, rev)
		}
		if !IsPalindrome(s + rev) {
			t.Errorf("IsPalindrome(%q + its reverse) = false; want true", s)
		}
	})
}
