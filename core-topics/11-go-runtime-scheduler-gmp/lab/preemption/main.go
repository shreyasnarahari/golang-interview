// Command preemption demonstrates Go 1.14+ asynchronous preemption: with a
// single P, a tight loop containing no function calls would hog the only P
// forever in Go < 1.14. Since Go 1.14 the runtime preempts such loops with
// SIGURG signals, so main still resumes.
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)

	go func() {
		for {
		}
	}()

	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("main resumed after %v despite a tight loop hogging the only P\n", time.Since(start))
}
