# Topic 14 Medium Questions

### Q: Why should `b.ResetTimer()` be called in Go benchmarks?
**Answer**: To discard setup time (e.g. reading config files, allocating mock datasets) so the benchmark metrics reflect only the target function execution runtime.
