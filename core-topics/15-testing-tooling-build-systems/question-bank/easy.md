# Topic 15 Easy Questions

### Q: What signature must a test function have in Go?
**Frequency**: often
**Why interviewers ask**: It is the most basic litmus test that a candidate has actually written Go tests.
**Answer**: A test lives in a `*_test.go` file and has the form `func TestXxx(t *testing.T)` — the name must start with `Test` followed by a capitalized (or non-lowercase) suffix, and it takes exactly one `*testing.T` parameter. Benchmarks analogously use `func BenchmarkXxx(b *testing.B)` and fuzz tests `func FuzzXxx(f *testing.F)`. Tests run via `go test`.
**Follow-ups**: What are `Benchmark` and `Example` functions? What is the difference between an internal (`package foo`) and external (`package foo_test`) test file?

### Q: What is a table-driven test and why is it the idiomatic Go testing style?
**Frequency**: often
**Why interviewers ask**: Table-driven tests are the canonical Go idiom; interviewers want to see if you write tests the "Go way" or copy JUnit-style one-test-per-case.
**Answer**: You define a slice of anonymous structs holding inputs and expected outputs, then loop over it running the same assertions — often wrapped in `t.Run(tc.name, ...)` so each case is a named subtest:
```go
tests := []struct {
	name  string
	input string
	want  int
}{
	{"empty", "", 0},
	{"ascii", "hello", 5},
}
for _, tc := range tests {
	t.Run(tc.name, func(t *testing.T) {
		if got := Len(tc.input); got != tc.want {
			t.Errorf("Len(%q) = %d, want %d", tc.input, got, tc.want)
		}
	})
}
```
It scales to many cases without code duplication, makes failures self-describing, and keeps adding a regression case a one-line change.
**Follow-ups**: What was the loop-variable-capture gotcha with `t.Parallel()` in table-driven tests before Go 1.22? How do you run just one case from the table?

### Q: How do you run tests in Go, and what do the common `go test` flags do?
**Frequency**: often
**Why interviewers ask**: Daily-driver knowledge — shows the candidate actually works with the toolchain rather than only reading about it.
**Answer**: `go test ./...` runs all tests in the module (results of passing packages are cached; source changes invalidate the cache). Key flags: `-run TestFoo` (regex filter, use `TestFoo/subname` for subtests), `-v` (verbose), `-count=1` (bypass the result cache), `-race` (race detector), `-cover` / `-coverprofile` (coverage), `-shuffle=on` (randomize test order to surface order dependence), `-short` (with `testing.Short()` to skip slow tests), `-timeout` (default 10m per package), and `-bench` / `-fuzz` for benchmarks and fuzzing.
**Follow-ups**: Why might a test "pass" in CI without actually running? (stale cache — fix with `-count=1`) What does `-shuffle=on` help you catch?

### Q: What does `gofmt` do, and why does Go enforce a single canonical format?
**Frequency**: often
**Why interviewers ask**: Checks whether the candidate understands Go's "no formatting debates" culture and basic tooling hygiene.
**Answer**: `gofmt` rewrites Go source into the one canonical style (tabs for indentation, brace placement, alignment). `gofmt -l .` lists non-conforming files, `gofmt -w .` rewrites in place, `gofmt -d` shows a diff. The Go authors deliberately made formatting mechanical so that code reviews never argue about style, diffs stay minimal, and tools can rewrite code safely. Related tools: `goimports` (also fixes import grouping) and `go vet` (static checks, not formatting). Most teams gate CI on `gofmt -l` returning empty.
**Follow-ups**: What is the difference between `gofmt` and `goimports`? How would you enforce formatting in CI?
