package solutions

// IsPalindrome checks if a string is a palindrome.
func IsPalindrome(s string) bool {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		if runes[i] != runes[n-1-i] {
			return false
		}
	}
	return true
}
