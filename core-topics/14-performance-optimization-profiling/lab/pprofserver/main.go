package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func burnCPU(w http.ResponseWriter, r *http.Request) {
	sum := sha256.Sum256([]byte("seed"))
	for i := 0; i < 2_000_000; i++ {
		sum = sha256.Sum256(sum[:])
	}
	fmt.Fprintf(w, "digest: %x\n", sum[:8])
}

func churnMemory(w http.ResponseWriter, r *http.Request) {
	var kept [][]byte
	for i := 0; i < 200_000; i++ {
		kept = append(kept, []byte(fmt.Sprintf("payload-%d", i)))
	}
	fmt.Fprintf(w, "allocated %d buffers\n", len(kept))
}

func main() {
	http.HandleFunc("/cpu", burnCPU)
	http.HandleFunc("/alloc", churnMemory)
	log.Println("listening on :8080 (pprof at http://localhost:8080/debug/pprof/)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
