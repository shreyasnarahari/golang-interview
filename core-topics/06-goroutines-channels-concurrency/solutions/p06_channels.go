package solutions

// SumWithGoroutines computes total sum of integer slice using two concurrent goroutines.
func SumWithGoroutines(nums []int) int {
	ch := make(chan int, 2)
	mid := len(nums) / 2

	go func(slice []int) {
		sum := 0
		for _, v := range slice {
			sum += v
		}
		ch <- sum
	}(nums[:mid])

	go func(slice []int) {
		sum := 0
		for _, v := range slice {
			sum += v
		}
		ch <- sum
	}(nums[mid:])

	part1 := <-ch
	part2 := <-ch
	return part1 + part2
}

// NonBlockingReceive attempts to read from a channel without blocking,
// using select with a default branch. Returns ok=false immediately if no
// value is ready.
func NonBlockingReceive(ch <-chan int) (int, bool) {
	select {
	case val, ok := <-ch:
		return val, ok
	default:
		return 0, false
	}
}
