# Topic 14 Real Interview Q&A

### Q: How do you enable live HTTP pprof profiling endpoints in a Go web service?
**Answer**: Import `_ "net/http/pprof"` for side-effects, which automatically registers debug profiling handlers under `/debug/pprof/`.

### Q: What flag displays memory allocation bytes in `go test -bench`?
**Answer**: `-benchmem`.
