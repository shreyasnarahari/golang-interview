// Command schedtrace spawns far more busy goroutines than there are Ps, so
// GODEBUG=schedtrace output shows runnable goroutines piling up in the local
// and global run queues.
//
// Run with:
//
//	GODEBUG=schedtrace=1000 go run ./lab/schedtrace
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Printf("NumCPU=%d GOMAXPROCS=%d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))

	stop := time.After(4 * time.Second)
	var wg sync.WaitGroup
	for range 4 * runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			for {
				select {
				case <-stop:
					fmt.Println("checksum:", sum)
					return
				default:
					sum++
				}
			}
		}()
	}
	wg.Wait()
}
