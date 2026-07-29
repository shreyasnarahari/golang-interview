package practice

// Stack represents a generic thread-safe LIFO stack.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

// Deduplicate removes duplicate elements from a slice of comparable items.
func Deduplicate[K comparable](slice []K) []K {
	seen := make(map[K]bool, len(slice))
	var res []K
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			res = append(res, item)
		}
	}
	return res
}
