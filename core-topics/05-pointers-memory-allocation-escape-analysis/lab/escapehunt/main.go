// Command escapehunt is an escape-analysis scavenger hunt.
//
// For each scenario, predict whether the variable stays on the stack or
// escapes to the heap, then check your predictions:
//
//	go build -gcflags='-m' ./lab/escapehunt/
package main

import "fmt"

// Scenario 1: value returned by copy.
func sum(a, b int) int {
	total := a + b
	return total
}

// Scenario 2: pointer to a local is returned.
func newCounter() *int {
	count := 0
	return &count
}

// Scenario 3: value passed to an interface{} parameter.
func report() {
	value := 42
	fmt.Println(value)
}

// Scenario 4: slice with a compile-time constant length.
func fixedWindow() int {
	buf := make([]int, 8)
	buf[0] = 1
	return buf[0]
}

// Scenario 5: slice with a runtime-determined length, returned to the caller.
func dynamicWindow(n int) []int {
	return make([]int, n)
}

// Scenario 6: slice with a runtime-determined length, kept local.
func localWindow(n int) int {
	buf := make([]int, n)
	buf[0] = 1
	return buf[0]
}

// Scenario 7: local array too large for the stack frame.
func bigBuffer() byte {
	var buf [1 << 20]byte // 1 MiB
	buf[0] = 7
	return buf[0]
}

// Scenario 8: variable captured by reference in a returned closure.
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// Scenario 9: pointer stored in a package-level variable.
var sink *int

func leak() {
	x := 99
	sink = &x
}

// Scenario 10: pointer sent through a channel.
func send() int {
	ch := make(chan *int, 1)
	v := 5
	ch <- &v
	return *<-ch
}

func main() {
	fmt.Println(sum(1, 2))
	fmt.Println(*newCounter())
	report()
	fmt.Println(fixedWindow())
	fmt.Println(len(dynamicWindow(16)))
	fmt.Println(localWindow(16))
	fmt.Println(bigBuffer())
	next := counter()
	fmt.Println(next(), next())
	leak()
	fmt.Println(*sink)
	fmt.Println(send())
}
