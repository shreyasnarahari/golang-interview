# Topic 12 Hard Questions

### Q: Explain the hybrid write barrier algorithm in Go 1.8+.
**Answer**: Hybrid write barrier combines Dijkstra (shading new/mutated pointers Grey) and Yuasa (shading overwritten pointers Grey) barriers. When GC is active, any pointer modified or created on heap/stack is colored Grey immediately. This eliminates stack re-scanning during Mark Termination, achieving sub-millisecond STW latency.
