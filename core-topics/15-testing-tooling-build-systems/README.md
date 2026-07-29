# Topic 15: Testing, Tooling, & Build Systems

## 📌 Overview
Go features a battery-included toolchain supporting unit testing, table-driven subtests, race detection, code coverage, static analysis, and module management.

---

## 🎯 Key Concepts

### 1. Table-Driven Tests & Subtests (`t.Run`)
Go's idiomatic testing pattern uses struct slices for test cases and `t.Run` for nested subtests:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -1, -2},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if res := Add(tt.a, tt.b); res != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, res, tt.expected)
            }
        })
    }
}
```

### 2. Race Detector (`-race`)
Go provides built-in data race detection instrumented via ThreadSanitizer (TSan):
```bash
go test -race ./...
```
Detects unsafe concurrent read/writes to un-synchronized variables.

### 3. Tooling Ecosystem
- **`gofmt` / `goimports`**: Opinionated code formatting.
- **`go vet`**: Examines Go source code for suspicious constructs (e.g. unreachable code, printf format specifier mismatches).
- **`golangci-lint`**: Multi-linter aggregator used in modern Go CI/CD pipelines.

---

## ⚠️ Common Gotchas
1. **Goroutine Loop Variable in `t.Run`**: Before Go 1.22, running `t.Parallel()` inside `t.Run` without passing `tt` by parameter caused race conditions on loop iteration variables!
