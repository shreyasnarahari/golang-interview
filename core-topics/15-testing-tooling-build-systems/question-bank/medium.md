# Topic 15 Medium Questions

### Q: What is `TestMain(m *testing.M)` used for?
**Frequency**: often
**Why interviewers ask**: It reveals whether the candidate can structure real test suites with shared setup (databases, servers) instead of per-test hacks.
**Answer**: `TestMain` runs once per package, wrapping the entire test run — it is for global setup/teardown such as starting a test database container or initializing shared state:
```go
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
```
Key subtleties: you must call `m.Run()` yourself and pass its exit code to `os.Exit`; because `os.Exit` skips deferred functions, `teardown()` must be called explicitly *before* exiting. `TestMain` runs even when `-run` filters out every test. There is at most one `TestMain` per package.
**Follow-ups**: Why don't `defer` statements in `TestMain` run? What happens to `TestMain` when you pass `-run TestThatDoesNotExist`?

### Q: What are subtests (`t.Run`), and what do they buy you over plain loops?
**Frequency**: often
**Why interviewers ask**: Subtests are where table-driven tests become powerful; the question probes practical knowledge of filtering, isolation, and parallelism.
**Answer**: `t.Run(name, func(t *testing.T) {...})` creates a named subtest with its own `*testing.T`. Benefits: (1) individual cases can be selected with `go test -run 'TestX/case_name'` (slash-separated regex path); (2) a failure in one subtest doesn't stop siblings; (3) subtests can call `t.Parallel()` to run concurrently; (4) verbose output and coverage attribute failures to the specific case. A `t.Parallel()` subtest pauses and only resumes once its parent test function has returned, so the parent must not rely on subtest results. Since Go 1.22 each loop iteration gets fresh variables, so the classic `tc := tc` capture workaround is no longer needed.
**Follow-ups**: How does `t.Parallel()` interact with subtests and the `-parallel` flag? What is `t.Cleanup` and in what order do cleanups run? (LIFO)

### Q: How do you write and run a fuzz test in Go?
**Frequency**: sometimes
**Why interviewers ask**: Native fuzzing (Go 1.18+) is a differentiator — it shows the candidate knows modern Go's built-in security/reliability tooling.
**Answer**: A fuzz target uses `testing.F`, seeds a corpus with `f.Add(...)`, and defines a property (not a fixed expected value) in `f.Fuzz`:
```go
func FuzzParse(f *testing.F) {
	f.Add("2024-01-02") // seed corpus
	f.Fuzz(func(t *testing.T, s string) {
		v1, err1 := Parse(s)
		if err1 != nil {
			return // invalid input is fine; only crashes/panics matter here
		}
		out := v1.String()
		if _, err2 := Parse(out); err2 != nil { // round-trip property
			t.Fatalf("round trip failed: %q -> %q", s, out)
		}
	})
}
```
Run it with `go test -fuzz=FuzzParse -fuzztime=30s ./...`. The engine is coverage-guided: it mutates inputs to explore new code paths, caches the generated corpus in `$GOCACHE/fuzz`, and writes any failing input to `testdata/fuzz/FuzzParse/` so it becomes a permanent regression test. Without `-fuzz`, the seed corpus runs as ordinary unit tests. `-fuzzminimizetime` controls how long a failing input is shrunk to a minimal reproducer.
**Follow-ups**: How does Go's coverage-guided fuzzer differ from property-based testing libraries? Where do crash reproducers get stored, and why is that design useful in CI?

### Q: What is the difference between `gofmt`, `go vet`, and `golangci-lint`?
**Frequency**: sometimes
**Why interviewers ask**: Separates candidates who know the standard toolchain from those who cargo-cult a linter config without understanding the layers.
**Answer**: They operate at three different layers. `gofmt` only rewrites source layout — it never judges semantics. `go vet` is the stdlib's built-in static analysis: it catches suspicious constructs that still compile, e.g. wrong `Printf` verbs, copying a `sync.Mutex` by value (`copylocks`), unreachable build constraints, or invalid struct tags; it runs automatically as part of `go test`. `golangci-lint` is a third-party *aggregator* that runs dozens of linters in one pass (govet, staticcheck, errcheck, revive, gosec…) with a shared `.golangci.yml` config, caching, and unified output. Typical pipeline: `gofmt -l` (format) → `go vet ./...` (stdlib checks) → `golangci-lint run` (everything else), all gated in CI. Since Go 1.24, tool dependencies like linters can also be pinned in `go.mod` via the `tool` directive.
**Follow-ups**: Which vet check would catch passing a `sync.Mutex` by value? How do you pin a linter version reproducibly in a Go 1.24+ project?
