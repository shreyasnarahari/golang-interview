# Topic 03 Hard Questions

### Q: Explain open-coded defers introduced in Go 1.14.
**Answer**: For functions with static, non-loop defer calls (or small loops), the Go compiler eliminates `_defer` struct allocations by emitting direct code jumps and a bitmask tracking active defers. This reduced defer overhead from ~35ns to ~1ns.
