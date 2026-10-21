# Topic 15: Testing, Tooling, & Build Systems

## 📌 Overview
Go features a battery-included toolchain supporting unit testing, table-driven subtests, native fuzzing, race detection, code coverage, structured logging (`log/slog`), static analysis, and module/workspace management.

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

### 4. Native Fuzzing (Go 1.18+)
Fuzzing feeds auto-generated pseudo-random inputs to your code to surface panics and broken invariants. Fuzz tests live in `_test.go` files alongside unit tests:

```go
func FuzzIsPalindrome(f *testing.F) {
    f.Add("racecar") // seed corpus: known-interesting inputs
    f.Add("golang")
    f.Fuzz(func(t *testing.T, s string) {
        // assert properties that must hold for ANY input
        if IsPalindrome(s) != IsPalindrome(ReverseRunes(s)) {
            t.Errorf("palindrome check not symmetric for %q", s)
        }
    })
}
```

- **Signature**: `func FuzzXxx(f *testing.F)`. The body passed to `f.Fuzz` must take `*testing.T` first, then the fuzzed parameters. Supported parameter types: `string`, `[]byte`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `rune`.
- **Seed corpus**: inputs registered via `f.Add(...)` always run, even during a plain `go test` (no `-fuzz` flag).
- **Run fuzzing**: `go test -fuzz=FuzzIsPalindrome -fuzztime=30s` (`-fuzztime` bounds the run; without it fuzzing continues until interrupted).
- **Crashers**: failing inputs are written to `testdata/fuzz/FuzzIsPalindrome/<hash>` and automatically replayed by every future `go test` run — a built-in regression suite.

See `solutions/p15_fuzz_test.go` for a working `FuzzIsPalindrome` example.

### 5. Structured Logging with `log/slog` (Go 1.21+)
`slog` is the standard library's structured, leveled logger. Attributes are passed as alternating key-value pairs:

```go
slog.Info("server started", "port", 8080, "env", "prod")
// default text handler: 2026/10/21 09:15:00 INFO server started port=8080 env=prod
```

Swap the global logger for JSON output (typical in production) with a handler and `slog.SetDefault`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
slog.SetDefault(logger)
slog.Info("request served", "method", "GET", "status", 200)
// {"time":"...","level":"INFO","msg":"request served","method":"GET","status":200}
```

### 6. Multi-Module Workspaces (`go.work`, Go 1.18+)
A workspace ties multiple modules together so local edits in one are picked up by another without editing `go.mod` `replace` directives:

```bash
go work init ./service   # create go.work, add ./service
go work use ./lib        # add more modules to the workspace
```

- **Use a workspace** when actively developing several interdependent modules locally (e.g. a service and a library it consumes) and you want cross-module changes to compile and test together.
- **Use a single module** for self-contained projects — simpler builds, versioning, and CI. `go.work` is a local development convenience; it is commonly git-ignored, and released artifacts should still be built per-module.

---

## ⚠️ Common Gotchas
1. **Goroutine Loop Variable in `t.Run`**: Before Go 1.22, running `t.Parallel()` inside `t.Run` without passing `tt` by parameter caused race conditions on loop iteration variables!
