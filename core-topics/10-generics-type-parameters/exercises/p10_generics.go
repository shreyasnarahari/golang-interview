package exercises

// Stack represents a generic LIFO stack.
// It is NOT safe for concurrent use; guard it with a sync.Mutex if shared across goroutines.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	panic("TODO: implement Push")
}

func (s *Stack[T]) Pop() (T, bool) {
	panic("TODO: implement Pop")
}

// Deduplicate removes duplicate elements from a slice of comparable items.
func Deduplicate[K comparable](slice []K) []K {
	panic("TODO: implement Deduplicate")
}
