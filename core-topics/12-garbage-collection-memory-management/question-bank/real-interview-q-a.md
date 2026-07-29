# Topic 12 Real Interview Q&A

### Q: What is GOMEMLIMIT in Go 1.19+ and why is it important for Cloud / Kubernetes?
**Answer**: `GOMEMLIMIT` sets a soft memory limit for the Go runtime heap. It prevents container OOM kills by aggressively triggering GC cycles when memory approaches container limits, avoiding the need for manual memory ballast hacks.

### Q: Does Go use reference counting or mark-sweep GC?
**Answer**: Go strictly uses a concurrent **Tri-Color Mark-Sweep** garbage collector.
