package solutions

import "fmt"

// SafeExecute runs a function that might panic and safely intercepts it using recover.
func SafeExecute(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered panic: %v", r)
		}
	}()
	fn()
	return nil
}

// IncrementWithDefer demonstrates defer modifying a named return parameter.
func IncrementWithDefer(val int) (result int) {
	defer func() {
		result += 10
	}()
	return val * 2
}
