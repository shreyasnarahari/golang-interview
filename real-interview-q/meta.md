# Meta — Go Real Interview Questions & Deep Dives

> **Target Roles**: Infrastructure Software Engineer, Developer Tooling Engineer, Network Operations Developer.

---

## 🏢 Why Meta Asks Go
Meta uses Go for internal developer tools, infrastructure automation, network probing agents, and container containerization pipelines. Meta interviews test **memory alignment & struct padding**, **cgo performance bottlenecks**, and **goroutine channel synchronization**.

---

## 📌 Question 1: Memory Alignment, Struct Padding & Offset Optimization

### Question
Given the two struct definitions below, calculate their memory sizes on a 64-bit architecture (`amd64`). Reorder fields to minimize struct size and explain field alignment rules.

```go
type UnoptimizedStruct struct {
	A bool    // 1 byte
	B int64   // 8 bytes
	C bool    // 1 byte
	D float64 // 8 bytes
}

type OptimizedStruct struct {
	B int64   // 8 bytes
	D float64 // 8 bytes
	A bool    // 1 byte
	C bool    // 1 byte
	// 6 bytes padding appended by compiler to align struct size to 8 bytes
}
```

### Detailed Solution & Explanation

#### Memory Sizes
- `UnoptimizedStruct`: **32 bytes**
  - `A`: 1 byte + 7 bytes padding (to align `B` at 8-byte boundary).
  - `B`: 8 bytes.
  - `C`: 1 byte + 7 bytes padding (to align `D` at 8-byte boundary).
  - `D`: 8 bytes.
  - Total = $1+7+8+1+7+8 = 32\text{ bytes}$.
- `OptimizedStruct`: **24 bytes** (Saved 25% memory overhead!)
  - `B`: 8 bytes.
  - `D`: 8 bytes.
  - `A`: 1 byte.
  - `C`: 1 byte.
  - 6 bytes tail padding (to make total struct size a multiple of maximum field alignment, 8 bytes).
  - Total = $8+8+1+1+6 = 24\text{ bytes}$.

#### Verification Code
```go
package main

import (
	"fmt"
	"unsafe"
)

type UnoptimizedStruct struct {
	A bool
	B int64
	C bool
	D float64
}

type OptimizedStruct struct {
	B int64
	D float64
	A bool
	C bool
}

func main() {
	fmt.Println("Unoptimized size:", unsafe.Sizeof(UnoptimizedStruct{})) // 32
	fmt.Println("Optimized size:", unsafe.Sizeof(OptimizedStruct{}))     // 24
}
```
